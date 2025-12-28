#pragma once

#include <string>
#include <stdexcept>
#include <variant>
#include <optional>

namespace ason {

/// Error types for ASON parsing and serialization.
enum class ErrorCode {
    UnexpectedEof,
    UnexpectedChar,
    InvalidEscape,
    UnclosedString,
    UnclosedComment,
    UnclosedBracket,
    FieldCountMismatch,
    InvalidSchema,
    InvalidIdentifier,
    ExpectedToken,
    ParseError
};

/// ASON Error with position and message.
class AsonError : public std::runtime_error {
public:
    AsonError(ErrorCode code, size_t position, const std::string& message)
        : std::runtime_error(message), code_(code), position_(position) {}
    
    ErrorCode code() const noexcept { return code_; }
    size_t position() const noexcept { return position_; }

    static AsonError unexpected_eof(size_t pos) {
        return AsonError(ErrorCode::UnexpectedEof, pos, "Unexpected end of input");
    }
    
    static AsonError unexpected_char(char ch, size_t pos) {
        return AsonError(ErrorCode::UnexpectedChar, pos, 
            std::string("Unexpected character '") + ch + "'");
    }
    
    static AsonError invalid_escape(char ch, size_t pos) {
        return AsonError(ErrorCode::InvalidEscape, pos,
            std::string("Invalid escape sequence '\\") + ch + "'");
    }
    
    static AsonError unclosed_string(size_t pos) {
        return AsonError(ErrorCode::UnclosedString, pos, "Unclosed string");
    }
    
    static AsonError unclosed_comment(size_t pos) {
        return AsonError(ErrorCode::UnclosedComment, pos, "Unclosed comment");
    }
    
    static AsonError unclosed_bracket(char bracket, size_t pos) {
        return AsonError(ErrorCode::UnclosedBracket, pos,
            std::string("Unclosed bracket '") + bracket + "'");
    }
    
    static AsonError field_count_mismatch(size_t expected, size_t actual, size_t pos) {
        return AsonError(ErrorCode::FieldCountMismatch, pos,
            "Field count mismatch: expected " + std::to_string(expected) + 
            ", got " + std::to_string(actual));
    }
    
    static AsonError invalid_schema(size_t pos, const std::string& msg) {
        return AsonError(ErrorCode::InvalidSchema, pos, "Invalid schema: " + msg);
    }
    
    static AsonError invalid_identifier(const std::string& name, size_t pos) {
        return AsonError(ErrorCode::InvalidIdentifier, pos,
            "Invalid identifier '" + name + "'");
    }
    
    static AsonError expected_token(const std::string& expected, const std::string& found, size_t pos) {
        return AsonError(ErrorCode::ExpectedToken, pos,
            "Expected " + expected + ", found " + found);
    }
    
    static AsonError parse_error(size_t pos, const std::string& msg) {
        return AsonError(ErrorCode::ParseError, pos, msg);
    }

private:
    ErrorCode code_;
    size_t position_;
};

/// Result type for operations that can fail.
template<typename T>
class Result {
public:
    Result(T value) : data_(std::move(value)) {}
    Result(AsonError error) : data_(std::move(error)) {}

    bool ok() const noexcept { return std::holds_alternative<T>(data_); }
    bool is_error() const noexcept { return std::holds_alternative<AsonError>(data_); }

    T& value() & { return std::get<T>(data_); }
    const T& value() const& { return std::get<T>(data_); }
    T&& value() && { return std::get<T>(std::move(data_)); }

    AsonError& error() & { return std::get<AsonError>(data_); }
    const AsonError& error() const& { return std::get<AsonError>(data_); }

    T value_or(T default_value) const {
        if (ok()) return std::get<T>(data_);
        return default_value;
    }

    /// Unwrap the value or throw the error.
    T unwrap() && {
        if (ok()) return std::get<T>(std::move(data_));
        throw std::get<AsonError>(data_);
    }

    const T& unwrap() const& {
        if (ok()) return std::get<T>(data_);
        throw std::get<AsonError>(data_);
    }

private:
    std::variant<T, AsonError> data_;
};

/// Specialization for void - uses std::monostate instead of void.
template<>
class Result<void> {
public:
    Result() : data_(std::monostate{}) {}
    Result(std::monostate) : data_(std::monostate{}) {}
    Result(AsonError error) : data_(std::move(error)) {}

    bool ok() const noexcept { return std::holds_alternative<std::monostate>(data_); }
    bool is_error() const noexcept { return std::holds_alternative<AsonError>(data_); }

    AsonError& error() & { return std::get<AsonError>(data_); }
    const AsonError& error() const& { return std::get<AsonError>(data_); }

    /// Unwrap or throw the error.
    void unwrap() const {
        if (is_error()) throw std::get<AsonError>(data_);
    }

private:
    std::variant<std::monostate, AsonError> data_;
};

} // namespace ason

