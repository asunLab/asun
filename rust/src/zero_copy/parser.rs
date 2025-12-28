//! Zero-copy parser for ASON.

use super::lexer::{Lexer, Token};
use super::value::Value;
use crate::error::{AsonError, Result};
use indexmap::IndexMap;
use std::borrow::Cow;

/// Parse an ASON string into a zero-copy Value.
///
/// The returned Value borrows string data from the input when possible,
/// avoiding memory allocation for strings without escape sequences.
pub fn parse(input: &str) -> Result<Value<'_>> {
    let mut parser = Parser::new(input);
    parser.parse()
}

struct Parser<'a> {
    lexer: Lexer<'a>,
    current: Token<'a>,
}

impl<'a> Parser<'a> {
    fn new(input: &'a str) -> Self {
        let mut lexer = Lexer::new(input);
        let current = lexer.next_token().map(|s| s.token).unwrap_or(Token::Eof);
        Self { lexer, current }
    }

    fn advance(&mut self) -> Result<()> {
        self.current = self.lexer.next_token()?.token;
        Ok(())
    }

    fn parse(&mut self) -> Result<Value<'a>> {
        match &self.current {
            Token::LBrace => self.parse_schema_with_data(),
            Token::LBracket => self.parse_array(),
            Token::LParen => self.parse_tuple(),
            _ => self.parse_primitive(),
        }
    }

    fn parse_primitive(&mut self) -> Result<Value<'a>> {
        let value = match std::mem::replace(&mut self.current, Token::Eof) {
            Token::True => Value::Bool(true),
            Token::False => Value::Bool(false),
            Token::Integer(n) => Value::Integer(n),
            Token::Float(n) => Value::Float(n),
            Token::Str(s) => Value::String(s),
            Token::Comma => Value::Null,
            Token::Eof => Value::Null,
            other => {
                return Err(AsonError::ParseError {
                    position: self.lexer.position(),
                    message: format!("unexpected token: {:?}", other),
                });
            }
        };
        self.advance()?;
        Ok(value)
    }

    fn parse_array(&mut self) -> Result<Value<'a>> {
        self.advance()?; // consume '['
        let mut items = Vec::new();

        while !matches!(self.current, Token::RBracket | Token::Eof) {
            if matches!(self.current, Token::Comma) {
                items.push(Value::Null);
                self.advance()?;
                continue;
            }

            items.push(self.parse()?);

            if matches!(self.current, Token::Comma) {
                self.advance()?;
            }
        }

        self.expect(Token::RBracket)?;
        Ok(Value::Array(items))
    }

    fn parse_tuple(&mut self) -> Result<Value<'a>> {
        self.advance()?; // consume '('
        let mut items = Vec::new();

        while !matches!(self.current, Token::RParen | Token::Eof) {
            if matches!(self.current, Token::Comma) {
                items.push(Value::Null);
                self.advance()?;
                continue;
            }

            items.push(self.parse()?);

            if matches!(self.current, Token::Comma) {
                self.advance()?;
            }
        }

        self.expect(Token::RParen)?;
        Ok(Value::Array(items))
    }

    fn parse_schema_with_data(&mut self) -> Result<Value<'a>> {
        // Parse schema
        let schema = self.parse_schema()?;

        // Expect ':'
        self.expect(Token::Colon)?;

        // Parse data records
        let mut records = Vec::new();
        loop {
            if !matches!(self.current, Token::LParen) {
                break;
            }
            let record = self.parse_record(&schema)?;
            records.push(record);

            // Check for comma between records
            if matches!(self.current, Token::Comma) {
                self.advance()?;
            } else {
                break;
            }
        }

        if records.len() == 1 {
            Ok(records.into_iter().next().unwrap())
        } else {
            Ok(Value::Array(records))
        }
    }

    fn expect(&mut self, expected: Token<'a>) -> Result<()> {
        if std::mem::discriminant(&self.current) == std::mem::discriminant(&expected) {
            self.advance()?;
            Ok(())
        } else {
            Err(AsonError::ExpectedToken {
                expected: format!("{:?}", expected),
                found: format!("{:?}", self.current),
                position: self.lexer.position(),
            })
        }
    }

    fn parse_schema(&mut self) -> Result<Vec<SchemaField<'a>>> {
        self.advance()?; // consume '{'
        let mut fields = Vec::new();

        while !matches!(self.current, Token::RBrace | Token::Eof) {
            let field = self.parse_schema_field()?;
            fields.push(field);

            if matches!(self.current, Token::Comma) {
                self.advance()?;
            }
        }

        self.expect(Token::RBrace)?;
        Ok(fields)
    }

    fn parse_schema_field(&mut self) -> Result<SchemaField<'a>> {
        let name = match std::mem::replace(&mut self.current, Token::Eof) {
            Token::Ident(s) => s,
            other => {
                return Err(AsonError::InvalidSchema {
                    position: self.lexer.position(),
                    message: format!("expected field name, found {:?}", other),
                });
            }
        };
        self.advance()?;

        // Check for nested object or array
        match &self.current {
            Token::LBrace => {
                let nested = self.parse_schema()?;
                Ok(SchemaField::NestedObject {
                    name,
                    schema: nested,
                })
            }
            Token::LBracket => {
                self.advance()?;
                if matches!(self.current, Token::LBrace) {
                    let nested = self.parse_schema()?;
                    self.expect(Token::RBracket)?;
                    Ok(SchemaField::ObjectArray {
                        name,
                        schema: nested,
                    })
                } else {
                    self.expect(Token::RBracket)?;
                    Ok(SchemaField::SimpleArray(name))
                }
            }
            _ => Ok(SchemaField::Simple(name)),
        }
    }

    fn parse_record(&mut self, schema: &[SchemaField<'a>]) -> Result<Value<'a>> {
        self.advance()?; // consume '('
        let mut obj = IndexMap::new();
        let mut field_idx = 0;

        while !matches!(self.current, Token::RParen | Token::Eof) {
            if field_idx >= schema.len() {
                return Err(AsonError::FieldCountMismatch {
                    expected: schema.len(),
                    actual: field_idx + 1,
                    position: self.lexer.position(),
                });
            }

            let field = &schema[field_idx];
            let value = self.parse_field_value(field)?;
            obj.insert(field.name().clone(), value);
            field_idx += 1;

            if matches!(self.current, Token::Comma) {
                self.advance()?;
            }
        }

        self.expect(Token::RParen)?;
        Ok(Value::Object(obj))
    }

    fn parse_field_value(&mut self, field: &SchemaField<'a>) -> Result<Value<'a>> {
        match field {
            SchemaField::Simple(_) => {
                if matches!(self.current, Token::Comma | Token::RParen) {
                    Ok(Value::Null)
                } else if matches!(self.current, Token::LParen) {
                    self.parse_tuple()
                } else if matches!(self.current, Token::LBracket) {
                    self.parse_array()
                } else {
                    self.parse_primitive()
                }
            }
            SchemaField::NestedObject { schema, .. } => {
                if matches!(self.current, Token::Comma | Token::RParen) {
                    Ok(Value::Null)
                } else {
                    self.parse_record(schema)
                }
            }
            SchemaField::SimpleArray(_) => {
                if matches!(self.current, Token::Comma | Token::RParen) {
                    Ok(Value::Null)
                } else {
                    self.parse_array()
                }
            }
            SchemaField::ObjectArray { schema, .. } => {
                if matches!(self.current, Token::Comma | Token::RParen) {
                    Ok(Value::Null)
                } else {
                    self.parse_object_array(schema)
                }
            }
        }
    }

    fn parse_object_array(&mut self, schema: &[SchemaField<'a>]) -> Result<Value<'a>> {
        self.advance()?; // consume '['
        let mut items = Vec::new();

        while !matches!(self.current, Token::RBracket | Token::Eof) {
            if matches!(self.current, Token::Comma) {
                items.push(Value::Null);
                self.advance()?;
                continue;
            }

            let record = self.parse_record(schema)?;
            items.push(record);

            if matches!(self.current, Token::Comma) {
                self.advance()?;
            }
        }

        self.expect(Token::RBracket)?;
        Ok(Value::Array(items))
    }
}

/// Schema field for zero-copy parsing.
#[derive(Debug, Clone)]
enum SchemaField<'a> {
    Simple(Cow<'a, str>),
    NestedObject {
        name: Cow<'a, str>,
        schema: Vec<SchemaField<'a>>,
    },
    SimpleArray(Cow<'a, str>),
    ObjectArray {
        name: Cow<'a, str>,
        schema: Vec<SchemaField<'a>>,
    },
}

impl<'a> SchemaField<'a> {
    fn name(&self) -> &Cow<'a, str> {
        match self {
            SchemaField::Simple(n) => n,
            SchemaField::NestedObject { name, .. } => name,
            SchemaField::SimpleArray(n) => n,
            SchemaField::ObjectArray { name, .. } => name,
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_zero_copy_simple() {
        let input = "{name,age}:(Alice,30)";
        let value = parse(input).unwrap();
        assert_eq!(value.get("name").unwrap().as_str(), Some("Alice"));
        assert_eq!(value.get("age").unwrap().as_i64(), Some(30));
    }

    #[test]
    fn test_zero_copy_borrows() {
        let input = "{name}:(hello)";
        let value = parse(input).unwrap();

        // The string should be borrowed, not allocated
        if let Value::Object(obj) = &value {
            if let Some(Value::String(cow)) = obj.get("name") {
                assert!(matches!(cow, Cow::Borrowed(_)));
            }
        }
    }

    #[test]
    fn test_zero_copy_escape_allocates() {
        let input = r#"{name}:("hello\nworld")"#;
        let value = parse(input).unwrap();

        // The string with escape should be allocated
        if let Value::Object(obj) = &value {
            if let Some(Value::String(cow)) = obj.get("name") {
                assert!(matches!(cow, Cow::Owned(_)));
                assert_eq!(cow.as_ref(), "hello\nworld");
            }
        }
    }

    #[test]
    fn test_zero_copy_array() {
        let input = "[1,2,3]";
        let value = parse(input).unwrap();
        let arr = value.as_array().unwrap();
        assert_eq!(arr.len(), 3);
    }

    #[test]
    fn test_zero_copy_nested() {
        let input = "{user{name,age}}:((Alice,30))";
        let value = parse(input).unwrap();
        let user = value.get("user").unwrap();
        assert_eq!(user.get("name").unwrap().as_str(), Some("Alice"));
    }

    #[test]
    fn test_into_owned() {
        let owned: crate::Value;
        {
            let input = "{name}:(Alice)".to_string();
            let value = parse(&input).unwrap();
            owned = value.into_owned();
        }
        // owned survives after input is dropped
        assert_eq!(owned.get("name").unwrap().as_str(), Some("Alice"));
    }
}
