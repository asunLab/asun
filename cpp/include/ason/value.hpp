#pragma once

#include <string>
#include <vector>
#include <map>
#include <variant>
#include <memory>
#include <optional>
#include <cstdint>

namespace ason {

// Forward declaration
class Value;

/// Null type for ASON null values.
struct Null {
    bool operator==(const Null&) const { return true; }
    bool operator!=(const Null&) const { return false; }
};

/// Object type using ordered map (maintains insertion order in C++17 with std::map).
/// Note: For true insertion-order preservation, consider using a custom ordered_map.
using Object = std::map<std::string, Value>;

/// Array type.
using Array = std::vector<Value>;

/// ASON Value - can hold any ASON data type.
class Value {
public:
    /// Underlying variant type.
    using Variant = std::variant<Null, bool, int64_t, double, std::string, Array, Object>;
    
    // Constructors
    Value() : data_(Null{}) {}
    Value(Null) : data_(Null{}) {}
    Value(bool b) : data_(b) {}
    Value(int i) : data_(static_cast<int64_t>(i)) {}
    Value(int64_t i) : data_(i) {}
    Value(double d) : data_(d) {}
    Value(const char* s) : data_(std::string(s)) {}
    Value(std::string s) : data_(std::move(s)) {}
    Value(Array arr) : data_(std::move(arr)) {}
    Value(Object obj) : data_(std::move(obj)) {}
    
    // Type checks
    bool is_null() const { return std::holds_alternative<Null>(data_); }
    bool is_bool() const { return std::holds_alternative<bool>(data_); }
    bool is_integer() const { return std::holds_alternative<int64_t>(data_); }
    bool is_float() const { return std::holds_alternative<double>(data_); }
    bool is_string() const { return std::holds_alternative<std::string>(data_); }
    bool is_array() const { return std::holds_alternative<Array>(data_); }
    bool is_object() const { return std::holds_alternative<Object>(data_); }
    
    // Type accessors (return optional)
    std::optional<bool> as_bool() const {
        if (auto* p = std::get_if<bool>(&data_)) return *p;
        return std::nullopt;
    }

    std::optional<int64_t> as_i64() const {
        if (auto* p = std::get_if<int64_t>(&data_)) return *p;
        return std::nullopt;
    }

    // Alias for as_i64
    std::optional<int64_t> as_integer() const { return as_i64(); }

    std::optional<double> as_f64() const {
        if (auto* p = std::get_if<double>(&data_)) return *p;
        if (auto* p = std::get_if<int64_t>(&data_)) return static_cast<double>(*p);
        return std::nullopt;
    }

    // Alias for as_f64
    std::optional<double> as_float() const { return as_f64(); }

    std::optional<std::string_view> as_str() const {
        if (auto* p = std::get_if<std::string>(&data_)) return std::string_view(*p);
        return std::nullopt;
    }

    // Return pointer to string for easier use
    const std::string* as_string() const {
        return std::get_if<std::string>(&data_);
    }
    
    const Array* as_array() const {
        return std::get_if<Array>(&data_);
    }
    
    Array* as_array() {
        return std::get_if<Array>(&data_);
    }
    
    const Object* as_object() const {
        return std::get_if<Object>(&data_);
    }
    
    Object* as_object() {
        return std::get_if<Object>(&data_);
    }
    
    /// Get a value from an object by key.
    const Value* get(const std::string& key) const {
        if (auto* obj = as_object()) {
            auto it = obj->find(key);
            if (it != obj->end()) return &it->second;
        }
        return nullptr;
    }
    
    /// Get a value from an array by index.
    const Value* get(size_t index) const {
        if (auto* arr = as_array()) {
            if (index < arr->size()) return &(*arr)[index];
        }
        return nullptr;
    }

    /// Alias for get(size_t) for compatibility.
    const Value* get_index(size_t index) const { return get(index); }
    
    /// Access the underlying variant.
    const Variant& variant() const { return data_; }
    Variant& variant() { return data_; }
    
    // Comparison
    bool operator==(const Value& other) const { return data_ == other.data_; }
    bool operator!=(const Value& other) const { return data_ != other.data_; }

private:
    Variant data_;
};

} // namespace ason

