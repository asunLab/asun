#pragma once

#include "error.hpp"
#include "value.hpp"
#include "lexer.hpp"
#include <string>
#include <string_view>
#include <memory>
#include <vector>

namespace ason {

/// Schema field type.
enum class FieldType {
    Simple,         // name
    NestedObject,   // addr{city,zip}
    SimpleArray,    // scores[]
    ObjectArray     // users[{id,name}]
};

/// Forward declaration
struct Schema;

/// A field in an ASON schema.
struct SchemaField {
    FieldType type;
    std::string name;
    std::unique_ptr<Schema> nested_schema;
    
    SchemaField(FieldType t, std::string n, std::unique_ptr<Schema> s = nullptr)
        : type(t), name(std::move(n)), nested_schema(std::move(s)) {}
    
    // Copy constructor
    SchemaField(const SchemaField& other);
    SchemaField& operator=(const SchemaField& other);
    
    // Move constructor/assignment
    SchemaField(SchemaField&&) = default;
    SchemaField& operator=(SchemaField&&) = default;
};

/// ASON schema definition.
struct Schema {
    std::vector<SchemaField> fields;
    
    Schema() = default;
    explicit Schema(std::vector<SchemaField> f) : fields(std::move(f)) {}
    
    size_t size() const { return fields.size(); }
    bool empty() const { return fields.empty(); }
};

/// Parser for ASON input.
class Parser {
public:
    explicit Parser(std::string_view input);
    
    /// Parse the entire ASON input.
    Result<Value> parse();

private:
    Lexer lexer_;
    Token current_;
    
    // Token operations
    TokenType current_type() const { return current_.type; }
    size_t position() const { return current_.start; }
    Result<void> advance();
    Result<void> expect(TokenType type);
    bool check(TokenType type) const { return current_.type == type; }
    
    // Schema parsing
    Result<Schema> parse_schema();
    Result<SchemaField> parse_schema_field();
    
    // Data parsing
    Result<Value> parse_data_with_schema(const Schema& schema);
    Result<Value> parse_object_with_schema(const Schema& schema);
    Result<Value> parse_field_value(const SchemaField& field);
    Result<Value> parse_simple_value();
    Result<Value> parse_simple_array();
    Result<Value> parse_object_array(const Schema& schema);
    Result<Value> parse_array();
    Result<Value> parse_value();
};

/// Parse an ASON string into a Value.
Result<Value> parse(std::string_view input);

} // namespace ason

