#pragma once

#include "value.hpp"
#include <string>
#include <string_view>

namespace ason {

/// Configuration for serialization.
struct SerializeConfig {
    /// Whether to include schema in output (for objects/arrays of objects).
    bool include_schema = true;
    /// Whether to pretty print with indentation.
    bool pretty = false;
    /// Indentation string (default: 2 spaces).
    std::string indent = "  ";
    
    /// Create a compact (non-pretty) configuration.
    static SerializeConfig compact() { return SerializeConfig{}; }
    
    /// Create a pretty-printed configuration.
    static SerializeConfig pretty_config() {
        SerializeConfig config;
        config.pretty = true;
        return config;
    }
};

/// Serializer for converting Value to ASON string.
class Serializer {
public:
    explicit Serializer(const SerializeConfig& config = SerializeConfig{});
    
    /// Serialize a Value to an ASON string.
    std::string serialize(const Value& value);

private:
    SerializeConfig config_;
    std::string output_;
    size_t depth_ = 0;
    
    void serialize_impl(const Value& value);
    void serialize_string(const std::string& s);
    void serialize_array(const Array& arr);
    void serialize_array_compact(const Array& arr);
    void serialize_array_pretty(const Array& arr);
    void serialize_object(const Object& obj);
    void serialize_object_compact(const Object& obj);
    void serialize_object_pretty(const Object& obj);
    void write_indent();
};

/// Check if a string needs to be quoted.
bool needs_quotes(std::string_view s);

/// Check if a string looks like a number.
bool looks_like_number(std::string_view s);

/// Serialize a Value to an ASON string.
std::string to_string(const Value& value);

/// Serialize a Value to a pretty-printed ASON string.
std::string to_string_pretty(const Value& value);

/// Serialize a Value to an ASON string with custom configuration.
std::string to_string_with_config(const Value& value, const SerializeConfig& config);

} // namespace ason

