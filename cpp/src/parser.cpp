#include "ason/parser.hpp"

namespace ason {

// SchemaField copy constructor
SchemaField::SchemaField(const SchemaField& other)
    : type(other.type), name(other.name) {
    if (other.nested_schema) {
        nested_schema = std::make_unique<Schema>(*other.nested_schema);
    }
}

SchemaField& SchemaField::operator=(const SchemaField& other) {
    if (this != &other) {
        type = other.type;
        name = other.name;
        if (other.nested_schema) {
            nested_schema = std::make_unique<Schema>(*other.nested_schema);
        } else {
            nested_schema.reset();
        }
    }
    return *this;
}

Parser::Parser(std::string_view input) : lexer_(input), current_(TokenType::Eof, 0, 0) {
    auto result = lexer_.next_token();
    if (result.ok()) {
        current_ = std::move(result).value();
    }
}

Result<void> Parser::advance() {
    auto result = lexer_.next_token();
    if (result.is_error()) return result.error();
    current_ = std::move(result).value();
    return Result<void>(std::monostate{});
}

Result<void> Parser::expect(TokenType type) {
    if (current_.type == type) {
        return advance();
    }
    return AsonError::expected_token(
        std::to_string(static_cast<int>(type)),
        std::to_string(static_cast<int>(current_.type)),
        position()
    );
}

Result<Schema> Parser::parse_schema() {
    auto expect_result = expect(TokenType::LBrace);
    if (expect_result.is_error()) return expect_result.error();
    
    std::vector<SchemaField> fields;
    
    // Handle empty schema
    if (check(TokenType::RBrace)) {
        auto adv = advance();
        if (adv.is_error()) return adv.error();
        return Schema(std::move(fields));
    }
    
    // Parse first field
    auto field_result = parse_schema_field();
    if (field_result.is_error()) return field_result.error();
    fields.push_back(std::move(field_result).value());
    
    // Parse remaining fields
    while (check(TokenType::Comma)) {
        auto adv = advance();
        if (adv.is_error()) return adv.error();
        if (check(TokenType::RBrace)) break; // Trailing comma
        
        auto next_field = parse_schema_field();
        if (next_field.is_error()) return next_field.error();
        fields.push_back(std::move(next_field).value());
    }
    
    expect_result = expect(TokenType::RBrace);
    if (expect_result.is_error()) return expect_result.error();
    
    return Schema(std::move(fields));
}

Result<SchemaField> Parser::parse_schema_field() {
    // Expect an identifier
    if (current_.type != TokenType::Ident) {
        return AsonError::invalid_schema(position(), "expected field name");
    }
    
    std::string name = *current_.str();
    auto adv = advance();
    if (adv.is_error()) return adv.error();
    
    // Check what follows: {, [, or nothing
    if (check(TokenType::LBrace)) {
        // Nested object: name{...}
        auto schema_result = parse_schema();
        if (schema_result.is_error()) return schema_result.error();
        return SchemaField(FieldType::NestedObject, std::move(name),
                          std::make_unique<Schema>(std::move(schema_result).value()));
    } else if (check(TokenType::LBracket)) {
        adv = advance();
        if (adv.is_error()) return adv.error();
        
        if (check(TokenType::LBrace)) {
            // Object array: name[{...}]
            auto schema_result = parse_schema();
            if (schema_result.is_error()) return schema_result.error();
            
            auto exp = expect(TokenType::RBracket);
            if (exp.is_error()) return exp.error();
            
            return SchemaField(FieldType::ObjectArray, std::move(name),
                              std::make_unique<Schema>(std::move(schema_result).value()));
        } else if (check(TokenType::RBracket)) {
            // Simple array: name[]
            adv = advance();
            if (adv.is_error()) return adv.error();
            return SchemaField(FieldType::SimpleArray, std::move(name));
        } else {
            return AsonError::invalid_schema(position(), "expected '{' or ']' after '['");
        }
    } else {
        // Simple field
        return SchemaField(FieldType::Simple, std::move(name));
    }
}

Result<Value> Parser::parse_data_with_schema(const Schema& schema) {
    // Could be single object or multiple objects
    auto first_result = parse_object_with_schema(schema);
    if (first_result.is_error()) return first_result.error();
    Value first = std::move(first_result).value();
    
    // Check if there are more objects
    if (check(TokenType::Comma)) {
        Array objects;
        objects.push_back(std::move(first));
        
        while (check(TokenType::Comma)) {
            auto adv = advance();
            if (adv.is_error()) return adv.error();
            if (check(TokenType::Eof)) break;
            
            auto obj_result = parse_object_with_schema(schema);
            if (obj_result.is_error()) return obj_result.error();
            objects.push_back(std::move(obj_result).value());
        }
        return Value(std::move(objects));
    }
    
    return first;
}

Result<Value> Parser::parse_object_with_schema(const Schema& schema) {
    auto exp = expect(TokenType::LParen);
    if (exp.is_error()) return exp.error();

    Object obj;
    size_t field_count = schema.size();

    for (size_t i = 0; i < schema.fields.size(); ++i) {
        if (i > 0) {
            exp = expect(TokenType::Comma);
            if (exp.is_error()) return exp.error();
        }

        auto val_result = parse_field_value(schema.fields[i]);
        if (val_result.is_error()) return val_result.error();
        obj[schema.fields[i].name] = std::move(val_result).value();
    }

    // Check for extra values
    if (check(TokenType::Comma)) {
        auto adv = advance();
        if (adv.is_error()) return adv.error();
        if (!check(TokenType::RParen)) {
            return AsonError::field_count_mismatch(field_count, field_count + 1, position());
        }
    }

    exp = expect(TokenType::RParen);
    if (exp.is_error()) return exp.error();

    return Value(std::move(obj));
}

Result<Value> Parser::parse_field_value(const SchemaField& field) {
    switch (field.type) {
        case FieldType::Simple:
            return parse_simple_value();
        case FieldType::NestedObject:
            return parse_object_with_schema(*field.nested_schema);
        case FieldType::SimpleArray:
            return parse_simple_array();
        case FieldType::ObjectArray:
            return parse_object_array(*field.nested_schema);
    }
    return Value(Null{});
}

Result<Value> Parser::parse_simple_value() {
    switch (current_.type) {
        case TokenType::Str: {
            std::string s = *current_.str();
            auto adv = advance();
            if (adv.is_error()) return adv.error();
            return Value(std::move(s));
        }
        case TokenType::Integer: {
            int64_t n = *current_.integer();
            auto adv = advance();
            if (adv.is_error()) return adv.error();
            return Value(n);
        }
        case TokenType::Float: {
            double n = *current_.floating();
            auto adv = advance();
            if (adv.is_error()) return adv.error();
            return Value(n);
        }
        case TokenType::True: {
            auto adv = advance();
            if (adv.is_error()) return adv.error();
            return Value(true);
        }
        case TokenType::False: {
            auto adv = advance();
            if (adv.is_error()) return adv.error();
            return Value(false);
        }
        case TokenType::Comma:
        case TokenType::RParen:
            // Empty value = null
            return Value(Null{});
        case TokenType::LBracket:
            // Inline array
            return parse_array();
        default:
            return AsonError::unexpected_char('?', position());
    }
}

Result<Value> Parser::parse_simple_array() {
    auto exp = expect(TokenType::LBracket);
    if (exp.is_error()) return exp.error();

    Array values;

    if (check(TokenType::RBracket)) {
        auto adv = advance();
        if (adv.is_error()) return adv.error();
        return Value(std::move(values));
    }

    auto val_result = parse_simple_value();
    if (val_result.is_error()) return val_result.error();
    values.push_back(std::move(val_result).value());

    while (check(TokenType::Comma)) {
        auto adv = advance();
        if (adv.is_error()) return adv.error();
        if (check(TokenType::RBracket)) break;

        val_result = parse_simple_value();
        if (val_result.is_error()) return val_result.error();
        values.push_back(std::move(val_result).value());
    }

    exp = expect(TokenType::RBracket);
    if (exp.is_error()) return exp.error();

    return Value(std::move(values));
}

Result<Value> Parser::parse_object_array(const Schema& schema) {
    auto exp = expect(TokenType::LBracket);
    if (exp.is_error()) return exp.error();

    Array objects;

    if (check(TokenType::RBracket)) {
        auto adv = advance();
        if (adv.is_error()) return adv.error();
        return Value(std::move(objects));
    }

    auto obj_result = parse_object_with_schema(schema);
    if (obj_result.is_error()) return obj_result.error();
    objects.push_back(std::move(obj_result).value());

    while (check(TokenType::Comma)) {
        auto adv = advance();
        if (adv.is_error()) return adv.error();
        if (check(TokenType::RBracket)) break;

        obj_result = parse_object_with_schema(schema);
        if (obj_result.is_error()) return obj_result.error();
        objects.push_back(std::move(obj_result).value());
    }

    exp = expect(TokenType::RBracket);
    if (exp.is_error()) return exp.error();

    return Value(std::move(objects));
}

Result<Value> Parser::parse_array() {
    auto exp = expect(TokenType::LBracket);
    if (exp.is_error()) return exp.error();

    Array values;

    if (check(TokenType::RBracket)) {
        auto adv = advance();
        if (adv.is_error()) return adv.error();
        return Value(std::move(values));
    }

    auto val_result = parse_value();
    if (val_result.is_error()) return val_result.error();
    values.push_back(std::move(val_result).value());

    while (check(TokenType::Comma)) {
        auto adv = advance();
        if (adv.is_error()) return adv.error();
        if (check(TokenType::RBracket)) break;

        val_result = parse_value();
        if (val_result.is_error()) return val_result.error();
        values.push_back(std::move(val_result).value());
    }

    exp = expect(TokenType::RBracket);
    if (exp.is_error()) return exp.error();

    return Value(std::move(values));
}

Result<Value> Parser::parse_value() {
    switch (current_.type) {
        case TokenType::Str: {
            std::string s = *current_.str();
            auto adv = advance();
            if (adv.is_error()) return adv.error();
            return Value(std::move(s));
        }
        case TokenType::Integer: {
            int64_t n = *current_.integer();
            auto adv = advance();
            if (adv.is_error()) return adv.error();
            return Value(n);
        }
        case TokenType::Float: {
            double n = *current_.floating();
            auto adv = advance();
            if (adv.is_error()) return adv.error();
            return Value(n);
        }
        case TokenType::True: {
            auto adv = advance();
            if (adv.is_error()) return adv.error();
            return Value(true);
        }
        case TokenType::False: {
            auto adv = advance();
            if (adv.is_error()) return adv.error();
            return Value(false);
        }
        case TokenType::LBracket:
            return parse_array();
        case TokenType::LParen: {
            // Anonymous tuple, parse as array
            auto adv = advance();
            if (adv.is_error()) return adv.error();

            Array values;
            if (!check(TokenType::RParen)) {
                auto val_result = parse_value();
                if (val_result.is_error()) return val_result.error();
                values.push_back(std::move(val_result).value());

                while (check(TokenType::Comma)) {
                    adv = advance();
                    if (adv.is_error()) return adv.error();
                    if (check(TokenType::RParen)) break;

                    val_result = parse_value();
                    if (val_result.is_error()) return val_result.error();
                    values.push_back(std::move(val_result).value());
                }
            }
            auto exp = expect(TokenType::RParen);
            if (exp.is_error()) return exp.error();
            return Value(std::move(values));
        }
        default:
            return Value(Null{});
    }
}

Result<Value> Parser::parse() {
    switch (current_.type) {
        case TokenType::LBrace: {
            // Schema expression: {fields}:data
            auto schema_result = parse_schema();
            if (schema_result.is_error()) return schema_result.error();

            auto exp = expect(TokenType::Colon);
            if (exp.is_error()) return exp.error();

            return parse_data_with_schema(schema_result.value());
        }
        case TokenType::LBracket:
            // Plain array
            return parse_array();
        case TokenType::Eof:
            return Value(Null{});
        default:
            // Single value
            return parse_value();
    }
}

Result<Value> parse(std::string_view input) {
    Parser parser(input);
    return parser.parse();
}

} // namespace ason

