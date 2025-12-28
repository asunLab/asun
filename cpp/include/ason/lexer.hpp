#pragma once

#include "error.hpp"
#include <string>
#include <string_view>
#include <variant>
#include <vector>
#include <cstdint>

namespace ason {

/// Token types for ASON lexer.
enum class TokenType {
    LBrace,     // {
    RBrace,     // }
    LParen,     // (
    RParen,     // )
    LBracket,   // [
    RBracket,   // ]
    Comma,      // ,
    Colon,      // :
    Ident,      // identifier (field name)
    Str,        // string value
    Integer,    // integer value
    Float,      // float value
    True,       // true
    False,      // false
    Eof         // end of input
};

/// Token with optional value.
struct Token {
    TokenType type;
    std::variant<std::monostate, std::string, int64_t, double> value;
    size_t start;
    size_t end;
    
    Token(TokenType t, size_t s, size_t e) : type(t), value(), start(s), end(e) {}
    Token(TokenType t, std::string v, size_t s, size_t e) : type(t), value(std::move(v)), start(s), end(e) {}
    Token(TokenType t, int64_t v, size_t s, size_t e) : type(t), value(v), start(s), end(e) {}
    Token(TokenType t, double v, size_t s, size_t e) : type(t), value(v), start(s), end(e) {}
    
    const std::string* str() const { return std::get_if<std::string>(&value); }
    const int64_t* integer() const { return std::get_if<int64_t>(&value); }
    const double* floating() const { return std::get_if<double>(&value); }
};

/// Lexer for tokenizing ASON input.
class Lexer {
public:
    explicit Lexer(std::string_view input);
    
    /// Get the next token.
    Result<Token> next_token();
    
    /// Tokenize entire input.
    Result<std::vector<Token>> tokenize();
    
    /// Get current position.
    size_t position() const { return pos_; }

private:
    std::string_view input_;
    size_t pos_ = 0;
    bool in_schema_ = false;
    size_t brace_depth_ = 0;
    
    // Character operations
    char peek() const;
    char peek_next() const;
    char advance();
    bool at_end() const;
    
    // Parsing helpers
    void skip_whitespace();
    Result<void> skip_comment();
    Result<void> skip_ws_and_comments();
    
    Result<std::string> parse_quoted_string();
    Result<std::string> parse_plain_string(char first);
    std::string parse_ident(char first);
    Result<Token> parse_value(char first, size_t start);
    Result<Token> parse_number(char first, size_t start);
    Result<Token> continue_as_string(std::string prefix, size_t start);
    
    static bool is_delimiter(char c);
};

} // namespace ason

