#include "ason/lexer.hpp"
#include <cctype>
#include <charconv>
#include <algorithm>

namespace ason {

Lexer::Lexer(std::string_view input) : input_(input) {}

char Lexer::peek() const {
    if (at_end()) return '\0';
    return input_[pos_];
}

char Lexer::peek_next() const {
    if (pos_ + 1 >= input_.size()) return '\0';
    return input_[pos_ + 1];
}

char Lexer::advance() {
    if (at_end()) return '\0';
    return input_[pos_++];
}

bool Lexer::at_end() const {
    return pos_ >= input_.size();
}

bool Lexer::is_delimiter(char c) {
    return c == ',' || c == '(' || c == ')' || c == '[' || c == ']' || c == '{' || c == '}' || c == ':';
}

void Lexer::skip_whitespace() {
    while (!at_end() && std::isspace(static_cast<unsigned char>(peek()))) {
        advance();
    }
}

Result<void> Lexer::skip_comment() {
    size_t start = pos_;
    // Already consumed '/', check for '*'
    if (peek() != '*') {
        return Result<void>(std::monostate{});
    }
    advance(); // consume '*'
    
    while (!at_end()) {
        if (advance() == '*' && peek() == '/') {
            advance(); // consume '/'
            return Result<void>(std::monostate{});
        }
    }
    
    return AsonError::unclosed_comment(start);
}

Result<void> Lexer::skip_ws_and_comments() {
    while (true) {
        skip_whitespace();
        if (peek() == '/') {
            size_t start = pos_;
            advance(); // consume '/'
            if (peek() == '*') {
                auto result = skip_comment();
                if (result.is_error()) return result.error();
            } else {
                // Not a comment, restore position
                pos_ = start;
                break;
            }
        } else {
            break;
        }
    }
    return Result<void>(std::monostate{});
}

Result<std::string> Lexer::parse_quoted_string() {
    size_t start = pos_ - 1; // -1 because we already consumed '"'
    std::string s;
    
    while (!at_end()) {
        char c = advance();
        if (c == '\\') {
            if (at_end()) {
                return AsonError::unexpected_eof(pos_);
            }
            char escaped = advance();
            switch (escaped) {
                case 'n': s.push_back('\n'); break;
                case 't': s.push_back('\t'); break;
                case '\\': s.push_back('\\'); break;
                case '"': s.push_back('"'); break;
                case ',': s.push_back(','); break;
                case '(': s.push_back('('); break;
                case ')': s.push_back(')'); break;
                case '[': s.push_back('['); break;
                case ']': s.push_back(']'); break;
                default:
                    return AsonError::invalid_escape(escaped, pos_ - 1);
            }
        } else if (c == '"') {
            return s;
        } else {
            s.push_back(c);
        }
    }
    
    return AsonError::unclosed_string(start);
}

Result<std::string> Lexer::parse_plain_string(char first) {
    std::string s;
    s.push_back(first);
    
    while (!at_end()) {
        char c = peek();
        if (is_delimiter(c)) break;
        
        if (c == '\\') {
            advance(); // consume '\'
            if (at_end()) {
                return AsonError::unexpected_eof(pos_);
            }
            char escaped = advance();
            switch (escaped) {
                case 'n': s.push_back('\n'); break;
                case 't': s.push_back('\t'); break;
                case '\\': s.push_back('\\'); break;
                case '"': s.push_back('"'); break;
                case ',': s.push_back(','); break;
                case '(': s.push_back('('); break;
                case ')': s.push_back(')'); break;
                case '[': s.push_back('['); break;
                case ']': s.push_back(']'); break;
                default:
                    return AsonError::invalid_escape(escaped, pos_ - 1);
            }
        } else {
            advance();
            s.push_back(c);
        }
    }
    
    // Trim whitespace
    auto start = s.find_first_not_of(" \t\n\r");
    auto end = s.find_last_not_of(" \t\n\r");
    if (start == std::string::npos) return std::string{};
    return s.substr(start, end - start + 1);
}

std::string Lexer::parse_ident(char first) {
    std::string s;
    s.push_back(first);

    while (!at_end()) {
        char c = peek();
        // Allow alphanumeric, underscore, and Unicode (non-ASCII)
        if (std::isalnum(static_cast<unsigned char>(c)) || c == '_' || static_cast<unsigned char>(c) > 127) {
            advance();
            s.push_back(c);
        } else {
            break;
        }
    }
    return s;
}

Result<Token> Lexer::parse_number(char first, size_t start) {
    std::string s;
    s.push_back(first);
    bool has_dot = false;

    while (!at_end()) {
        char c = peek();
        if (std::isdigit(static_cast<unsigned char>(c))) {
            advance();
            s.push_back(c);
        } else if (c == '.' && !has_dot) {
            advance();
            s.push_back('.');
            has_dot = true;
        } else if (is_delimiter(c) || std::isspace(static_cast<unsigned char>(c))) {
            break;
        } else {
            // Part of a string, continue as plain string
            return continue_as_string(std::move(s), start);
        }
    }

    if (has_dot) {
        double val = 0.0;
        auto result = std::from_chars(s.data(), s.data() + s.size(), val);
        if (result.ec != std::errc{}) {
            return AsonError::parse_error(pos_, "invalid float: " + s);
        }
        return Token(TokenType::Float, val, start, pos_);
    } else {
        int64_t val = 0;
        auto result = std::from_chars(s.data(), s.data() + s.size(), val);
        if (result.ec != std::errc{}) {
            return AsonError::parse_error(pos_, "invalid integer: " + s);
        }
        return Token(TokenType::Integer, val, start, pos_);
    }
}

Result<Token> Lexer::continue_as_string(std::string prefix, size_t start) {
    std::string s = std::move(prefix);

    while (!at_end()) {
        char c = peek();
        if (is_delimiter(c)) break;

        if (c == '\\') {
            advance();
            if (at_end()) break;
            char escaped = advance();
            switch (escaped) {
                case 'n': s.push_back('\n'); break;
                case 't': s.push_back('\t'); break;
                case '\\': s.push_back('\\'); break;
                case '"': s.push_back('"'); break;
                case ',': s.push_back(','); break;
                case '(': s.push_back('('); break;
                case ')': s.push_back(')'); break;
                case '[': s.push_back('['); break;
                case ']': s.push_back(']'); break;
                default:
                    return AsonError::invalid_escape(escaped, pos_ - 1);
            }
        } else {
            advance();
            s.push_back(c);
        }
    }

    // Trim whitespace
    auto str_start = s.find_first_not_of(" \t\n\r");
    auto str_end = s.find_last_not_of(" \t\n\r");
    if (str_start == std::string::npos) {
        return Token(TokenType::Str, std::string{}, start, pos_);
    }
    return Token(TokenType::Str, s.substr(str_start, str_end - str_start + 1), start, pos_);
}

Result<Token> Lexer::parse_value(char first, size_t start) {
    // Check for negative number
    if (first == '-' && std::isdigit(static_cast<unsigned char>(peek()))) {
        return parse_number(first, start);
    }

    // Check for positive number
    if (std::isdigit(static_cast<unsigned char>(first))) {
        return parse_number(first, start);
    }

    // Parse as plain string
    auto result = parse_plain_string(first);
    if (result.is_error()) return result.error();

    std::string s = std::move(result).value();

    if (s == "true") {
        return Token(TokenType::True, start, pos_);
    } else if (s == "false") {
        return Token(TokenType::False, start, pos_);
    } else if (s.empty()) {
        return Token(TokenType::Str, std::string{}, start, pos_);
    } else {
        // Try to parse as number
        int64_t int_val = 0;
        auto int_result = std::from_chars(s.data(), s.data() + s.size(), int_val);
        if (int_result.ec == std::errc{} && int_result.ptr == s.data() + s.size()) {
            return Token(TokenType::Integer, int_val, start, pos_);
        }

        double float_val = 0.0;
        auto float_result = std::from_chars(s.data(), s.data() + s.size(), float_val);
        if (float_result.ec == std::errc{} && float_result.ptr == s.data() + s.size()) {
            return Token(TokenType::Float, float_val, start, pos_);
        }

        return Token(TokenType::Str, std::move(s), start, pos_);
    }
}

Result<Token> Lexer::next_token() {
    auto skip_result = skip_ws_and_comments();
    if (skip_result.is_error()) return skip_result.error();

    size_t start = pos_;

    if (at_end()) {
        return Token(TokenType::Eof, start, start);
    }

    char c = advance();

    switch (c) {
        case '{':
            in_schema_ = true;
            brace_depth_++;
            return Token(TokenType::LBrace, start, pos_);
        case '}':
            if (brace_depth_ > 0) brace_depth_--;
            if (brace_depth_ == 0) in_schema_ = false;
            return Token(TokenType::RBrace, start, pos_);
        case '(':
            return Token(TokenType::LParen, start, pos_);
        case ')':
            return Token(TokenType::RParen, start, pos_);
        case '[':
            return Token(TokenType::LBracket, start, pos_);
        case ']':
            return Token(TokenType::RBracket, start, pos_);
        case ',':
            return Token(TokenType::Comma, start, pos_);
        case ':':
            return Token(TokenType::Colon, start, pos_);
        case '"': {
            auto result = parse_quoted_string();
            if (result.is_error()) return result.error();
            return Token(TokenType::Str, std::move(result).value(), start, pos_);
        }
        default:
            if (in_schema_) {
                // In schema mode, parse as identifier
                if (std::isalpha(static_cast<unsigned char>(c)) || c == '_' || static_cast<unsigned char>(c) > 127) {
                    std::string ident = parse_ident(c);
                    return Token(TokenType::Ident, std::move(ident), start, pos_);
                } else {
                    return AsonError::unexpected_char(c, start);
                }
            } else {
                // In data mode, parse as value
                return parse_value(c, start);
            }
    }
}

Result<std::vector<Token>> Lexer::tokenize() {
    std::vector<Token> tokens;
    while (true) {
        auto result = next_token();
        if (result.is_error()) return result.error();

        Token token = std::move(result).value();
        bool is_eof = token.type == TokenType::Eof;
        tokens.push_back(std::move(token));
        if (is_eof) break;
    }
    return tokens;
}

} // namespace ason

