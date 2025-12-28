//! Serializer for ASON format.
//!
//! Converts Value to ASON string representation.

use crate::ast::Value;
use std::fmt::Write;

/// Configuration for serialization.
#[derive(Debug, Clone)]
pub struct SerializeConfig {
    /// Whether to include schema in output (for objects/arrays of objects).
    pub include_schema: bool,
    /// Whether to pretty print with indentation.
    pub pretty: bool,
    /// Indentation string (default: 2 spaces).
    pub indent: String,
}

impl Default for SerializeConfig {
    fn default() -> Self {
        Self {
            include_schema: true,
            pretty: false,
            indent: "  ".to_string(),
        }
    }
}

impl SerializeConfig {
    /// Create a compact (non-pretty) configuration.
    pub fn compact() -> Self {
        Self::default()
    }

    /// Create a pretty-printed configuration.
    pub fn pretty() -> Self {
        Self {
            pretty: true,
            ..Self::default()
        }
    }
}

/// Serialize a Value to an ASON string.
pub fn to_string(value: &Value) -> String {
    to_string_with_config(value, &SerializeConfig::default())
}

/// Serialize a Value to a pretty-printed ASON string.
pub fn to_string_pretty(value: &Value) -> String {
    to_string_with_config(value, &SerializeConfig::pretty())
}

/// Serialize a Value to an ASON string with custom configuration.
pub fn to_string_with_config(value: &Value, config: &SerializeConfig) -> String {
    let mut output = String::new();
    let mut serializer = Serializer::new(&mut output, config);
    serializer.serialize(value);
    output
}

/// Serializer for converting Value to ASON string.
struct Serializer<'a> {
    output: &'a mut String,
    config: &'a SerializeConfig,
    depth: usize,
}

impl<'a> Serializer<'a> {
    fn new(output: &'a mut String, config: &'a SerializeConfig) -> Self {
        Self {
            output,
            config,
            depth: 0,
        }
    }

    fn serialize(&mut self, value: &Value) {
        match value {
            Value::Null => {
                // Null is represented as empty in ASON
            }
            Value::Bool(b) => {
                write!(self.output, "{}", b).unwrap();
            }
            Value::Integer(n) => {
                write!(self.output, "{}", n).unwrap();
            }
            Value::Float(n) => {
                write!(self.output, "{}", n).unwrap();
            }
            Value::String(s) => {
                self.serialize_string(s);
            }
            Value::Array(arr) => {
                self.serialize_array(arr);
            }
            Value::Object(obj) => {
                self.serialize_object(obj);
            }
        }
    }

    /// Serialize a string, adding quotes if necessary.
    fn serialize_string(&mut self, s: &str) {
        if needs_quotes(s) {
            self.output.push('"');
            for ch in s.chars() {
                match ch {
                    '"' => self.output.push_str("\\\""),
                    '\\' => self.output.push_str("\\\\"),
                    '\n' => self.output.push_str("\\n"),
                    '\t' => self.output.push_str("\\t"),
                    _ => self.output.push(ch),
                }
            }
            self.output.push('"');
        } else {
            // Escape delimiters in plain strings
            for ch in s.chars() {
                match ch {
                    ',' => self.output.push_str("\\,"),
                    '(' => self.output.push_str("\\("),
                    ')' => self.output.push_str("\\)"),
                    '[' => self.output.push_str("\\["),
                    ']' => self.output.push_str("\\]"),
                    '\\' => self.output.push_str("\\\\"),
                    _ => self.output.push(ch),
                }
            }
        }
    }

    /// Serialize an array.
    fn serialize_array(&mut self, arr: &[Value]) {
        if self.config.pretty && !arr.is_empty() && has_complex_elements(arr) {
            self.serialize_array_pretty(arr);
        } else {
            self.serialize_array_compact(arr);
        }
    }

    fn serialize_array_compact(&mut self, arr: &[Value]) {
        self.output.push('[');
        for (i, item) in arr.iter().enumerate() {
            if i > 0 {
                self.output.push(',');
            }
            self.serialize(item);
        }
        self.output.push(']');
    }

    fn serialize_array_pretty(&mut self, arr: &[Value]) {
        self.output.push('[');
        self.output.push('\n');
        self.depth += 1;

        for (i, item) in arr.iter().enumerate() {
            if i > 0 {
                self.output.push(',');
                self.output.push('\n');
            }
            self.write_indent();
            self.serialize(item);
        }

        self.output.push('\n');
        self.depth -= 1;
        self.write_indent();
        self.output.push(']');
    }

    /// Serialize object data.
    fn serialize_object(&mut self, obj: &indexmap::IndexMap<String, Value>) {
        if self.config.pretty && !obj.is_empty() && has_complex_values(obj) {
            self.serialize_object_pretty(obj);
        } else {
            self.serialize_object_compact(obj);
        }
    }

    fn serialize_object_compact(&mut self, obj: &indexmap::IndexMap<String, Value>) {
        self.output.push('(');
        for (i, (_, value)) in obj.iter().enumerate() {
            if i > 0 {
                self.output.push(',');
            }
            self.serialize(value);
        }
        self.output.push(')');
    }

    fn serialize_object_pretty(&mut self, obj: &indexmap::IndexMap<String, Value>) {
        self.output.push('(');
        self.output.push('\n');
        self.depth += 1;

        for (i, (_, value)) in obj.iter().enumerate() {
            if i > 0 {
                self.output.push(',');
                self.output.push('\n');
            }
            self.write_indent();
            self.serialize(value);
        }

        self.output.push('\n');
        self.depth -= 1;
        self.write_indent();
        self.output.push(')');
    }

    fn write_indent(&mut self) {
        for _ in 0..self.depth {
            self.output.push_str(&self.config.indent);
        }
    }
}

/// Check if array contains complex elements (objects or arrays).
fn has_complex_elements(arr: &[Value]) -> bool {
    arr.iter()
        .any(|v| matches!(v, Value::Object(_) | Value::Array(_)))
}

/// Check if object contains complex values (objects or arrays).
fn has_complex_values(obj: &indexmap::IndexMap<String, Value>) -> bool {
    obj.values()
        .any(|v| matches!(v, Value::Object(_) | Value::Array(_)))
}

/// Determine if a string needs to be quoted.
///
/// A string needs quotes if:
/// - It's empty
/// - It has leading or trailing whitespace
/// - It contains special characters that would be ambiguous
/// - It looks like a number or boolean
fn needs_quotes(s: &str) -> bool {
    if s.is_empty() {
        return true;
    }

    // Leading or trailing whitespace
    if s != s.trim() {
        return true;
    }

    // Contains characters that need quoting
    if s.contains('"') || s.contains('\n') || s.contains('\t') {
        return true;
    }

    // Looks like a boolean
    if s == "true" || s == "false" {
        return true;
    }

    // Looks like a number
    if looks_like_number(s) {
        return true;
    }

    false
}

/// Check if a string looks like a number.
fn looks_like_number(s: &str) -> bool {
    if s.is_empty() {
        return false;
    }

    let chars: Vec<char> = s.chars().collect();
    let mut i = 0;

    // Optional sign
    if chars[i] == '-' || chars[i] == '+' {
        i += 1;
        if i >= chars.len() {
            return false;
        }
    }

    // Must start with digit
    if !chars[i].is_ascii_digit() {
        return false;
    }

    // Integer part
    while i < chars.len() && chars[i].is_ascii_digit() {
        i += 1;
    }

    // Optional decimal part
    if i < chars.len() && chars[i] == '.' {
        i += 1;
        // Must have at least one digit after decimal
        if i >= chars.len() || !chars[i].is_ascii_digit() {
            return false;
        }
        while i < chars.len() && chars[i].is_ascii_digit() {
            i += 1;
        }
    }

    // Optional exponent
    if i < chars.len() && (chars[i] == 'e' || chars[i] == 'E') {
        i += 1;
        if i < chars.len() && (chars[i] == '-' || chars[i] == '+') {
            i += 1;
        }
        if i >= chars.len() || !chars[i].is_ascii_digit() {
            return false;
        }
        while i < chars.len() && chars[i].is_ascii_digit() {
            i += 1;
        }
    }

    i == chars.len()
}

#[cfg(test)]
mod tests {
    use super::*;
    use indexmap::IndexMap;

    #[test]
    fn test_serialize_primitives() {
        assert_eq!(to_string(&Value::Bool(true)), "true");
        assert_eq!(to_string(&Value::Bool(false)), "false");
        assert_eq!(to_string(&Value::Integer(42)), "42");
        assert_eq!(to_string(&Value::Float(3.14)), "3.14");
    }

    #[test]
    fn test_serialize_string_no_quotes() {
        assert_eq!(to_string(&Value::String("hello".to_string())), "hello");
        assert_eq!(to_string(&Value::String("Alice".to_string())), "Alice");
    }

    #[test]
    fn test_serialize_string_with_quotes() {
        // Empty string needs quotes
        assert_eq!(to_string(&Value::String("".to_string())), r#""""#);

        // String that looks like boolean
        assert_eq!(to_string(&Value::String("true".to_string())), r#""true""#);

        // String that looks like number
        assert_eq!(to_string(&Value::String("123".to_string())), r#""123""#);

        // String with whitespace
        assert_eq!(
            to_string(&Value::String("  hello  ".to_string())),
            r#""  hello  ""#
        );
    }

    #[test]
    fn test_serialize_string_escapes() {
        // String with internal quotes
        assert_eq!(
            to_string(&Value::String(r#"say "hi""#.to_string())),
            r#""say \"hi\"""#
        );
    }

    #[test]
    fn test_serialize_array() {
        let arr = Value::Array(vec![
            Value::Integer(1),
            Value::Integer(2),
            Value::Integer(3),
        ]);
        assert_eq!(to_string(&arr), "[1,2,3]");
    }

    #[test]
    fn test_serialize_mixed_array() {
        let arr = Value::Array(vec![
            Value::Integer(1),
            Value::String("hello".to_string()),
            Value::Bool(true),
        ]);
        assert_eq!(to_string(&arr), "[1,hello,true]");
    }

    #[test]
    fn test_serialize_object() {
        let mut obj = IndexMap::new();
        obj.insert("name".to_string(), Value::String("Alice".to_string()));
        obj.insert("age".to_string(), Value::Integer(30));
        assert_eq!(to_string(&Value::Object(obj)), "(Alice,30)");
    }

    #[test]
    fn test_serialize_nested() {
        let mut inner = IndexMap::new();
        inner.insert("city".to_string(), Value::String("NYC".to_string()));
        inner.insert("zip".to_string(), Value::Integer(10001));

        let mut outer = IndexMap::new();
        outer.insert("name".to_string(), Value::String("Alice".to_string()));
        outer.insert("addr".to_string(), Value::Object(inner));

        assert_eq!(to_string(&Value::Object(outer)), "(Alice,(NYC,10001))");
    }

    #[test]
    fn test_serialize_null_in_object() {
        let mut obj = IndexMap::new();
        obj.insert("name".to_string(), Value::String("Alice".to_string()));
        obj.insert("age".to_string(), Value::Null);
        assert_eq!(to_string(&Value::Object(obj)), "(Alice,)");
    }

    #[test]
    fn test_needs_quotes() {
        // Should NOT need quotes
        assert!(!needs_quotes("hello"));
        assert!(!needs_quotes("Alice"));
        assert!(!needs_quotes("hello-world"));
        assert!(!needs_quotes("foo_bar"));

        // Should need quotes
        assert!(needs_quotes(""));
        assert!(needs_quotes("true"));
        assert!(needs_quotes("false"));
        assert!(needs_quotes("123"));
        assert!(needs_quotes("3.14"));
        assert!(needs_quotes("  hello"));
        assert!(needs_quotes("hello  "));
        assert!(needs_quotes("say \"hi\""));
    }

    #[test]
    fn test_looks_like_number() {
        assert!(looks_like_number("123"));
        assert!(looks_like_number("-123"));
        assert!(looks_like_number("3.14"));
        assert!(looks_like_number("-3.14"));
        assert!(looks_like_number("1e10"));
        assert!(looks_like_number("1.5e-3"));

        assert!(!looks_like_number(""));
        assert!(!looks_like_number("abc"));
        assert!(!looks_like_number("123abc"));
        assert!(!looks_like_number("-"));
        assert!(!looks_like_number("1."));
    }

    #[test]
    fn test_pretty_array_of_objects() {
        let mut obj1 = IndexMap::new();
        obj1.insert("name".to_string(), Value::String("Alice".to_string()));
        obj1.insert("age".to_string(), Value::Integer(30));

        let mut obj2 = IndexMap::new();
        obj2.insert("name".to_string(), Value::String("Bob".to_string()));
        obj2.insert("age".to_string(), Value::Integer(25));

        let arr = Value::Array(vec![Value::Object(obj1), Value::Object(obj2)]);

        let pretty = to_string_pretty(&arr);
        assert!(pretty.contains('\n'));
        assert!(pretty.contains("  ")); // indentation
        assert!(pretty.starts_with('['));
        assert!(pretty.ends_with(']'));
    }

    #[test]
    fn test_pretty_nested_object() {
        let mut inner = IndexMap::new();
        inner.insert("city".to_string(), Value::String("NYC".to_string()));
        inner.insert("zip".to_string(), Value::Integer(10001));

        let mut outer = IndexMap::new();
        outer.insert("name".to_string(), Value::String("Alice".to_string()));
        outer.insert("addr".to_string(), Value::Object(inner));

        let pretty = to_string_pretty(&Value::Object(outer));
        assert!(pretty.contains('\n'));
        assert!(pretty.contains("  ")); // indentation
    }

    #[test]
    fn test_pretty_simple_stays_compact() {
        // Simple values without nesting should stay compact
        let arr = Value::Array(vec![
            Value::Integer(1),
            Value::Integer(2),
            Value::Integer(3),
        ]);
        let pretty = to_string_pretty(&arr);
        assert_eq!(pretty, "[1,2,3]");
    }

    #[test]
    fn test_pretty_simple_object_stays_compact() {
        let mut obj = IndexMap::new();
        obj.insert("name".to_string(), Value::String("Alice".to_string()));
        obj.insert("age".to_string(), Value::Integer(30));

        let pretty = to_string_pretty(&Value::Object(obj));
        assert_eq!(pretty, "(Alice,30)");
    }
}
