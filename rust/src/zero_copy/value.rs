//! Zero-copy Value type that borrows strings from the input.

use indexmap::IndexMap;
use std::borrow::Cow;

/// A value that borrows strings from the input when possible.
///
/// Uses `Cow<'a, str>` to avoid allocation when the string doesn't need
/// transformation (no escape sequences). When escape sequences are present,
/// the string is allocated.
#[derive(Debug, Clone, PartialEq)]
pub enum Value<'a> {
    /// Null value.
    Null,

    /// Boolean value.
    Bool(bool),

    /// Integer value.
    Integer(i64),

    /// Floating-point value.
    Float(f64),

    /// String value (borrowed or owned).
    String(Cow<'a, str>),

    /// Array of values.
    Array(Vec<Value<'a>>),

    /// Object with string keys.
    Object(IndexMap<Cow<'a, str>, Value<'a>>),
}

impl<'a> Value<'a> {
    /// Check if this is null.
    pub fn is_null(&self) -> bool {
        matches!(self, Value::Null)
    }

    /// Check if this is a boolean.
    pub fn is_bool(&self) -> bool {
        matches!(self, Value::Bool(_))
    }

    /// Check if this is an integer.
    pub fn is_integer(&self) -> bool {
        matches!(self, Value::Integer(_))
    }

    /// Check if this is a float.
    pub fn is_float(&self) -> bool {
        matches!(self, Value::Float(_))
    }

    /// Check if this is a string.
    pub fn is_string(&self) -> bool {
        matches!(self, Value::String(_))
    }

    /// Check if this is an array.
    pub fn is_array(&self) -> bool {
        matches!(self, Value::Array(_))
    }

    /// Check if this is an object.
    pub fn is_object(&self) -> bool {
        matches!(self, Value::Object(_))
    }

    /// Get as boolean.
    pub fn as_bool(&self) -> Option<bool> {
        match self {
            Value::Bool(b) => Some(*b),
            _ => None,
        }
    }

    /// Get as i64.
    pub fn as_i64(&self) -> Option<i64> {
        match self {
            Value::Integer(n) => Some(*n),
            _ => None,
        }
    }

    /// Get as f64.
    pub fn as_f64(&self) -> Option<f64> {
        match self {
            Value::Float(n) => Some(*n),
            Value::Integer(n) => Some(*n as f64),
            _ => None,
        }
    }

    /// Get as string slice.
    pub fn as_str(&self) -> Option<&str> {
        match self {
            Value::String(s) => Some(s.as_ref()),
            _ => None,
        }
    }

    /// Get as array.
    pub fn as_array(&self) -> Option<&Vec<Value<'a>>> {
        match self {
            Value::Array(arr) => Some(arr),
            _ => None,
        }
    }

    /// Get as object.
    pub fn as_object(&self) -> Option<&IndexMap<Cow<'a, str>, Value<'a>>> {
        match self {
            Value::Object(obj) => Some(obj),
            _ => None,
        }
    }

    /// Get a field from an object by key.
    pub fn get(&self, key: &str) -> Option<&Value<'a>> {
        self.as_object().and_then(|obj| obj.get(key))
    }

    /// Get element from array by index.
    pub fn get_index(&self, index: usize) -> Option<&Value<'a>> {
        self.as_array().and_then(|arr| arr.get(index))
    }

    /// Convert to owned Value (allocates all strings).
    pub fn into_owned(self) -> crate::Value {
        match self {
            Value::Null => crate::Value::Null,
            Value::Bool(b) => crate::Value::Bool(b),
            Value::Integer(n) => crate::Value::Integer(n),
            Value::Float(n) => crate::Value::Float(n),
            Value::String(s) => crate::Value::String(s.into_owned()),
            Value::Array(arr) => {
                crate::Value::Array(arr.into_iter().map(|v| v.into_owned()).collect())
            }
            Value::Object(obj) => {
                let mut map = IndexMap::new();
                for (k, v) in obj {
                    map.insert(k.into_owned(), v.into_owned());
                }
                crate::Value::Object(map)
            }
        }
    }
}
