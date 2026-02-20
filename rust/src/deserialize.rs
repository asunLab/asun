use crate::error::{Error, Result};
use serde::Deserialize;
use serde::de::{self, DeserializeSeed, MapAccess, SeqAccess, Visitor};

pub struct Deserializer<'de> {
    input: &'de [u8],
    pos: usize,
    /// Schema field names for current object context (positional mapping)
    schema_fields: Option<Vec<&'de str>>,
    /// Current field index within a tuple
    field_index: usize,
}

pub fn from_str<'a, T: Deserialize<'a>>(s: &'a str) -> Result<T> {
    let mut de = Deserializer {
        input: s.as_bytes(),
        pos: 0,
        schema_fields: None,
        field_index: 0,
    };
    de.skip_whitespace_and_comments();
    let value = T::deserialize(&mut de)?;
    de.skip_whitespace_and_comments();
    if de.pos < de.input.len() {
        if de.input[de.pos..].iter().all(|&b| b.is_ascii_whitespace()) {
            Ok(value)
        } else {
            Err(Error::TrailingCharacters)
        }
    } else {
        Ok(value)
    }
}

/// Deserialize a Vec<T> from ASON with schema: `{field1,field2,...}:(v1,v2),(v3,v4)`
pub fn from_str_vec<'a, T: Deserialize<'a>>(s: &'a str) -> Result<Vec<T>> {
    let mut de = Deserializer {
        input: s.as_bytes(),
        pos: 0,
        schema_fields: None,
        field_index: 0,
    };
    de.skip_whitespace_and_comments();

    // Parse schema
    if de.peek_byte()? != b'{' {
        return Err(Error::ExpectedOpenBrace);
    }
    let fields = de.parse_schema()?;

    de.skip_whitespace_and_comments();
    if de.next_byte()? != b':' {
        return Err(Error::ExpectedColon);
    }

    de.schema_fields = Some(fields);

    let mut results = Vec::new();
    loop {
        de.skip_whitespace_and_comments();
        if de.pos >= de.input.len() {
            break;
        }
        if de.peek_byte()? != b'(' {
            break;
        }
        de.field_index = 0;
        let val = T::deserialize(&mut de)?;
        results.push(val);
        de.skip_whitespace_and_comments();
        if de.pos < de.input.len() && de.peek_byte().ok() == Some(b',') {
            de.pos += 1;
            de.skip_whitespace_and_comments();
            if de.pos >= de.input.len() {
                break;
            }
            if de.peek_byte().ok() != Some(b'(') {
                break;
            }
        }
    }

    Ok(results)
}

impl<'de> Deserializer<'de> {
    #[inline(always)]
    fn peek_byte(&self) -> Result<u8> {
        if self.pos < self.input.len() {
            Ok(self.input[self.pos])
        } else {
            Err(Error::Eof)
        }
    }

    #[inline(always)]
    fn next_byte(&mut self) -> Result<u8> {
        if self.pos < self.input.len() {
            let b = self.input[self.pos];
            self.pos += 1;
            Ok(b)
        } else {
            Err(Error::Eof)
        }
    }

    #[inline(always)]
    fn skip_whitespace(&mut self) {
        while self.pos < self.input.len() {
            match self.input[self.pos] {
                b' ' | b'\t' | b'\n' | b'\r' => self.pos += 1,
                _ => break,
            }
        }
    }

    #[inline]
    fn skip_whitespace_and_comments(&mut self) {
        loop {
            self.skip_whitespace();
            if self.pos + 1 < self.input.len()
                && self.input[self.pos] == b'/'
                && self.input[self.pos + 1] == b'*'
            {
                self.pos += 2;
                while self.pos + 1 < self.input.len() {
                    if self.input[self.pos] == b'*' && self.input[self.pos + 1] == b'/' {
                        self.pos += 2;
                        break;
                    }
                    self.pos += 1;
                }
            } else {
                break;
            }
        }
    }

    fn parse_schema(&mut self) -> Result<Vec<&'de str>> {
        if self.next_byte()? != b'{' {
            return Err(Error::ExpectedOpenBrace);
        }
        let mut fields = Vec::new();
        loop {
            self.skip_whitespace();
            if self.peek_byte()? == b'}' {
                self.pos += 1;
                break;
            }
            if !fields.is_empty() {
                if self.next_byte()? != b',' {
                    return Err(Error::ExpectedComma);
                }
                self.skip_whitespace();
            }
            let start = self.pos;
            while self.pos < self.input.len() {
                match self.input[self.pos] {
                    b',' | b'}' | b':' | b' ' | b'\t' => break,
                    _ => self.pos += 1,
                }
            }
            let name = unsafe { core::str::from_utf8_unchecked(&self.input[start..self.pos]) };
            self.skip_whitespace();

            // Skip optional type hint
            if self.pos < self.input.len() && self.input[self.pos] == b':' {
                self.pos += 1;
                self.skip_whitespace();
                if self.pos < self.input.len() && self.input[self.pos] == b'{' {
                    self.skip_balanced(b'{', b'}')?;
                } else if self.pos < self.input.len() && self.input[self.pos] == b'[' {
                    self.skip_balanced(b'[', b']')?;
                } else if self.pos + 3 < self.input.len()
                    && &self.input[self.pos..self.pos + 3] == b"map"
                {
                    self.pos += 3;
                    if self.pos < self.input.len() && self.input[self.pos] == b'[' {
                        self.skip_balanced(b'[', b']')?;
                    }
                } else {
                    while self.pos < self.input.len() {
                        match self.input[self.pos] {
                            b',' | b'}' | b' ' | b'\t' => break,
                            _ => self.pos += 1,
                        }
                    }
                }
            }

            fields.push(name);
        }
        Ok(fields)
    }

    #[inline]
    fn skip_balanced(&mut self, open: u8, close: u8) -> Result<()> {
        let mut depth = 0u32;
        loop {
            if self.pos >= self.input.len() {
                return Err(Error::Eof);
            }
            let b = self.input[self.pos];
            self.pos += 1;
            if b == open {
                depth += 1;
            } else if b == close {
                if depth == 0 {
                    return Err(Error::Message("unbalanced brackets".into()));
                }
                depth -= 1;
                if depth == 0 {
                    return Ok(());
                }
            }
        }
    }

    /// Parse a plain (unquoted) string value, stopping at delimiters.
    /// Returns zerocopy borrowed str.
    #[inline]
    fn parse_plain_value(&mut self) -> Result<&'de str> {
        let start = self.pos;
        while self.pos < self.input.len() {
            match self.input[self.pos] {
                b',' | b')' | b']' => break,
                b'\\' => {
                    self.pos += 2;
                }
                _ => self.pos += 1,
            }
        }
        let raw = unsafe { core::str::from_utf8_unchecked(&self.input[start..self.pos]) };
        Ok(raw.trim())
    }

    /// Parse a quoted string. Zerocopy when no escapes; allocates only when escapes present.
    #[inline]
    fn parse_quoted_string_cow(&mut self) -> Result<CowStr<'de>> {
        // Skip opening quote
        self.pos += 1;
        let start = self.pos;

        // Fast scan: look for the closing quote without any escape
        let mut scan = self.pos;
        while scan < self.input.len() {
            match self.input[scan] {
                b'"' => {
                    // No escapes found — zerocopy path
                    let s = unsafe { core::str::from_utf8_unchecked(&self.input[start..scan]) };
                    self.pos = scan + 1;
                    return Ok(CowStr::Borrowed(s));
                }
                b'\\' => break, // escape found, fall through to slow path
                _ => scan += 1,
            }
        }

        // Slow path: build owned string with escapes
        // Copy what we scanned so far
        let mut result = String::with_capacity(scan - start + 16);
        if scan > start {
            let prefix = unsafe { core::str::from_utf8_unchecked(&self.input[start..scan]) };
            result.push_str(prefix);
        }
        self.pos = scan;

        loop {
            if self.pos >= self.input.len() {
                return Err(Error::UnclosedString);
            }
            let b = self.input[self.pos];
            if b == b'"' {
                self.pos += 1;
                return Ok(CowStr::Owned(result));
            }
            if b == b'\\' {
                self.pos += 1;
                if self.pos >= self.input.len() {
                    return Err(Error::UnclosedString);
                }
                let esc = self.input[self.pos];
                self.pos += 1;
                match esc {
                    b'"' => result.push('"'),
                    b'\\' => result.push('\\'),
                    b'n' => result.push('\n'),
                    b't' => result.push('\t'),
                    b',' => result.push(','),
                    b'(' => result.push('('),
                    b')' => result.push(')'),
                    b'[' => result.push('['),
                    b']' => result.push(']'),
                    b'u' => {
                        if self.pos + 4 > self.input.len() {
                            return Err(Error::InvalidUnicodeEscape);
                        }
                        let hex = unsafe {
                            core::str::from_utf8_unchecked(&self.input[self.pos..self.pos + 4])
                        };
                        let cp = u32::from_str_radix(hex, 16)
                            .map_err(|_| Error::InvalidUnicodeEscape)?;
                        let ch = char::from_u32(cp).ok_or(Error::InvalidUnicodeEscape)?;
                        result.push(ch);
                        self.pos += 4;
                    }
                    _ => return Err(Error::InvalidEscape(esc as char)),
                }
            } else {
                result.push(b as char);
                self.pos += 1;
            }
        }
    }

    /// Parse any value as a string.
    #[inline]
    fn parse_any_value_str(&mut self) -> Result<CowStr<'de>> {
        self.skip_whitespace_and_comments();
        if self.pos >= self.input.len() {
            return Ok(CowStr::Borrowed(""));
        }
        if self.input[self.pos] == b'"' {
            self.parse_quoted_string_cow()
        } else {
            let v = self.parse_plain_value()?;
            if v.contains('\\') {
                Ok(CowStr::Owned(unescape_plain(v)?))
            } else {
                Ok(CowStr::Borrowed(v))
            }
        }
    }

    /// Parse number directly without intermediate string::parse for integers.
    #[inline]
    fn parse_i64(&mut self) -> Result<i64> {
        let negative = self.pos < self.input.len() && self.input[self.pos] == b'-';
        if negative {
            self.pos += 1;
        }
        let mut val: u64 = 0;
        let mut digits = 0u32;
        while self.pos < self.input.len() && self.input[self.pos].is_ascii_digit() {
            val = val * 10 + (self.input[self.pos] - b'0') as u64;
            self.pos += 1;
            digits += 1;
        }
        if digits == 0 {
            return Err(Error::InvalidNumber);
        }
        if negative {
            Ok(-(val as i64))
        } else {
            Ok(val as i64)
        }
    }

    /// Parse u64 directly.
    #[inline]
    fn parse_u64(&mut self) -> Result<u64> {
        let mut val: u64 = 0;
        let mut digits = 0u32;
        while self.pos < self.input.len() && self.input[self.pos].is_ascii_digit() {
            val = val * 10 + (self.input[self.pos] - b'0') as u64;
            self.pos += 1;
            digits += 1;
        }
        if digits == 0 {
            return Err(Error::InvalidNumber);
        }
        Ok(val)
    }

    /// Parse number string for float parsing (still needs str::parse for floats).
    #[inline]
    fn parse_number_str(&mut self) -> Result<&'de str> {
        let start = self.pos;
        if self.pos < self.input.len() && self.input[self.pos] == b'-' {
            self.pos += 1;
        }
        while self.pos < self.input.len() && self.input[self.pos].is_ascii_digit() {
            self.pos += 1;
        }
        if self.pos < self.input.len() && self.input[self.pos] == b'.' {
            self.pos += 1;
            while self.pos < self.input.len() && self.input[self.pos].is_ascii_digit() {
                self.pos += 1;
            }
        }
        if self.pos == start || (self.pos == start + 1 && self.input[start] == b'-') {
            return Err(Error::InvalidNumber);
        }
        Ok(unsafe { core::str::from_utf8_unchecked(&self.input[start..self.pos]) })
    }

    /// Peek ahead to determine value type without consuming
    #[inline]
    fn peek_value_type(&self) -> ValueType {
        if self.pos >= self.input.len() {
            return ValueType::Null;
        }
        match self.input[self.pos] {
            b'"' => ValueType::String,
            b'(' => ValueType::Tuple,
            b'[' => ValueType::Array,
            b't' | b'f' => ValueType::Bool,
            b'-' => {
                if self.pos + 1 < self.input.len() && self.input[self.pos + 1].is_ascii_digit() {
                    ValueType::Number
                } else {
                    ValueType::String
                }
            }
            b'0'..=b'9' => ValueType::Number,
            b',' | b')' | b']' => ValueType::Null,
            _ => ValueType::String,
        }
    }

    #[inline(always)]
    fn at_value_end(&self) -> bool {
        if self.pos >= self.input.len() {
            return true;
        }
        matches!(self.input[self.pos], b',' | b')' | b']')
    }
}

/// Lightweight Cow-like enum to avoid std::borrow::Cow overhead
enum CowStr<'a> {
    Borrowed(&'a str),
    Owned(String),
}

impl<'a> CowStr<'a> {
    #[inline]
    fn as_str(&self) -> &str {
        match self {
            CowStr::Borrowed(s) => s,
            CowStr::Owned(s) => s,
        }
    }
}

#[derive(Debug, Clone, Copy, PartialEq)]
enum ValueType {
    Null,
    Bool,
    Number,
    String,
    Array,
    Tuple,
}

fn unescape_plain(s: &str) -> Result<String> {
    let mut result = String::with_capacity(s.len());
    let bytes = s.as_bytes();
    let mut i = 0;
    while i < bytes.len() {
        if bytes[i] == b'\\' {
            i += 1;
            if i >= bytes.len() {
                return Err(Error::Eof);
            }
            match bytes[i] {
                b',' => result.push(','),
                b'(' => result.push('('),
                b')' => result.push(')'),
                b'[' => result.push('['),
                b']' => result.push(']'),
                b'"' => result.push('"'),
                b'\\' => result.push('\\'),
                b'n' => result.push('\n'),
                b't' => result.push('\t'),
                b'u' => {
                    if i + 4 >= bytes.len() {
                        return Err(Error::InvalidUnicodeEscape);
                    }
                    let hex = unsafe { core::str::from_utf8_unchecked(&bytes[i + 1..i + 5]) };
                    let cp =
                        u32::from_str_radix(hex, 16).map_err(|_| Error::InvalidUnicodeEscape)?;
                    let ch = char::from_u32(cp).ok_or(Error::InvalidUnicodeEscape)?;
                    result.push(ch);
                    i += 4;
                }
                other => return Err(Error::InvalidEscape(other as char)),
            }
        } else {
            result.push(bytes[i] as char);
        }
        i += 1;
    }
    Ok(result)
}

impl<'de, 'a> de::Deserializer<'de> for &'a mut Deserializer<'de> {
    type Error = Error;

    #[inline]
    fn deserialize_any<V: Visitor<'de>>(self, visitor: V) -> Result<V::Value> {
        self.skip_whitespace_and_comments();
        match self.peek_value_type() {
            ValueType::Null => visitor.visit_none(),
            ValueType::Bool => self.deserialize_bool(visitor),
            ValueType::Number => {
                let s = self.parse_number_str()?;
                if s.contains('.') {
                    let f: f64 = s.parse().map_err(|_| Error::InvalidNumber)?;
                    visitor.visit_f64(f)
                } else {
                    let i: i64 = s.parse().map_err(|_| Error::InvalidNumber)?;
                    visitor.visit_i64(i)
                }
            }
            ValueType::String => {
                let cow = self.parse_any_value_str()?;
                match cow {
                    CowStr::Borrowed(s) => visitor.visit_borrowed_str(s),
                    CowStr::Owned(s) => visitor.visit_string(s),
                }
            }
            ValueType::Array => self.deserialize_seq(visitor),
            ValueType::Tuple => self.deserialize_map(visitor),
        }
    }

    #[inline]
    fn deserialize_bool<V: Visitor<'de>>(self, visitor: V) -> Result<V::Value> {
        self.skip_whitespace_and_comments();
        if self.pos + 4 <= self.input.len() && &self.input[self.pos..self.pos + 4] == b"true" {
            if self.pos + 4 >= self.input.len()
                || matches!(
                    self.input[self.pos + 4],
                    b',' | b')' | b']' | b' ' | b'\t' | b'\n' | b'\r'
                )
            {
                self.pos += 4;
                return visitor.visit_bool(true);
            }
        }
        if self.pos + 5 <= self.input.len() && &self.input[self.pos..self.pos + 5] == b"false" {
            if self.pos + 5 >= self.input.len()
                || matches!(
                    self.input[self.pos + 5],
                    b',' | b')' | b']' | b' ' | b'\t' | b'\n' | b'\r'
                )
            {
                self.pos += 5;
                return visitor.visit_bool(false);
            }
        }
        Err(Error::InvalidBool)
    }

    #[inline]
    fn deserialize_i8<V: Visitor<'de>>(self, visitor: V) -> Result<V::Value> {
        self.skip_whitespace_and_comments();
        let v = self.parse_i64()?;
        visitor.visit_i8(v as i8)
    }

    #[inline]
    fn deserialize_i16<V: Visitor<'de>>(self, visitor: V) -> Result<V::Value> {
        self.skip_whitespace_and_comments();
        let v = self.parse_i64()?;
        visitor.visit_i16(v as i16)
    }

    #[inline]
    fn deserialize_i32<V: Visitor<'de>>(self, visitor: V) -> Result<V::Value> {
        self.skip_whitespace_and_comments();
        let v = self.parse_i64()?;
        visitor.visit_i32(v as i32)
    }

    #[inline]
    fn deserialize_i64<V: Visitor<'de>>(self, visitor: V) -> Result<V::Value> {
        self.skip_whitespace_and_comments();
        let v = self.parse_i64()?;
        visitor.visit_i64(v)
    }

    #[inline]
    fn deserialize_u8<V: Visitor<'de>>(self, visitor: V) -> Result<V::Value> {
        self.skip_whitespace_and_comments();
        let v = self.parse_u64()?;
        visitor.visit_u8(v as u8)
    }

    #[inline]
    fn deserialize_u16<V: Visitor<'de>>(self, visitor: V) -> Result<V::Value> {
        self.skip_whitespace_and_comments();
        let v = self.parse_u64()?;
        visitor.visit_u16(v as u16)
    }

    #[inline]
    fn deserialize_u32<V: Visitor<'de>>(self, visitor: V) -> Result<V::Value> {
        self.skip_whitespace_and_comments();
        let v = self.parse_u64()?;
        visitor.visit_u32(v as u32)
    }

    #[inline]
    fn deserialize_u64<V: Visitor<'de>>(self, visitor: V) -> Result<V::Value> {
        self.skip_whitespace_and_comments();
        let v = self.parse_u64()?;
        visitor.visit_u64(v)
    }

    #[inline]
    fn deserialize_f32<V: Visitor<'de>>(self, visitor: V) -> Result<V::Value> {
        self.skip_whitespace_and_comments();
        let s = self.parse_number_str()?;
        let v: f32 = s.parse().map_err(|_| Error::InvalidNumber)?;
        visitor.visit_f32(v)
    }

    #[inline]
    fn deserialize_f64<V: Visitor<'de>>(self, visitor: V) -> Result<V::Value> {
        self.skip_whitespace_and_comments();
        let s = self.parse_number_str()?;
        let v: f64 = s.parse().map_err(|_| Error::InvalidNumber)?;
        visitor.visit_f64(v)
    }

    #[inline]
    fn deserialize_char<V: Visitor<'de>>(self, visitor: V) -> Result<V::Value> {
        self.skip_whitespace_and_comments();
        let cow = self.parse_any_value_str()?;
        let s = cow.as_str();
        let mut chars = s.chars();
        let c = chars.next().ok_or(Error::ExpectedValue)?;
        visitor.visit_char(c)
    }

    #[inline]
    fn deserialize_str<V: Visitor<'de>>(self, visitor: V) -> Result<V::Value> {
        self.skip_whitespace_and_comments();
        if self.pos < self.input.len() && self.input[self.pos] == b'"' {
            let cow = self.parse_quoted_string_cow()?;
            match cow {
                CowStr::Borrowed(s) => visitor.visit_borrowed_str(s),
                CowStr::Owned(s) => visitor.visit_string(s),
            }
        } else {
            let v = self.parse_plain_value()?;
            if v.contains('\\') {
                visitor.visit_string(unescape_plain(v)?)
            } else {
                visitor.visit_borrowed_str(v)
            }
        }
    }

    #[inline]
    fn deserialize_string<V: Visitor<'de>>(self, visitor: V) -> Result<V::Value> {
        self.deserialize_str(visitor)
    }

    #[inline]
    fn deserialize_bytes<V: Visitor<'de>>(self, visitor: V) -> Result<V::Value> {
        self.deserialize_seq(visitor)
    }

    #[inline]
    fn deserialize_byte_buf<V: Visitor<'de>>(self, visitor: V) -> Result<V::Value> {
        self.deserialize_seq(visitor)
    }

    #[inline]
    fn deserialize_option<V: Visitor<'de>>(self, visitor: V) -> Result<V::Value> {
        self.skip_whitespace_and_comments();
        if self.at_value_end() {
            visitor.visit_none()
        } else {
            visitor.visit_some(self)
        }
    }

    #[inline]
    fn deserialize_unit<V: Visitor<'de>>(self, visitor: V) -> Result<V::Value> {
        self.skip_whitespace_and_comments();
        if self.pos + 1 < self.input.len()
            && self.input[self.pos] == b'('
            && self.input[self.pos + 1] == b')'
        {
            self.pos += 2;
            visitor.visit_unit()
        } else if self.at_value_end() {
            visitor.visit_unit()
        } else {
            Err(Error::ExpectedValue)
        }
    }

    #[inline]
    fn deserialize_unit_struct<V: Visitor<'de>>(
        self,
        _name: &'static str,
        visitor: V,
    ) -> Result<V::Value> {
        self.deserialize_unit(visitor)
    }

    #[inline]
    fn deserialize_newtype_struct<V: Visitor<'de>>(
        self,
        _name: &'static str,
        visitor: V,
    ) -> Result<V::Value> {
        visitor.visit_newtype_struct(self)
    }

    #[inline]
    fn deserialize_seq<V: Visitor<'de>>(self, visitor: V) -> Result<V::Value> {
        self.skip_whitespace_and_comments();
        if self.next_byte()? != b'[' {
            return Err(Error::ExpectedOpenBracket);
        }
        let value = visitor.visit_seq(AsonSeqAccess {
            de: self,
            first: true,
        })?;
        self.skip_whitespace_and_comments();
        if self.pos < self.input.len() && self.input[self.pos] == b']' {
            self.pos += 1;
        }
        Ok(value)
    }

    #[inline]
    fn deserialize_tuple<V: Visitor<'de>>(self, _len: usize, visitor: V) -> Result<V::Value> {
        self.skip_whitespace_and_comments();
        if self.next_byte()? != b'(' {
            return Err(Error::ExpectedOpenParen);
        }
        let value = visitor.visit_seq(AsonTupleAccess {
            de: self,
            first: true,
        })?;
        self.skip_whitespace_and_comments();
        if self.pos < self.input.len() && self.input[self.pos] == b')' {
            self.pos += 1;
        }
        Ok(value)
    }

    #[inline]
    fn deserialize_tuple_struct<V: Visitor<'de>>(
        self,
        _name: &'static str,
        len: usize,
        visitor: V,
    ) -> Result<V::Value> {
        self.deserialize_tuple(len, visitor)
    }

    #[inline]
    fn deserialize_map<V: Visitor<'de>>(self, visitor: V) -> Result<V::Value> {
        self.skip_whitespace_and_comments();

        if self.schema_fields.is_some() {
            if self.peek_byte()? == b'[' {
                self.pos += 1;
                let value = visitor.visit_map(AsonMapAccess {
                    de: self,
                    first: true,
                })?;
                self.skip_whitespace_and_comments();
                if self.pos < self.input.len() && self.input[self.pos] == b']' {
                    self.pos += 1;
                }
                return Ok(value);
            }
            if self.next_byte()? != b'(' {
                return Err(Error::ExpectedOpenParen);
            }
            self.field_index = 0;
            let value = visitor.visit_map(AsonStructAccess {
                de: self,
                field_index: 0,
            })?;
            self.skip_whitespace_and_comments();
            if self.pos < self.input.len() && self.input[self.pos] == b')' {
                self.pos += 1;
            }
            Ok(value)
        } else if self.peek_byte()? == b'[' {
            self.pos += 1;
            let value = visitor.visit_map(AsonMapAccess {
                de: self,
                first: true,
            })?;
            self.skip_whitespace_and_comments();
            if self.pos < self.input.len() && self.input[self.pos] == b']' {
                self.pos += 1;
            }
            Ok(value)
        } else if self.peek_byte()? == b'{' {
            let fields = self.parse_schema()?;
            self.skip_whitespace_and_comments();
            if self.next_byte()? != b':' {
                return Err(Error::ExpectedColon);
            }
            self.skip_whitespace_and_comments();
            self.schema_fields = Some(fields);
            if self.next_byte()? != b'(' {
                return Err(Error::ExpectedOpenParen);
            }
            self.field_index = 0;
            let value = visitor.visit_map(AsonStructAccess {
                de: self,
                field_index: 0,
            })?;
            self.skip_whitespace_and_comments();
            if self.pos < self.input.len() && self.input[self.pos] == b')' {
                self.pos += 1;
            }
            self.schema_fields = None;
            Ok(value)
        } else {
            Err(Error::ExpectedOpenBrace)
        }
    }

    #[inline]
    fn deserialize_struct<V: Visitor<'de>>(
        self,
        _name: &'static str,
        fields: &'static [&'static str],
        visitor: V,
    ) -> Result<V::Value> {
        self.skip_whitespace_and_comments();

        if self.schema_fields.is_some() {
            if self.peek_byte()? == b'(' {
                self.pos += 1;
                self.field_index = 0;
                let parent_schema = self.schema_fields.take();
                let nested_fields: Vec<&'de str> = fields
                    .iter()
                    .map(|&f| unsafe { core::mem::transmute::<&str, &'de str>(f) })
                    .collect();
                self.schema_fields = Some(nested_fields);

                let value = visitor.visit_map(AsonStructAccess {
                    de: self,
                    field_index: 0,
                })?;
                self.skip_whitespace_and_comments();
                if self.pos < self.input.len() && self.input[self.pos] == b')' {
                    self.pos += 1;
                }
                self.schema_fields = parent_schema;
                return Ok(value);
            }
            let parent_schema = self.schema_fields.take();
            let nested_fields: Vec<&'de str> = fields
                .iter()
                .map(|&f| unsafe { core::mem::transmute::<&str, &'de str>(f) })
                .collect();
            self.schema_fields = Some(nested_fields);
            self.field_index = 0;
            let value = visitor.visit_map(AsonStructAccess {
                de: self,
                field_index: 0,
            })?;
            self.schema_fields = parent_schema;
            Ok(value)
        } else {
            if self.peek_byte()? == b'{' {
                let parsed_fields = self.parse_schema()?;
                self.skip_whitespace_and_comments();
                if self.next_byte()? != b':' {
                    return Err(Error::ExpectedColon);
                }
                self.skip_whitespace_and_comments();
                self.schema_fields = Some(parsed_fields);
                if self.next_byte()? != b'(' {
                    return Err(Error::ExpectedOpenParen);
                }
                self.field_index = 0;
                let value = visitor.visit_map(AsonStructAccess {
                    de: self,
                    field_index: 0,
                })?;
                self.skip_whitespace_and_comments();
                if self.pos < self.input.len() && self.input[self.pos] == b')' {
                    self.pos += 1;
                }
                self.schema_fields = None;
                Ok(value)
            } else if self.peek_byte()? == b'(' {
                self.pos += 1;
                let static_fields: Vec<&'de str> = fields
                    .iter()
                    .map(|&f| unsafe { core::mem::transmute::<&str, &'de str>(f) })
                    .collect();
                self.schema_fields = Some(static_fields);
                self.field_index = 0;
                let value = visitor.visit_map(AsonStructAccess {
                    de: self,
                    field_index: 0,
                })?;
                self.skip_whitespace_and_comments();
                if self.pos < self.input.len() && self.input[self.pos] == b')' {
                    self.pos += 1;
                }
                self.schema_fields = None;
                Ok(value)
            } else {
                Err(Error::ExpectedOpenBrace)
            }
        }
    }

    #[inline]
    fn deserialize_enum<V: Visitor<'de>>(
        self,
        _name: &'static str,
        _variants: &'static [&'static str],
        visitor: V,
    ) -> Result<V::Value> {
        self.skip_whitespace_and_comments();
        if self.peek_byte()? == b'(' {
            self.pos += 1;
            let value = visitor.visit_enum(AsonEnumAccess { de: self })?;
            self.skip_whitespace_and_comments();
            if self.pos < self.input.len() && self.input[self.pos] == b')' {
                self.pos += 1;
            }
            Ok(value)
        } else {
            visitor.visit_enum(AsonEnumAccess { de: self })
        }
    }

    #[inline]
    fn deserialize_identifier<V: Visitor<'de>>(self, visitor: V) -> Result<V::Value> {
        self.deserialize_str(visitor)
    }

    #[inline]
    fn deserialize_ignored_any<V: Visitor<'de>>(self, visitor: V) -> Result<V::Value> {
        self.deserialize_any(visitor)
    }
}

// --- Seq Access ---
struct AsonSeqAccess<'a, 'de: 'a> {
    de: &'a mut Deserializer<'de>,
    first: bool,
}

impl<'a, 'de> SeqAccess<'de> for AsonSeqAccess<'a, 'de> {
    type Error = Error;

    #[inline]
    fn next_element_seed<T: DeserializeSeed<'de>>(&mut self, seed: T) -> Result<Option<T::Value>> {
        self.de.skip_whitespace_and_comments();
        if self.de.pos >= self.de.input.len() {
            return Ok(None);
        }
        if self.de.input[self.de.pos] == b']' {
            return Ok(None);
        }
        if !self.first {
            if self.de.input[self.de.pos] == b',' {
                self.de.pos += 1;
                self.de.skip_whitespace_and_comments();
                if self.de.pos < self.de.input.len() && self.de.input[self.de.pos] == b']' {
                    return Ok(None);
                }
            } else {
                return Ok(None);
            }
        }
        self.first = false;
        seed.deserialize(&mut *self.de).map(Some)
    }
}

// --- Tuple Access ---
struct AsonTupleAccess<'a, 'de: 'a> {
    de: &'a mut Deserializer<'de>,
    first: bool,
}

impl<'a, 'de> SeqAccess<'de> for AsonTupleAccess<'a, 'de> {
    type Error = Error;

    #[inline]
    fn next_element_seed<T: DeserializeSeed<'de>>(&mut self, seed: T) -> Result<Option<T::Value>> {
        self.de.skip_whitespace_and_comments();
        if self.de.pos >= self.de.input.len() {
            return Ok(None);
        }
        if self.de.input[self.de.pos] == b')' {
            return Ok(None);
        }
        if !self.first {
            if self.de.input[self.de.pos] == b',' {
                self.de.pos += 1;
                self.de.skip_whitespace_and_comments();
                if self.de.pos < self.de.input.len() && self.de.input[self.de.pos] == b')' {
                    return Ok(None);
                }
            } else {
                return Ok(None);
            }
        }
        self.first = false;
        seed.deserialize(&mut *self.de).map(Some)
    }
}

// --- Struct (positional) Access ---
struct AsonStructAccess<'a, 'de: 'a> {
    de: &'a mut Deserializer<'de>,
    field_index: usize,
}

impl<'a, 'de> MapAccess<'de> for AsonStructAccess<'a, 'de> {
    type Error = Error;

    #[inline]
    fn next_key_seed<K: DeserializeSeed<'de>>(&mut self, seed: K) -> Result<Option<K::Value>> {
        self.de.skip_whitespace_and_comments();
        if self.de.pos >= self.de.input.len() {
            return Ok(None);
        }
        if self.de.input[self.de.pos] == b')' {
            return Ok(None);
        }

        let field_count = match &self.de.schema_fields {
            Some(f) => f.len(),
            None => return Ok(None),
        };

        if self.field_index >= field_count {
            return Ok(None);
        }

        if self.field_index > 0 {
            if self.de.pos < self.de.input.len() && self.de.input[self.de.pos] == b',' {
                self.de.pos += 1;
                self.de.skip_whitespace_and_comments();
                if self.de.pos < self.de.input.len() && self.de.input[self.de.pos] == b')' {
                    return Ok(None);
                }
            } else if self.de.input[self.de.pos] != b')' {
                return Ok(None);
            } else {
                return Ok(None);
            }
        }

        let field_name = self.de.schema_fields.as_ref().unwrap()[self.field_index];
        self.field_index += 1;
        self.de.field_index = self.field_index;

        seed.deserialize(FieldNameDeserializer { name: field_name })
            .map(Some)
    }

    #[inline]
    fn next_value_seed<V: DeserializeSeed<'de>>(&mut self, seed: V) -> Result<V::Value> {
        seed.deserialize(&mut *self.de)
    }
}

// --- Map Access for [(k,v)] ---
struct AsonMapAccess<'a, 'de: 'a> {
    de: &'a mut Deserializer<'de>,
    first: bool,
}

impl<'a, 'de> MapAccess<'de> for AsonMapAccess<'a, 'de> {
    type Error = Error;

    #[inline]
    fn next_key_seed<K: DeserializeSeed<'de>>(&mut self, seed: K) -> Result<Option<K::Value>> {
        self.de.skip_whitespace_and_comments();
        if self.de.pos >= self.de.input.len() {
            return Ok(None);
        }
        if self.de.input[self.de.pos] == b']' {
            return Ok(None);
        }
        if !self.first {
            if self.de.input[self.de.pos] == b',' {
                self.de.pos += 1;
                self.de.skip_whitespace_and_comments();
                if self.de.pos < self.de.input.len() && self.de.input[self.de.pos] == b']' {
                    return Ok(None);
                }
            } else {
                return Ok(None);
            }
        }
        self.first = false;

        self.de.skip_whitespace_and_comments();
        if self.de.next_byte()? != b'(' {
            return Err(Error::ExpectedOpenParen);
        }
        self.de.skip_whitespace_and_comments();
        seed.deserialize(&mut *self.de).map(Some)
    }

    #[inline]
    fn next_value_seed<V: DeserializeSeed<'de>>(&mut self, seed: V) -> Result<V::Value> {
        self.de.skip_whitespace_and_comments();
        if self.de.pos < self.de.input.len() && self.de.input[self.de.pos] == b',' {
            self.de.pos += 1;
        }
        self.de.skip_whitespace_and_comments();
        let value = seed.deserialize(&mut *self.de)?;
        self.de.skip_whitespace_and_comments();
        if self.de.pos < self.de.input.len() && self.de.input[self.de.pos] == b')' {
            self.de.pos += 1;
        }
        Ok(value)
    }
}

// --- Enum Access ---
struct AsonEnumAccess<'a, 'de: 'a> {
    de: &'a mut Deserializer<'de>,
}

impl<'a, 'de> de::EnumAccess<'de> for AsonEnumAccess<'a, 'de> {
    type Error = Error;
    type Variant = Self;

    #[inline]
    fn variant_seed<V: DeserializeSeed<'de>>(self, seed: V) -> Result<(V::Value, Self::Variant)> {
        self.de.skip_whitespace_and_comments();
        let val = seed.deserialize(&mut *self.de)?;
        Ok((val, self))
    }
}

impl<'a, 'de> de::VariantAccess<'de> for AsonEnumAccess<'a, 'de> {
    type Error = Error;

    #[inline]
    fn unit_variant(self) -> Result<()> {
        Ok(())
    }

    #[inline]
    fn newtype_variant_seed<T: DeserializeSeed<'de>>(self, seed: T) -> Result<T::Value> {
        self.de.skip_whitespace_and_comments();
        if self.de.pos < self.de.input.len() && self.de.input[self.de.pos] == b',' {
            self.de.pos += 1;
        }
        seed.deserialize(&mut *self.de)
    }

    #[inline]
    fn tuple_variant<V: Visitor<'de>>(self, _len: usize, visitor: V) -> Result<V::Value> {
        self.de.skip_whitespace_and_comments();
        if self.de.pos < self.de.input.len() && self.de.input[self.de.pos] == b',' {
            self.de.pos += 1;
        }
        let value = visitor.visit_seq(AsonTupleAccess {
            de: self.de,
            first: true,
        })?;
        Ok(value)
    }

    #[inline]
    fn struct_variant<V: Visitor<'de>>(
        self,
        fields: &'static [&'static str],
        visitor: V,
    ) -> Result<V::Value> {
        self.de.skip_whitespace_and_comments();
        if self.de.pos < self.de.input.len() && self.de.input[self.de.pos] == b',' {
            self.de.pos += 1;
        }
        let parent_schema = self.de.schema_fields.take();
        let parent_field_index = self.de.field_index;
        let static_fields: Vec<&'de str> = fields
            .iter()
            .map(|&f| unsafe { core::mem::transmute::<&str, &'de str>(f) })
            .collect();
        self.de.schema_fields = Some(static_fields);
        self.de.field_index = 0;
        let value = visitor.visit_map(AsonStructAccess {
            de: self.de,
            field_index: 0,
        })?;
        self.de.schema_fields = parent_schema;
        self.de.field_index = parent_field_index;
        Ok(value)
    }
}

// --- FieldName Deserializer (zerocopy: just returns &str) ---
struct FieldNameDeserializer<'de> {
    name: &'de str,
}

impl<'de> de::Deserializer<'de> for FieldNameDeserializer<'de> {
    type Error = Error;

    #[inline]
    fn deserialize_any<V: Visitor<'de>>(self, visitor: V) -> Result<V::Value> {
        visitor.visit_borrowed_str(self.name)
    }

    #[inline]
    fn deserialize_identifier<V: Visitor<'de>>(self, visitor: V) -> Result<V::Value> {
        visitor.visit_borrowed_str(self.name)
    }

    #[inline]
    fn deserialize_str<V: Visitor<'de>>(self, visitor: V) -> Result<V::Value> {
        visitor.visit_borrowed_str(self.name)
    }

    #[inline]
    fn deserialize_string<V: Visitor<'de>>(self, visitor: V) -> Result<V::Value> {
        visitor.visit_borrowed_str(self.name)
    }

    serde::forward_to_deserialize_any! {
        bool i8 i16 i32 i64 u8 u16 u32 u64 f32 f64 char bytes byte_buf
        option unit unit_struct newtype_struct seq tuple tuple_struct
        map struct enum ignored_any
    }
}
