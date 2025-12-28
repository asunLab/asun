#include "ason/serializer.hpp"
#include <cctype>

namespace ason {

namespace {

/// Check if array contains complex elements (objects or arrays).
bool has_complex_elements(const Array& arr) {
    for (const auto& v : arr) {
        if (v.is_array() || v.is_object()) return true;
    }
    return false;
}

/// Check if object contains complex values (objects or arrays).
bool has_complex_values(const Object& obj) {
    for (const auto& [_, v] : obj) {
        if (v.is_array() || v.is_object()) return true;
    }
    return false;
}

} // anonymous namespace

Serializer::Serializer(const SerializeConfig& config) : config_(config) {}

std::string Serializer::serialize(const Value& value) {
    output_.clear();
    depth_ = 0;
    serialize_impl(value);
    return output_;
}

void Serializer::serialize_impl(const Value& value) {
    if (value.is_null()) {
        // Null is represented as empty in ASON
        return;
    }
    
    if (auto b = value.as_bool()) {
        output_ += *b ? "true" : "false";
    } else if (auto n = value.as_integer()) {
        output_ += std::to_string(*n);
    } else if (auto f = value.as_float()) {
        output_ += std::to_string(*f);
        // Remove trailing zeros
        auto pos = output_.find_last_not_of('0');
        if (pos != std::string::npos && output_[pos] == '.') {
            output_.erase(pos + 2); // Keep at least one digit after decimal
        } else if (pos != std::string::npos) {
            output_.erase(pos + 1);
        }
    } else if (auto s = value.as_string()) {
        serialize_string(*s);
    } else if (auto arr = value.as_array()) {
        serialize_array(*arr);
    } else if (auto obj = value.as_object()) {
        serialize_object(*obj);
    }
}

void Serializer::serialize_string(const std::string& s) {
    if (needs_quotes(s)) {
        output_ += '"';
        for (char ch : s) {
            switch (ch) {
                case '"': output_ += "\\\""; break;
                case '\\': output_ += "\\\\"; break;
                case '\n': output_ += "\\n"; break;
                case '\t': output_ += "\\t"; break;
                default: output_ += ch; break;
            }
        }
        output_ += '"';
    } else {
        // Escape delimiters in plain strings
        for (char ch : s) {
            switch (ch) {
                case ',': output_ += "\\,"; break;
                case '(': output_ += "\\("; break;
                case ')': output_ += "\\)"; break;
                case '[': output_ += "\\["; break;
                case ']': output_ += "\\]"; break;
                case '\\': output_ += "\\\\"; break;
                default: output_ += ch; break;
            }
        }
    }
}

void Serializer::serialize_array(const Array& arr) {
    if (config_.pretty && !arr.empty() && has_complex_elements(arr)) {
        serialize_array_pretty(arr);
    } else {
        serialize_array_compact(arr);
    }
}

void Serializer::serialize_array_compact(const Array& arr) {
    output_ += '[';
    for (size_t i = 0; i < arr.size(); ++i) {
        if (i > 0) output_ += ',';
        serialize_impl(arr[i]);
    }
    output_ += ']';
}

void Serializer::serialize_array_pretty(const Array& arr) {
    output_ += "[\n";
    ++depth_;
    
    for (size_t i = 0; i < arr.size(); ++i) {
        if (i > 0) output_ += ",\n";
        write_indent();
        serialize_impl(arr[i]);
    }
    
    output_ += '\n';
    --depth_;
    write_indent();
    output_ += ']';
}

void Serializer::serialize_object(const Object& obj) {
    if (config_.pretty && !obj.empty() && has_complex_values(obj)) {
        serialize_object_pretty(obj);
    } else {
        serialize_object_compact(obj);
    }
}

void Serializer::serialize_object_compact(const Object& obj) {
    output_ += '(';
    bool first = true;
    for (const auto& [_, value] : obj) {
        if (!first) output_ += ',';
        first = false;
        serialize_impl(value);
    }
    output_ += ')';
}

void Serializer::serialize_object_pretty(const Object& obj) {
    output_ += "(\n";
    ++depth_;
    
    bool first = true;
    for (const auto& [_, value] : obj) {
        if (!first) output_ += ",\n";
        first = false;
        write_indent();
        serialize_impl(value);
    }
    
    output_ += '\n';
    --depth_;
    write_indent();
    output_ += ')';
}

void Serializer::write_indent() {
    for (size_t i = 0; i < depth_; ++i) {
        output_ += config_.indent;
    }
}

bool needs_quotes(std::string_view s) {
    if (s.empty()) return true;

    // Leading or trailing whitespace
    if (std::isspace(static_cast<unsigned char>(s.front())) ||
        std::isspace(static_cast<unsigned char>(s.back()))) {
        return true;
    }

    // Contains characters that need quoting
    for (char ch : s) {
        if (ch == '"' || ch == '\n' || ch == '\t') return true;
    }

    // Looks like a boolean
    if (s == "true" || s == "false") return true;

    // Looks like a number
    if (looks_like_number(s)) return true;

    return false;
}

bool looks_like_number(std::string_view s) {
    if (s.empty()) return false;

    size_t i = 0;

    // Optional sign
    if (s[i] == '-' || s[i] == '+') {
        ++i;
        if (i >= s.size()) return false;
    }

    // Must start with digit
    if (!std::isdigit(static_cast<unsigned char>(s[i]))) return false;

    // Integer part
    while (i < s.size() && std::isdigit(static_cast<unsigned char>(s[i]))) {
        ++i;
    }

    // Optional decimal part
    if (i < s.size() && s[i] == '.') {
        ++i;
        // Must have at least one digit after decimal
        if (i >= s.size() || !std::isdigit(static_cast<unsigned char>(s[i]))) {
            return false;
        }
        while (i < s.size() && std::isdigit(static_cast<unsigned char>(s[i]))) {
            ++i;
        }
    }

    // Optional exponent
    if (i < s.size() && (s[i] == 'e' || s[i] == 'E')) {
        ++i;
        if (i < s.size() && (s[i] == '-' || s[i] == '+')) {
            ++i;
        }
        if (i >= s.size() || !std::isdigit(static_cast<unsigned char>(s[i]))) {
            return false;
        }
        while (i < s.size() && std::isdigit(static_cast<unsigned char>(s[i]))) {
            ++i;
        }
    }

    return i == s.size();
}

std::string to_string(const Value& value) {
    return to_string_with_config(value, SerializeConfig::compact());
}

std::string to_string_pretty(const Value& value) {
    return to_string_with_config(value, SerializeConfig::pretty_config());
}

std::string to_string_with_config(const Value& value, const SerializeConfig& config) {
    Serializer serializer(config);
    return serializer.serialize(value);
}

} // namespace ason

