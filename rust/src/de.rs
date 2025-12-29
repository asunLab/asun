//! ASON Deserializer implementation.

use crate::error::{Error, Result};
use serde::Deserialize;
use serde::de::{self, DeserializeSeed, MapAccess, SeqAccess, Visitor};

/// Deserialize an ASON string into a type.
pub fn from_str<'a, T: Deserialize<'a>>(s: &'a str) -> Result<T> {
    let mut de = Deserializer::new(s);
    T::deserialize(&mut de)
}

/// Schema field info.
#[derive(Debug, Clone)]
struct SchemaField {
    name: String,
    #[allow(dead_code)]
    is_array: bool,
    children: Vec<SchemaField>,
}

/// ASON Deserializer.
pub struct Deserializer<'de> {
    input: &'de str,
    pos: usize,
    schema: Vec<SchemaField>,
    #[allow(dead_code)]
    schema_idx: usize,
}

impl<'de> Deserializer<'de> {
    pub fn new(input: &'de str) -> Self {
        Deserializer {
            input,
            pos: 0,
            schema: Vec::new(),
            schema_idx: 0,
        }
    }

    fn peek(&self) -> Option<char> {
        self.input[self.pos..].chars().next()
    }

    #[allow(dead_code)]
    fn peek_n(&self, n: usize) -> Option<char> {
        self.input[self.pos..].chars().nth(n)
    }

    fn advance(&mut self) -> Option<char> {
        let c = self.peek()?;
        self.pos += c.len_utf8();
        Some(c)
    }

    fn skip_ws(&mut self) {
        while let Some(c) = self.peek() {
            if c.is_whitespace() {
                self.advance();
            } else {
                break;
            }
        }
    }

    fn expect(&mut self, expected: char) -> Result<()> {
        self.skip_ws();
        match self.advance() {
            Some(c) if c == expected => Ok(()),
            Some(c) => Err(Error::new(format!("expected '{}', got '{}'", expected, c))),
            None => Err(Error::new(format!("expected '{}', got EOF", expected))),
        }
    }

    fn parse_schema(&mut self) -> Result<Vec<SchemaField>> {
        self.expect('{')?;
        let mut fields = Vec::new();
        loop {
            self.skip_ws();
            if self.peek() == Some('}') {
                self.advance();
                break;
            }
            if !fields.is_empty() {
                self.expect(',')?;
            }
            fields.push(self.parse_schema_field()?);
        }
        Ok(fields)
    }

    fn parse_schema_field(&mut self) -> Result<SchemaField> {
        self.skip_ws();
        let name = self.parse_ident()?;
        let mut is_array = false;
        let mut children = Vec::new();

        self.skip_ws();
        // Check for []
        if self.peek() == Some('[') {
            self.advance();
            self.expect(']')?;
            is_array = true;
        }
        // Check for nested {}
        self.skip_ws();
        if self.peek() == Some('{') {
            children = self.parse_schema()?;
        }

        Ok(SchemaField {
            name,
            is_array,
            children,
        })
    }

    fn parse_ident(&mut self) -> Result<String> {
        self.skip_ws();
        let start = self.pos;
        while let Some(c) = self.peek() {
            if c.is_alphanumeric() || c == '_' || c == '-' {
                self.advance();
            } else {
                break;
            }
        }
        if self.pos == start {
            return Err(Error::new("expected identifier"));
        }
        Ok(self.input[start..self.pos].to_string())
    }

    fn parse_string(&mut self) -> Result<&'de str> {
        self.skip_ws();
        if self.peek() == Some('"') {
            self.parse_quoted_string()
        } else {
            self.parse_unquoted_string()
        }
    }

    fn parse_quoted_string(&mut self) -> Result<&'de str> {
        self.expect('"')?;
        let start = self.pos;
        while let Some(c) = self.peek() {
            if c == '"' {
                break;
            } else if c == '\\' {
                self.advance();
                self.advance();
            } else {
                self.advance();
            }
        }
        let end = self.pos;
        self.expect('"')?;
        Ok(&self.input[start..end])
    }

    fn parse_unquoted_string(&mut self) -> Result<&'de str> {
        self.skip_ws();
        let start = self.pos;
        while let Some(c) = self.peek() {
            if c == '('
                || c == ')'
                || c == '['
                || c == ']'
                || c == '{'
                || c == '}'
                || c == ','
                || c == ':'
                || c.is_whitespace()
            {
                break;
            }
            self.advance();
        }
        Ok(&self.input[start..self.pos])
    }

    fn parse_number<T: std::str::FromStr>(&mut self) -> Result<T> {
        self.skip_ws();
        let start = self.pos;
        if self.peek() == Some('-') || self.peek() == Some('+') {
            self.advance();
        }
        while let Some(c) = self.peek() {
            if c.is_ascii_digit() || c == '.' || c == 'e' || c == 'E' || c == '+' || c == '-' {
                // Handle sign only after e/E
                if (c == '+' || c == '-') && self.pos > start {
                    let prev = self.input[..self.pos].chars().last();
                    if prev != Some('e') && prev != Some('E') {
                        break;
                    }
                }
                self.advance();
            } else {
                break;
            }
        }
        self.input[start..self.pos]
            .parse()
            .map_err(|_| Error::new(format!("invalid number: {}", &self.input[start..self.pos])))
    }
}

impl<'de, 'a> de::Deserializer<'de> for &'a mut Deserializer<'de> {
    type Error = Error;

    fn deserialize_any<V: Visitor<'de>>(self, visitor: V) -> Result<V::Value> {
        self.skip_ws();
        match self.peek() {
            Some('{') => {
                // Could be schema or map
                self.schema = self.parse_schema()?;
                self.expect(':')?;
                self.skip_ws();
                if self.peek() == Some('[') {
                    // Array of objects
                    self.deserialize_seq(visitor)
                } else {
                    self.deserialize_map(visitor)
                }
            }
            Some('[') => self.deserialize_seq(visitor),
            Some('(') => self.deserialize_map(visitor),
            Some('"') => self.deserialize_str(visitor),
            Some('t') | Some('f') => self.deserialize_bool(visitor),
            Some('n') => {
                self.advance();
                self.advance();
                self.advance();
                self.advance(); // null
                visitor.visit_unit()
            }
            Some(c) if c == '-' || c == '+' || c.is_ascii_digit() => self.deserialize_i64(visitor),
            _ => self.deserialize_str(visitor),
        }
    }

    fn deserialize_bool<V: Visitor<'de>>(self, visitor: V) -> Result<V::Value> {
        self.skip_ws();
        if self.input[self.pos..].starts_with("true") {
            self.pos += 4;
            visitor.visit_bool(true)
        } else if self.input[self.pos..].starts_with("false") {
            self.pos += 5;
            visitor.visit_bool(false)
        } else {
            Err(Error::new("expected bool"))
        }
    }

    fn deserialize_i8<V: Visitor<'de>>(self, visitor: V) -> Result<V::Value> {
        visitor.visit_i8(self.parse_number()?)
    }
    fn deserialize_i16<V: Visitor<'de>>(self, visitor: V) -> Result<V::Value> {
        visitor.visit_i16(self.parse_number()?)
    }
    fn deserialize_i32<V: Visitor<'de>>(self, visitor: V) -> Result<V::Value> {
        visitor.visit_i32(self.parse_number()?)
    }
    fn deserialize_i64<V: Visitor<'de>>(self, visitor: V) -> Result<V::Value> {
        visitor.visit_i64(self.parse_number()?)
    }
    fn deserialize_u8<V: Visitor<'de>>(self, visitor: V) -> Result<V::Value> {
        visitor.visit_u8(self.parse_number()?)
    }
    fn deserialize_u16<V: Visitor<'de>>(self, visitor: V) -> Result<V::Value> {
        visitor.visit_u16(self.parse_number()?)
    }
    fn deserialize_u32<V: Visitor<'de>>(self, visitor: V) -> Result<V::Value> {
        visitor.visit_u32(self.parse_number()?)
    }
    fn deserialize_u64<V: Visitor<'de>>(self, visitor: V) -> Result<V::Value> {
        visitor.visit_u64(self.parse_number()?)
    }
    fn deserialize_f32<V: Visitor<'de>>(self, visitor: V) -> Result<V::Value> {
        visitor.visit_f32(self.parse_number()?)
    }
    fn deserialize_f64<V: Visitor<'de>>(self, visitor: V) -> Result<V::Value> {
        visitor.visit_f64(self.parse_number()?)
    }

    fn deserialize_char<V: Visitor<'de>>(self, visitor: V) -> Result<V::Value> {
        let s = self.parse_string()?;
        let mut chars = s.chars();
        match chars.next() {
            Some(c) if chars.next().is_none() => visitor.visit_char(c),
            _ => Err(Error::new("expected single char")),
        }
    }

    fn deserialize_str<V: Visitor<'de>>(self, visitor: V) -> Result<V::Value> {
        visitor.visit_borrowed_str(self.parse_string()?)
    }

    fn deserialize_string<V: Visitor<'de>>(self, visitor: V) -> Result<V::Value> {
        self.deserialize_str(visitor)
    }

    fn deserialize_bytes<V: Visitor<'de>>(self, _visitor: V) -> Result<V::Value> {
        Err(Error::new("bytes not supported"))
    }

    fn deserialize_byte_buf<V: Visitor<'de>>(self, _visitor: V) -> Result<V::Value> {
        Err(Error::new("byte_buf not supported"))
    }

    fn deserialize_option<V: Visitor<'de>>(self, visitor: V) -> Result<V::Value> {
        self.skip_ws();
        if self.input[self.pos..].starts_with("null") {
            self.pos += 4;
            visitor.visit_none()
        } else {
            visitor.visit_some(self)
        }
    }

    fn deserialize_unit<V: Visitor<'de>>(self, visitor: V) -> Result<V::Value> {
        self.skip_ws();
        if self.input[self.pos..].starts_with("null") {
            self.pos += 4;
            visitor.visit_unit()
        } else {
            Err(Error::new("expected null"))
        }
    }

    fn deserialize_unit_struct<V: Visitor<'de>>(
        self,
        _: &'static str,
        visitor: V,
    ) -> Result<V::Value> {
        self.deserialize_unit(visitor)
    }

    fn deserialize_newtype_struct<V: Visitor<'de>>(
        self,
        _: &'static str,
        visitor: V,
    ) -> Result<V::Value> {
        visitor.visit_newtype_struct(self)
    }

    fn deserialize_seq<V: Visitor<'de>>(self, visitor: V) -> Result<V::Value> {
        self.skip_ws();
        if self.peek() == Some('[') {
            self.advance();
            let value = visitor.visit_seq(SeqDeserializer::new(self))?;
            self.expect(']')?;
            Ok(value)
        } else if self.peek() == Some('{') {
            // Array of structs: {schema}:(v1),(v2),...
            self.schema = self.parse_schema()?;
            self.expect(':')?;
            let value = visitor.visit_seq(TupleSeqDeserializer::new(self))?;
            Ok(value)
        } else if self.peek() == Some('(') {
            // Single object as array element
            let value = visitor.visit_seq(TupleSeqDeserializer::new(self))?;
            Ok(value)
        } else {
            Err(Error::new("expected '[' or '(' or '{'"))
        }
    }

    fn deserialize_tuple<V: Visitor<'de>>(self, _len: usize, visitor: V) -> Result<V::Value> {
        self.deserialize_seq(visitor)
    }

    fn deserialize_tuple_struct<V: Visitor<'de>>(
        self,
        _: &'static str,
        _len: usize,
        visitor: V,
    ) -> Result<V::Value> {
        self.deserialize_seq(visitor)
    }

    fn deserialize_map<V: Visitor<'de>>(self, visitor: V) -> Result<V::Value> {
        self.skip_ws();
        self.expect('(')?;
        let schema = std::mem::take(&mut self.schema);
        let value = visitor.visit_map(StructDeserializer::new(self, schema))?;
        self.expect(')')?;
        Ok(value)
    }

    fn deserialize_struct<V: Visitor<'de>>(
        self,
        _: &'static str,
        fields: &'static [&'static str],
        visitor: V,
    ) -> Result<V::Value> {
        self.skip_ws();
        // Check if there's a schema
        if self.peek() == Some('{') {
            self.schema = self.parse_schema()?;
            self.expect(':')?;
        } else if self.schema.is_empty() {
            // Auto-generate schema from struct fields
            self.schema = fields
                .iter()
                .map(|f| SchemaField {
                    name: f.to_string(),
                    is_array: false,
                    children: Vec::new(),
                })
                .collect();
        }
        self.deserialize_map(visitor)
    }

    fn deserialize_enum<V: Visitor<'de>>(
        self,
        _: &'static str,
        _: &'static [&'static str],
        visitor: V,
    ) -> Result<V::Value> {
        visitor.visit_enum(EnumDeserializer::new(self))
    }

    fn deserialize_identifier<V: Visitor<'de>>(self, visitor: V) -> Result<V::Value> {
        self.deserialize_str(visitor)
    }

    fn deserialize_ignored_any<V: Visitor<'de>>(self, visitor: V) -> Result<V::Value> {
        self.deserialize_any(visitor)
    }
}

// Sequence deserializer for [...]
struct SeqDeserializer<'a, 'de: 'a> {
    de: &'a mut Deserializer<'de>,
    first: bool,
}

impl<'a, 'de> SeqDeserializer<'a, 'de> {
    fn new(de: &'a mut Deserializer<'de>) -> Self {
        SeqDeserializer { de, first: true }
    }
}

impl<'de, 'a> SeqAccess<'de> for SeqDeserializer<'a, 'de> {
    type Error = Error;

    fn next_element_seed<T: DeserializeSeed<'de>>(&mut self, seed: T) -> Result<Option<T::Value>> {
        self.de.skip_ws();
        if self.de.peek() == Some(']') {
            return Ok(None);
        }
        if !self.first {
            self.de.expect(',')?;
        }
        self.first = false;
        seed.deserialize(&mut *self.de).map(Some)
    }
}

// Tuple sequence for (v1),(v2),...
struct TupleSeqDeserializer<'a, 'de: 'a> {
    de: &'a mut Deserializer<'de>,
    first: bool,
}

impl<'a, 'de> TupleSeqDeserializer<'a, 'de> {
    fn new(de: &'a mut Deserializer<'de>) -> Self {
        TupleSeqDeserializer { de, first: true }
    }
}

impl<'de, 'a> SeqAccess<'de> for TupleSeqDeserializer<'a, 'de> {
    type Error = Error;

    fn next_element_seed<T: DeserializeSeed<'de>>(&mut self, seed: T) -> Result<Option<T::Value>> {
        self.de.skip_ws();

        if !self.first {
            // After first element, check for comma before next tuple
            if self.de.peek() == Some(',') {
                self.de.advance();
                self.de.skip_ws();
            } else {
                return Ok(None);
            }
        }

        if self.de.peek() != Some('(') {
            return Ok(None);
        }

        self.first = false;
        seed.deserialize(&mut *self.de).map(Some)
    }
}

// Struct deserializer using schema
struct StructDeserializer<'a, 'de: 'a> {
    de: &'a mut Deserializer<'de>,
    schema: Vec<SchemaField>,
    idx: usize,
    first: bool,
}

impl<'a, 'de> StructDeserializer<'a, 'de> {
    fn new(de: &'a mut Deserializer<'de>, schema: Vec<SchemaField>) -> Self {
        StructDeserializer {
            de,
            schema,
            idx: 0,
            first: true,
        }
    }
}

impl<'de, 'a> MapAccess<'de> for StructDeserializer<'a, 'de> {
    type Error = Error;

    fn next_key_seed<K: DeserializeSeed<'de>>(&mut self, seed: K) -> Result<Option<K::Value>> {
        if self.idx >= self.schema.len() {
            return Ok(None);
        }
        let field = &self.schema[self.idx];
        // Use a string deserializer for the key
        seed.deserialize(StrDeserializer { s: &field.name })
            .map(Some)
    }

    fn next_value_seed<V: DeserializeSeed<'de>>(&mut self, seed: V) -> Result<V::Value> {
        if !self.first {
            self.de.expect(',')?;
        }
        self.first = false;

        let field = &self.schema[self.idx];
        self.idx += 1;

        // Set child schema for nested structs
        if !field.children.is_empty() {
            self.de.schema = field.children.clone();
        }

        seed.deserialize(&mut *self.de)
    }
}

// Simple string deserializer for keys
struct StrDeserializer<'a> {
    s: &'a str,
}

impl<'de, 'a> de::Deserializer<'de> for StrDeserializer<'a> {
    type Error = Error;

    fn deserialize_any<V: Visitor<'de>>(self, visitor: V) -> Result<V::Value> {
        visitor.visit_str(self.s)
    }

    serde::forward_to_deserialize_any! {
        bool i8 i16 i32 i64 u8 u16 u32 u64 f32 f64 char str string bytes
        byte_buf option unit unit_struct newtype_struct seq tuple
        tuple_struct map struct enum identifier ignored_any
    }
}

// Enum deserializer
struct EnumDeserializer<'a, 'de: 'a> {
    de: &'a mut Deserializer<'de>,
}

impl<'a, 'de> EnumDeserializer<'a, 'de> {
    fn new(de: &'a mut Deserializer<'de>) -> Self {
        EnumDeserializer { de }
    }
}

impl<'de, 'a> de::EnumAccess<'de> for EnumDeserializer<'a, 'de> {
    type Error = Error;
    type Variant = Self;

    fn variant_seed<V: DeserializeSeed<'de>>(self, seed: V) -> Result<(V::Value, Self::Variant)> {
        let val = seed.deserialize(&mut *self.de)?;
        Ok((val, self))
    }
}

impl<'de, 'a> de::VariantAccess<'de> for EnumDeserializer<'a, 'de> {
    type Error = Error;

    fn unit_variant(self) -> Result<()> {
        Ok(())
    }

    fn newtype_variant_seed<T: DeserializeSeed<'de>>(self, seed: T) -> Result<T::Value> {
        self.de.expect(':')?;
        seed.deserialize(&mut *self.de)
    }

    fn tuple_variant<V: Visitor<'de>>(self, _len: usize, visitor: V) -> Result<V::Value> {
        self.de.expect(':')?;
        de::Deserializer::deserialize_seq(&mut *self.de, visitor)
    }

    fn struct_variant<V: Visitor<'de>>(
        self,
        fields: &'static [&'static str],
        visitor: V,
    ) -> Result<V::Value> {
        self.de.expect(':')?;
        de::Deserializer::deserialize_struct(&mut *self.de, "", fields, visitor)
    }
}
