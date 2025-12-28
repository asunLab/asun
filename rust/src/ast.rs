//! AST (Abstract Syntax Tree) types for ASON.
//!
//! This module defines the core data structures used to represent ASON documents.

use indexmap::IndexMap;
use std::fmt;

/// Represents any ASON value.
#[derive(Debug, Clone, PartialEq)]
pub enum Value {
    /// Null value (represented as empty/blank in ASON).
    Null,

    /// Boolean value (`true` or `false`).
    Bool(bool),

    /// Integer value.
    Integer(i64),

    /// Floating-point value.
    Float(f64),

    /// String value.
    String(String),

    /// Array of values.
    Array(Vec<Value>),

    /// Object with string keys and values.
    /// Uses IndexMap to preserve insertion order.
    Object(IndexMap<String, Value>),
}

/// Represents a field in an ASON schema.
#[derive(Debug, Clone, PartialEq)]
pub enum SchemaField {
    /// Simple field: `name`
    Simple(String),

    /// Nested object field: `addr{city,zip}`
    NestedObject { name: String, schema: Schema },

    /// Simple array field: `scores[]`
    SimpleArray(String),

    /// Object array field: `users[{id,name}]`
    ObjectArray { name: String, schema: Schema },
}

/// Represents an ASON schema definition.
///
/// A schema defines the structure of objects, specifying field names
/// and their types (simple, nested, or array).
#[derive(Debug, Clone, PartialEq, Default)]
pub struct Schema {
    /// Ordered list of fields in this schema.
    pub fields: Vec<SchemaField>,
}

impl SchemaField {
    /// Get the field name.
    pub fn name(&self) -> &str {
        match self {
            SchemaField::Simple(name) => name,
            SchemaField::NestedObject { name, .. } => name,
            SchemaField::SimpleArray(name) => name,
            SchemaField::ObjectArray { name, .. } => name,
        }
    }

    /// Check if this field is an array type.
    pub fn is_array(&self) -> bool {
        matches!(
            self,
            SchemaField::SimpleArray(_) | SchemaField::ObjectArray { .. }
        )
    }

    /// Check if this field has a nested schema.
    pub fn has_schema(&self) -> bool {
        matches!(
            self,
            SchemaField::NestedObject { .. } | SchemaField::ObjectArray { .. }
        )
    }

    /// Get the nested schema if present.
    pub fn schema(&self) -> Option<&Schema> {
        match self {
            SchemaField::NestedObject { schema, .. } => Some(schema),
            SchemaField::ObjectArray { schema, .. } => Some(schema),
            _ => None,
        }
    }
}

impl Schema {
    /// Create a new empty schema.
    pub fn new() -> Self {
        Self { fields: Vec::new() }
    }

    /// Create a schema with the given fields.
    pub fn with_fields(fields: Vec<SchemaField>) -> Self {
        Self { fields }
    }

    /// Get the number of fields in this schema.
    pub fn len(&self) -> usize {
        self.fields.len()
    }

    /// Check if the schema is empty.
    pub fn is_empty(&self) -> bool {
        self.fields.is_empty()
    }

    /// Get field names as a vector.
    pub fn field_names(&self) -> Vec<&str> {
        self.fields.iter().map(|f| f.name()).collect()
    }
}

// ============================================================================
// Value implementation
// ============================================================================

impl Value {
    // ------------------------------------------------------------------------
    // Type checking methods
    // ------------------------------------------------------------------------

    /// Returns `true` if the value is null.
    pub fn is_null(&self) -> bool {
        matches!(self, Value::Null)
    }

    /// Returns `true` if the value is a boolean.
    pub fn is_bool(&self) -> bool {
        matches!(self, Value::Bool(_))
    }

    /// Returns `true` if the value is an integer.
    pub fn is_integer(&self) -> bool {
        matches!(self, Value::Integer(_))
    }

    /// Returns `true` if the value is a float.
    pub fn is_float(&self) -> bool {
        matches!(self, Value::Float(_))
    }

    /// Returns `true` if the value is a number (integer or float).
    pub fn is_number(&self) -> bool {
        matches!(self, Value::Integer(_) | Value::Float(_))
    }

    /// Returns `true` if the value is a string.
    pub fn is_string(&self) -> bool {
        matches!(self, Value::String(_))
    }

    /// Returns `true` if the value is an array.
    pub fn is_array(&self) -> bool {
        matches!(self, Value::Array(_))
    }

    /// Returns `true` if the value is an object.
    pub fn is_object(&self) -> bool {
        matches!(self, Value::Object(_))
    }

    // ------------------------------------------------------------------------
    // Accessor methods
    // ------------------------------------------------------------------------

    /// Returns the boolean value if this is a Bool, otherwise None.
    pub fn as_bool(&self) -> Option<bool> {
        match self {
            Value::Bool(b) => Some(*b),
            _ => None,
        }
    }

    /// Returns the integer value if this is an Integer, otherwise None.
    pub fn as_i64(&self) -> Option<i64> {
        match self {
            Value::Integer(n) => Some(*n),
            _ => None,
        }
    }

    /// Returns the float value if this is a Float, otherwise None.
    pub fn as_f64(&self) -> Option<f64> {
        match self {
            Value::Float(n) => Some(*n),
            Value::Integer(n) => Some(*n as f64),
            _ => None,
        }
    }

    /// Returns the string value if this is a String, otherwise None.
    pub fn as_str(&self) -> Option<&str> {
        match self {
            Value::String(s) => Some(s),
            _ => None,
        }
    }

    /// Returns the array if this is an Array, otherwise None.
    pub fn as_array(&self) -> Option<&Vec<Value>> {
        match self {
            Value::Array(arr) => Some(arr),
            _ => None,
        }
    }

    /// Returns the object if this is an Object, otherwise None.
    pub fn as_object(&self) -> Option<&IndexMap<String, Value>> {
        match self {
            Value::Object(obj) => Some(obj),
            _ => None,
        }
    }

    // ------------------------------------------------------------------------
    // Mutable accessor methods
    // ------------------------------------------------------------------------

    /// Returns a mutable reference to the array if this is an Array.
    pub fn as_array_mut(&mut self) -> Option<&mut Vec<Value>> {
        match self {
            Value::Array(arr) => Some(arr),
            _ => None,
        }
    }

    /// Returns a mutable reference to the object if this is an Object.
    pub fn as_object_mut(&mut self) -> Option<&mut IndexMap<String, Value>> {
        match self {
            Value::Object(obj) => Some(obj),
            _ => None,
        }
    }

    // ------------------------------------------------------------------------
    // Object access by key
    // ------------------------------------------------------------------------

    /// Get a value from an object by key.
    pub fn get(&self, key: &str) -> Option<&Value> {
        self.as_object().and_then(|obj| obj.get(key))
    }

    /// Get a value from an array by index.
    pub fn get_index(&self, index: usize) -> Option<&Value> {
        self.as_array().and_then(|arr| arr.get(index))
    }
}

// ============================================================================
// Display implementations
// ============================================================================

impl fmt::Display for Value {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Value::Null => write!(f, "null"),
            Value::Bool(b) => write!(f, "{}", b),
            Value::Integer(n) => write!(f, "{}", n),
            Value::Float(n) => write!(f, "{}", n),
            Value::String(s) => write!(f, "{}", s),
            Value::Array(arr) => {
                write!(f, "[")?;
                for (i, v) in arr.iter().enumerate() {
                    if i > 0 {
                        write!(f, ",")?;
                    }
                    write!(f, "{}", v)?;
                }
                write!(f, "]")
            }
            Value::Object(obj) => {
                write!(f, "(")?;
                for (i, (_, v)) in obj.iter().enumerate() {
                    if i > 0 {
                        write!(f, ",")?;
                    }
                    write!(f, "{}", v)?;
                }
                write!(f, ")")
            }
        }
    }
}

// ============================================================================
// From implementations for convenient construction
// ============================================================================

impl From<bool> for Value {
    fn from(v: bool) -> Self {
        Value::Bool(v)
    }
}

impl From<i64> for Value {
    fn from(v: i64) -> Self {
        Value::Integer(v)
    }
}

impl From<i32> for Value {
    fn from(v: i32) -> Self {
        Value::Integer(v as i64)
    }
}

impl From<f64> for Value {
    fn from(v: f64) -> Self {
        Value::Float(v)
    }
}

impl From<f32> for Value {
    fn from(v: f32) -> Self {
        Value::Float(v as f64)
    }
}

impl From<String> for Value {
    fn from(v: String) -> Self {
        Value::String(v)
    }
}

impl From<&str> for Value {
    fn from(v: &str) -> Self {
        Value::String(v.to_string())
    }
}

impl<T: Into<Value>> From<Vec<T>> for Value {
    fn from(v: Vec<T>) -> Self {
        Value::Array(v.into_iter().map(Into::into).collect())
    }
}

impl From<IndexMap<String, Value>> for Value {
    fn from(v: IndexMap<String, Value>) -> Self {
        Value::Object(v)
    }
}

impl<T: Into<Value>> From<Option<T>> for Value {
    fn from(v: Option<T>) -> Self {
        match v {
            Some(v) => v.into(),
            None => Value::Null,
        }
    }
}

// ============================================================================
// Tests
// ============================================================================

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_value_type_checks() {
        assert!(Value::Null.is_null());
        assert!(Value::Bool(true).is_bool());
        assert!(Value::Integer(42).is_integer());
        assert!(Value::Float(3.14).is_float());
        assert!(Value::String("hello".to_string()).is_string());
        assert!(Value::Array(vec![]).is_array());
        assert!(Value::Object(IndexMap::new()).is_object());
    }

    #[test]
    fn test_value_accessors() {
        assert_eq!(Value::Bool(true).as_bool(), Some(true));
        assert_eq!(Value::Integer(42).as_i64(), Some(42));
        assert_eq!(Value::Float(3.14).as_f64(), Some(3.14));
        assert_eq!(Value::Integer(42).as_f64(), Some(42.0));
        assert_eq!(Value::String("hello".to_string()).as_str(), Some("hello"));
    }

    #[test]
    fn test_value_from_impls() {
        let v: Value = true.into();
        assert_eq!(v, Value::Bool(true));

        let v: Value = 42i64.into();
        assert_eq!(v, Value::Integer(42));

        let v: Value = 3.14f64.into();
        assert_eq!(v, Value::Float(3.14));

        let v: Value = "hello".into();
        assert_eq!(v, Value::String("hello".to_string()));

        let v: Value = vec![1i64, 2i64, 3i64].into();
        assert_eq!(
            v,
            Value::Array(vec![
                Value::Integer(1),
                Value::Integer(2),
                Value::Integer(3)
            ])
        );
    }

    #[test]
    fn test_value_from_option() {
        let v: Value = Some(42i64).into();
        assert_eq!(v, Value::Integer(42));

        let v: Value = None::<i64>.into();
        assert_eq!(v, Value::Null);
    }

    #[test]
    fn test_schema_field() {
        let field = SchemaField::Simple("name".to_string());
        assert_eq!(field.name(), "name");
        assert!(!field.is_array());
        assert!(!field.has_schema());

        let field = SchemaField::SimpleArray("scores".to_string());
        assert_eq!(field.name(), "scores");
        assert!(field.is_array());
        assert!(!field.has_schema());

        let inner_schema = Schema::with_fields(vec![
            SchemaField::Simple("city".to_string()),
            SchemaField::Simple("zip".to_string()),
        ]);
        let field = SchemaField::NestedObject {
            name: "addr".to_string(),
            schema: inner_schema.clone(),
        };
        assert_eq!(field.name(), "addr");
        assert!(!field.is_array());
        assert!(field.has_schema());
        assert_eq!(field.schema(), Some(&inner_schema));
    }

    #[test]
    fn test_schema() {
        let schema = Schema::with_fields(vec![
            SchemaField::Simple("name".to_string()),
            SchemaField::Simple("age".to_string()),
        ]);
        assert_eq!(schema.len(), 2);
        assert!(!schema.is_empty());
        assert_eq!(schema.field_names(), vec!["name", "age"]);
    }

    #[test]
    fn test_value_display() {
        assert_eq!(Value::Null.to_string(), "null");
        assert_eq!(Value::Bool(true).to_string(), "true");
        assert_eq!(Value::Integer(42).to_string(), "42");
        assert_eq!(Value::Float(3.14).to_string(), "3.14");
        assert_eq!(Value::String("hello".to_string()).to_string(), "hello");
        assert_eq!(
            Value::Array(vec![Value::Integer(1), Value::Integer(2)]).to_string(),
            "[1,2]"
        );
    }

    #[test]
    fn test_object_access() {
        let mut obj = IndexMap::new();
        obj.insert("name".to_string(), Value::String("Alice".to_string()));
        obj.insert("age".to_string(), Value::Integer(30));
        let v = Value::Object(obj);

        assert_eq!(v.get("name"), Some(&Value::String("Alice".to_string())));
        assert_eq!(v.get("age"), Some(&Value::Integer(30)));
        assert_eq!(v.get("unknown"), None);
    }

    #[test]
    fn test_array_access() {
        let v = Value::Array(vec![
            Value::Integer(1),
            Value::Integer(2),
            Value::Integer(3),
        ]);

        assert_eq!(v.get_index(0), Some(&Value::Integer(1)));
        assert_eq!(v.get_index(2), Some(&Value::Integer(3)));
        assert_eq!(v.get_index(10), None);
    }
}
