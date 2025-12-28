//! Parser for ASON format.
//!
//! Converts a token stream into an AST (Abstract Syntax Tree).

use crate::ast::{Schema, SchemaField, Value};
use crate::error::{AsonError, Result};
use crate::lexer::{Lexer, Spanned, Token};
use indexmap::IndexMap;

/// Parser for ASON input.
pub struct Parser<'a> {
    lexer: Lexer<'a>,
    /// Current token
    current: Spanned,
    /// Peeked token (lookahead)
    peeked: Option<Spanned>,
}

impl<'a> Parser<'a> {
    /// Create a new parser for the given input.
    pub fn new(input: &'a str) -> Result<Self> {
        let mut lexer = Lexer::new(input);
        let current = lexer.next_token()?;
        Ok(Self {
            lexer,
            current,
            peeked: None,
        })
    }

    /// Get the current token.
    fn current(&self) -> &Token {
        &self.current.token
    }

    /// Get the current position.
    fn position(&self) -> usize {
        self.current.start
    }

    /// Advance to the next token.
    fn advance(&mut self) -> Result<()> {
        if let Some(peeked) = self.peeked.take() {
            self.current = peeked;
        } else {
            self.current = self.lexer.next_token()?;
        }
        Ok(())
    }

    /// Expect the current token to be a specific type, then advance.
    fn expect(&mut self, expected: Token) -> Result<()> {
        if self.current() == &expected {
            self.advance()
        } else {
            Err(AsonError::ExpectedToken {
                expected: format!("{:?}", expected),
                found: format!("{:?}", self.current()),
                position: self.position(),
            })
        }
    }

    /// Check if the current token matches.
    fn check(&self, token: &Token) -> bool {
        self.current() == token
    }

    /// Parse the entire ASON input.
    pub fn parse(&mut self) -> Result<Value> {
        // Check what kind of input we have
        match self.current() {
            Token::LBrace => {
                // Schema expression: {fields}:data
                let schema = self.parse_schema()?;
                self.expect(Token::Colon)?;
                self.parse_data_with_schema(&schema)
            }
            Token::LBracket => {
                // Plain array
                self.parse_array()
            }
            Token::Eof => Ok(Value::Null),
            _ => {
                // Single value
                self.parse_value()
            }
        }
    }

    /// Parse a schema definition: {field1,field2,...}
    fn parse_schema(&mut self) -> Result<Schema> {
        self.expect(Token::LBrace)?;

        let mut fields = Vec::new();

        // Handle empty schema
        if self.check(&Token::RBrace) {
            self.advance()?;
            return Ok(Schema::with_fields(fields));
        }

        // Parse first field
        fields.push(self.parse_schema_field()?);

        // Parse remaining fields
        while self.check(&Token::Comma) {
            self.advance()?;
            if self.check(&Token::RBrace) {
                break; // Trailing comma
            }
            fields.push(self.parse_schema_field()?);
        }

        self.expect(Token::RBrace)?;
        Ok(Schema::with_fields(fields))
    }

    /// Parse a single schema field.
    fn parse_schema_field(&mut self) -> Result<SchemaField> {
        // Expect an identifier
        let name = match self.current().clone() {
            Token::Ident(name) => {
                self.advance()?;
                name
            }
            _ => {
                return Err(AsonError::InvalidSchema {
                    position: self.position(),
                    message: format!("expected field name, found {:?}", self.current()),
                });
            }
        };

        // Check what follows: {, [, or nothing
        match self.current() {
            Token::LBrace => {
                // Nested object: name{...}
                let schema = self.parse_schema()?;
                Ok(SchemaField::NestedObject { name, schema })
            }
            Token::LBracket => {
                self.advance()?;
                if self.check(&Token::LBrace) {
                    // Object array: name[{...}]
                    let schema = self.parse_schema()?;
                    self.expect(Token::RBracket)?;
                    Ok(SchemaField::ObjectArray { name, schema })
                } else if self.check(&Token::RBracket) {
                    // Simple array: name[]
                    self.advance()?;
                    Ok(SchemaField::SimpleArray(name))
                } else {
                    Err(AsonError::InvalidSchema {
                        position: self.position(),
                        message: format!(
                            "expected '{{' or ']' after '[', found {:?}",
                            self.current()
                        ),
                    })
                }
            }
            _ => {
                // Simple field
                Ok(SchemaField::Simple(name))
            }
        }
    }

    /// Parse data with a schema and return a Value.
    fn parse_data_with_schema(&mut self, schema: &Schema) -> Result<Value> {
        // Could be single object or multiple objects
        let first = self.parse_object_with_schema(schema)?;

        // Check if there are more objects
        if self.check(&Token::Comma) {
            let mut objects = vec![first];
            while self.check(&Token::Comma) {
                self.advance()?;
                if self.check(&Token::Eof) {
                    break;
                }
                objects.push(self.parse_object_with_schema(schema)?);
            }
            Ok(Value::Array(objects))
        } else {
            Ok(first)
        }
    }

    /// Parse an object with schema: (val1,val2,...)
    fn parse_object_with_schema(&mut self, schema: &Schema) -> Result<Value> {
        self.expect(Token::LParen)?;

        let mut obj = IndexMap::new();
        let field_count = schema.len();

        for (i, field) in schema.fields.iter().enumerate() {
            if i > 0 {
                self.expect(Token::Comma)?;
            }

            let value = self.parse_field_value(field)?;
            obj.insert(field.name().to_string(), value);
        }

        // Check for extra values
        if self.check(&Token::Comma) {
            // Peek to see if there's actually a value or just trailing comma
            self.advance()?;
            if !self.check(&Token::RParen) {
                return Err(AsonError::FieldCountMismatch {
                    expected: field_count,
                    actual: field_count + 1,
                    position: self.position(),
                });
            }
        }

        self.expect(Token::RParen)?;
        Ok(Value::Object(obj))
    }

    /// Parse a value for a specific field type.
    fn parse_field_value(&mut self, field: &SchemaField) -> Result<Value> {
        match field {
            SchemaField::Simple(_) => self.parse_simple_value(),
            SchemaField::NestedObject { schema, .. } => self.parse_object_with_schema(schema),
            SchemaField::SimpleArray(_) => self.parse_simple_array(),
            SchemaField::ObjectArray { schema, .. } => self.parse_object_array(schema),
        }
    }

    /// Parse a simple value (not an object or typed array).
    fn parse_simple_value(&mut self) -> Result<Value> {
        match self.current().clone() {
            Token::Str(s) => {
                self.advance()?;
                Ok(Value::String(s))
            }
            Token::Integer(n) => {
                self.advance()?;
                Ok(Value::Integer(n))
            }
            Token::Float(n) => {
                self.advance()?;
                Ok(Value::Float(n))
            }
            Token::True => {
                self.advance()?;
                Ok(Value::Bool(true))
            }
            Token::False => {
                self.advance()?;
                Ok(Value::Bool(false))
            }
            Token::Comma | Token::RParen => {
                // Empty value = null
                Ok(Value::Null)
            }
            Token::LBracket => {
                // Inline array
                self.parse_array()
            }
            _ => Err(AsonError::UnexpectedChar {
                ch: '?',
                position: self.position(),
            }),
        }
    }

    /// Parse a simple array: [val1,val2,...]
    fn parse_simple_array(&mut self) -> Result<Value> {
        self.expect(Token::LBracket)?;

        let mut values = Vec::new();

        if self.check(&Token::RBracket) {
            self.advance()?;
            return Ok(Value::Array(values));
        }

        values.push(self.parse_simple_value()?);

        while self.check(&Token::Comma) {
            self.advance()?;
            if self.check(&Token::RBracket) {
                break;
            }
            values.push(self.parse_simple_value()?);
        }

        self.expect(Token::RBracket)?;
        Ok(Value::Array(values))
    }

    /// Parse an object array: [(obj1),(obj2),...]
    fn parse_object_array(&mut self, schema: &Schema) -> Result<Value> {
        self.expect(Token::LBracket)?;

        let mut objects = Vec::new();

        if self.check(&Token::RBracket) {
            self.advance()?;
            return Ok(Value::Array(objects));
        }

        objects.push(self.parse_object_with_schema(schema)?);

        while self.check(&Token::Comma) {
            self.advance()?;
            if self.check(&Token::RBracket) {
                break;
            }
            objects.push(self.parse_object_with_schema(schema)?);
        }

        self.expect(Token::RBracket)?;
        Ok(Value::Array(objects))
    }

    /// Parse a standalone array (no schema).
    fn parse_array(&mut self) -> Result<Value> {
        self.expect(Token::LBracket)?;

        let mut values = Vec::new();

        if self.check(&Token::RBracket) {
            self.advance()?;
            return Ok(Value::Array(values));
        }

        values.push(self.parse_value()?);

        while self.check(&Token::Comma) {
            self.advance()?;
            if self.check(&Token::RBracket) {
                break;
            }
            values.push(self.parse_value()?);
        }

        self.expect(Token::RBracket)?;
        Ok(Value::Array(values))
    }

    /// Parse any value (for standalone arrays).
    fn parse_value(&mut self) -> Result<Value> {
        match self.current().clone() {
            Token::Str(s) => {
                self.advance()?;
                Ok(Value::String(s))
            }
            Token::Integer(n) => {
                self.advance()?;
                Ok(Value::Integer(n))
            }
            Token::Float(n) => {
                self.advance()?;
                Ok(Value::Float(n))
            }
            Token::True => {
                self.advance()?;
                Ok(Value::Bool(true))
            }
            Token::False => {
                self.advance()?;
                Ok(Value::Bool(false))
            }
            Token::LBracket => self.parse_array(),
            Token::LParen => {
                // Anonymous tuple, parse as array
                self.advance()?;
                let mut values = Vec::new();
                if !self.check(&Token::RParen) {
                    values.push(self.parse_value()?);
                    while self.check(&Token::Comma) {
                        self.advance()?;
                        if self.check(&Token::RParen) {
                            break;
                        }
                        values.push(self.parse_value()?);
                    }
                }
                self.expect(Token::RParen)?;
                Ok(Value::Array(values))
            }
            _ => Ok(Value::Null),
        }
    }
}

/// Parse an ASON string into a Value.
pub fn parse(input: &str) -> Result<Value> {
    Parser::new(input)?.parse()
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_simple_object() {
        let result = parse("{name,age}:(Alice,30)").unwrap();
        assert!(result.is_object());
        assert_eq!(
            result.get("name"),
            Some(&Value::String("Alice".to_string()))
        );
        assert_eq!(result.get("age"), Some(&Value::Integer(30)));
    }

    #[test]
    fn test_multiple_objects() {
        let result = parse("{name,age}:(Alice,30),(Bob,25)").unwrap();
        assert!(result.is_array());
        let arr = result.as_array().unwrap();
        assert_eq!(arr.len(), 2);

        assert_eq!(
            arr[0].get("name"),
            Some(&Value::String("Alice".to_string()))
        );
        assert_eq!(arr[0].get("age"), Some(&Value::Integer(30)));
        assert_eq!(arr[1].get("name"), Some(&Value::String("Bob".to_string())));
        assert_eq!(arr[1].get("age"), Some(&Value::Integer(25)));
    }

    #[test]
    fn test_nested_object() {
        let result = parse("{name,addr{city,zip}}:(Alice,(NYC,10001))").unwrap();
        assert!(result.is_object());
        assert_eq!(
            result.get("name"),
            Some(&Value::String("Alice".to_string()))
        );

        let addr = result.get("addr").unwrap();
        assert!(addr.is_object());
        assert_eq!(addr.get("city"), Some(&Value::String("NYC".to_string())));
        assert_eq!(addr.get("zip"), Some(&Value::Integer(10001)));
    }

    #[test]
    fn test_simple_array_field() {
        let result = parse("{name,scores[]}:(Alice,[90,85,95])").unwrap();
        assert!(result.is_object());
        assert_eq!(
            result.get("name"),
            Some(&Value::String("Alice".to_string()))
        );

        let scores = result.get("scores").unwrap();
        assert!(scores.is_array());
        let arr = scores.as_array().unwrap();
        assert_eq!(arr.len(), 3);
        assert_eq!(arr[0], Value::Integer(90));
        assert_eq!(arr[1], Value::Integer(85));
        assert_eq!(arr[2], Value::Integer(95));
    }

    #[test]
    fn test_object_array_field() {
        let result = parse("{name,friends[{id,name}]}:(Alice,[(1,Bob),(2,Carol)])").unwrap();
        assert!(result.is_object());

        let friends = result.get("friends").unwrap();
        assert!(friends.is_array());
        let arr = friends.as_array().unwrap();
        assert_eq!(arr.len(), 2);

        assert_eq!(arr[0].get("id"), Some(&Value::Integer(1)));
        assert_eq!(arr[0].get("name"), Some(&Value::String("Bob".to_string())));
        assert_eq!(arr[1].get("id"), Some(&Value::Integer(2)));
        assert_eq!(
            arr[1].get("name"),
            Some(&Value::String("Carol".to_string()))
        );
    }

    #[test]
    fn test_null_values() {
        let result = parse("{name,age,city}:(Alice,,NYC)").unwrap();
        assert_eq!(
            result.get("name"),
            Some(&Value::String("Alice".to_string()))
        );
        assert_eq!(result.get("age"), Some(&Value::Null));
        assert_eq!(result.get("city"), Some(&Value::String("NYC".to_string())));
    }

    #[test]
    fn test_empty_string() {
        let result = parse(r#"{name,bio}:(Alice,"")"#).unwrap();
        assert_eq!(
            result.get("name"),
            Some(&Value::String("Alice".to_string()))
        );
        assert_eq!(result.get("bio"), Some(&Value::String("".to_string())));
    }

    #[test]
    fn test_quoted_strings() {
        let result = parse(r#"{name,city}:(Alice,"  New York  ")"#).unwrap();
        assert_eq!(
            result.get("name"),
            Some(&Value::String("Alice".to_string()))
        );
        assert_eq!(
            result.get("city"),
            Some(&Value::String("  New York  ".to_string()))
        );
    }

    #[test]
    fn test_booleans() {
        let result = parse("{name,active,verified}:(Alice,true,false)").unwrap();
        assert_eq!(result.get("active"), Some(&Value::Bool(true)));
        assert_eq!(result.get("verified"), Some(&Value::Bool(false)));
    }

    #[test]
    fn test_floats() {
        let result = parse("{name,score}:(Alice,98.5)").unwrap();
        assert_eq!(result.get("score"), Some(&Value::Float(98.5)));
    }

    #[test]
    fn test_standalone_array() {
        let result = parse("[1,2,3]").unwrap();
        assert!(result.is_array());
        let arr = result.as_array().unwrap();
        assert_eq!(arr.len(), 3);
        assert_eq!(arr[0], Value::Integer(1));
    }

    #[test]
    fn test_empty_array() {
        let result = parse("{name,tags[]}:(Alice,[])").unwrap();
        let tags = result.get("tags").unwrap();
        assert!(tags.is_array());
        assert!(tags.as_array().unwrap().is_empty());
    }

    #[test]
    fn test_comments() {
        let result = parse("/* header */ {name,age} /* schema */ : /* data */ (Alice,30)").unwrap();
        assert_eq!(
            result.get("name"),
            Some(&Value::String("Alice".to_string()))
        );
        assert_eq!(result.get("age"), Some(&Value::Integer(30)));
    }

    #[test]
    fn test_complex_nested() {
        let input = "{company,employees[{id,name,skills[]}],active}:(ACME,[(1,Alice,[rust,go]),(2,Bob,[python])],true)";
        let result = parse(input).unwrap();

        assert_eq!(
            result.get("company"),
            Some(&Value::String("ACME".to_string()))
        );
        assert_eq!(result.get("active"), Some(&Value::Bool(true)));

        let employees = result.get("employees").unwrap().as_array().unwrap();
        assert_eq!(employees.len(), 2);

        let alice = &employees[0];
        assert_eq!(alice.get("id"), Some(&Value::Integer(1)));
        assert_eq!(alice.get("name"), Some(&Value::String("Alice".to_string())));
        let skills = alice.get("skills").unwrap().as_array().unwrap();
        assert_eq!(skills.len(), 2);
        assert_eq!(skills[0], Value::String("rust".to_string()));
        assert_eq!(skills[1], Value::String("go".to_string()));
    }

    #[test]
    fn test_trailing_null() {
        let result = parse("{name,age}:(Alice,)").unwrap();
        assert_eq!(
            result.get("name"),
            Some(&Value::String("Alice".to_string()))
        );
        assert_eq!(result.get("age"), Some(&Value::Null));
    }

    #[test]
    fn test_field_count_mismatch() {
        let result = parse("{a,b}:(1,2,3)");
        assert!(result.is_err());
    }

    #[test]
    fn test_escape_in_data() {
        let result = parse(r#"{msg}:("say \"hello\"")"#).unwrap();
        assert_eq!(
            result.get("msg"),
            Some(&Value::String("say \"hello\"".to_string()))
        );
    }

    // ========================================================================
    // Edge case tests
    // ========================================================================

    #[test]
    fn test_unclosed_schema() {
        let result = parse("{name,age");
        assert!(result.is_err());
    }

    #[test]
    fn test_unclosed_data() {
        let result = parse("{name,age}:(Alice,30");
        assert!(result.is_err());
    }

    #[test]
    fn test_missing_colon() {
        let result = parse("{name,age}(Alice,30)");
        assert!(result.is_err());
    }

    #[test]
    fn test_deeply_nested() {
        // 3 levels of nesting
        let input = "{a{b{c}}}:(((inner)))";
        let result = parse(input).unwrap();
        let a = result.get("a").unwrap();
        let b = a.get("b").unwrap();
        let c = b.get("c").unwrap();
        assert_eq!(c, &Value::String("inner".to_string()));
    }

    #[test]
    fn test_unicode_strings() {
        let result = parse("{name,city}:(小明,北京)").unwrap();
        assert_eq!(result.get("name"), Some(&Value::String("小明".to_string())));
        assert_eq!(result.get("city"), Some(&Value::String("北京".to_string())));
    }

    #[test]
    fn test_unicode_with_emoji() {
        let result = parse("{msg}:(Hello 🌍!)").unwrap();
        assert_eq!(
            result.get("msg"),
            Some(&Value::String("Hello 🌍!".to_string()))
        );
    }

    #[test]
    fn test_negative_numbers() {
        let result = parse("{value}:(-42)").unwrap();
        assert_eq!(result.get("value"), Some(&Value::Integer(-42)));
    }

    #[test]
    fn test_negative_float() {
        let result = parse("{value}:(-3.14)").unwrap();
        assert_eq!(result.get("value"), Some(&Value::Float(-3.14)));
    }

    #[test]
    fn test_very_small_float() {
        let result = parse("{value}:(0.00001)").unwrap();
        assert_eq!(result.get("value"), Some(&Value::Float(0.00001)));
    }

    #[test]
    fn test_large_integer() {
        let result = parse("{value}:(9223372036854775807)").unwrap();
        assert_eq!(result.get("value"), Some(&Value::Integer(i64::MAX)));
    }

    #[test]
    fn test_all_nulls() {
        let result = parse("{a,b,c}:(,,)").unwrap();
        assert_eq!(result.get("a"), Some(&Value::Null));
        assert_eq!(result.get("b"), Some(&Value::Null));
        assert_eq!(result.get("c"), Some(&Value::Null));
    }

    #[test]
    fn test_nested_arrays() {
        let result = parse("[[1,2],[3,4],[5,6]]").unwrap();
        let arr = result.as_array().unwrap();
        assert_eq!(arr.len(), 3);

        let inner = arr[0].as_array().unwrap();
        assert_eq!(inner[0], Value::Integer(1));
        assert_eq!(inner[1], Value::Integer(2));
    }

    #[test]
    fn test_mixed_type_array() {
        let result = parse("[1,hello,true,2.5]").unwrap();
        let arr = result.as_array().unwrap();
        assert_eq!(arr[0], Value::Integer(1));
        assert_eq!(arr[1], Value::String("hello".to_string()));
        assert_eq!(arr[2], Value::Bool(true));
        assert_eq!(arr[3], Value::Float(2.5));
    }

    #[test]
    fn test_escaped_comma_in_unquoted() {
        let result = parse(r#"{msg}:(hello\, world)"#).unwrap();
        assert_eq!(
            result.get("msg"),
            Some(&Value::String("hello, world".to_string()))
        );
    }

    #[test]
    fn test_special_chars_quoted() {
        let result = parse(r#"{msg}:("a,b(c)d[e]f")"#).unwrap();
        assert_eq!(
            result.get("msg"),
            Some(&Value::String("a,b(c)d[e]f".to_string()))
        );
    }

    #[test]
    fn test_single_field_object() {
        let result = parse("{name}:(Alice)").unwrap();
        assert!(result.is_object());
        assert_eq!(
            result.get("name"),
            Some(&Value::String("Alice".to_string()))
        );
    }

    #[test]
    fn test_many_records() {
        let input = "{id}:(1),(2),(3),(4),(5)";
        let result = parse(input).unwrap();
        let arr = result.as_array().unwrap();
        assert_eq!(arr.len(), 5);
        assert_eq!(arr[4].get("id"), Some(&Value::Integer(5)));
    }

    #[test]
    fn test_empty_object_array_field() {
        let result = parse("{name,items[{id}]}:(Alice,[])").unwrap();
        let items = result.get("items").unwrap().as_array().unwrap();
        assert!(items.is_empty());
    }

    #[test]
    fn test_newline_in_quoted_string() {
        let result = parse(r#"{msg}:("line1\nline2")"#).unwrap();
        assert_eq!(
            result.get("msg"),
            Some(&Value::String("line1\nline2".to_string()))
        );
    }

    #[test]
    fn test_tab_in_quoted_string() {
        let result = parse(r#"{msg}:("col1\tcol2")"#).unwrap();
        assert_eq!(
            result.get("msg"),
            Some(&Value::String("col1\tcol2".to_string()))
        );
    }
}
