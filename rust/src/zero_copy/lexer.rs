//! Zero-copy lexer for ASON.

use crate::error::{AsonError, Result};
use std::borrow::Cow;

/// Token with borrowed strings.
#[derive(Debug, Clone, PartialEq)]
pub enum Token<'a> {
    LBrace,
    RBrace,
    LParen,
    RParen,
    LBracket,
    RBracket,
    Comma,
    Colon,
    /// Identifier (field name) - borrowed when possible.
    Ident(Cow<'a, str>),
    /// String value - borrowed when no escapes.
    Str(Cow<'a, str>),
    Integer(i64),
    Float(f64),
    True,
    False,
    Eof,
}

/// Spanned token with position info.
#[derive(Debug, Clone)]
#[allow(dead_code)]
pub struct Spanned<'a> {
    pub token: Token<'a>,
    pub start: usize,
    pub end: usize,
}

impl<'a> Spanned<'a> {
    pub fn new(token: Token<'a>, start: usize, end: usize) -> Self {
        Self { token, start, end }
    }
}

/// Zero-copy lexer.
pub struct Lexer<'a> {
    input: &'a str,
    bytes: &'a [u8],
    pos: usize,
    in_schema: bool,
}

impl<'a> Lexer<'a> {
    pub fn new(input: &'a str) -> Self {
        Self {
            input,
            bytes: input.as_bytes(),
            pos: 0,
            in_schema: false,
        }
    }

    pub fn position(&self) -> usize {
        self.pos
    }

    fn peek(&self) -> Option<u8> {
        self.bytes.get(self.pos).copied()
    }

    fn advance(&mut self) -> Option<u8> {
        let b = self.bytes.get(self.pos).copied();
        if b.is_some() {
            self.pos += 1;
        }
        b
    }

    fn skip_whitespace_and_comments(&mut self) -> Result<()> {
        loop {
            // Skip whitespace
            while let Some(b) = self.peek() {
                if b == b' ' || b == b'\t' || b == b'\n' || b == b'\r' {
                    self.advance();
                } else {
                    break;
                }
            }

            // Check for comments
            if self.peek() == Some(b'/') {
                let next_pos = self.pos + 1;
                if next_pos < self.bytes.len() {
                    match self.bytes[next_pos] {
                        b'/' => {
                            // Line comment
                            self.pos += 2;
                            while let Some(b) = self.peek() {
                                self.advance();
                                if b == b'\n' {
                                    break;
                                }
                            }
                            continue;
                        }
                        b'*' => {
                            // Block comment
                            let start = self.pos;
                            self.pos += 2;
                            loop {
                                match self.advance() {
                                    Some(b'*') if self.peek() == Some(b'/') => {
                                        self.advance();
                                        break;
                                    }
                                    Some(_) => continue,
                                    None => {
                                        return Err(AsonError::UnclosedComment { position: start });
                                    }
                                }
                            }
                            continue;
                        }
                        _ => {}
                    }
                }
            }
            break;
        }
        Ok(())
    }

    pub fn next_token(&mut self) -> Result<Spanned<'a>> {
        self.skip_whitespace_and_comments()?;

        let start = self.pos;

        let Some(b) = self.advance() else {
            return Ok(Spanned::new(Token::Eof, start, start));
        };

        let token = match b {
            b'{' => {
                self.in_schema = true;
                Token::LBrace
            }
            b'}' => {
                self.in_schema = false;
                Token::RBrace
            }
            b'(' => Token::LParen,
            b')' => Token::RParen,
            b'[' => Token::LBracket,
            b']' => Token::RBracket,
            b',' => Token::Comma,
            b':' => Token::Colon,
            b'"' => self.parse_quoted_string(start)?,
            _ => {
                if self.in_schema {
                    self.parse_identifier(start, b)?
                } else {
                    self.parse_value(start, b)?
                }
            }
        };

        Ok(Spanned::new(token, start, self.pos))
    }

    fn parse_quoted_string(&mut self, start: usize) -> Result<Token<'a>> {
        let content_start = self.pos;
        let mut has_escape = false;

        // First pass: check if we have escapes
        loop {
            match self.advance() {
                Some(b'"') => break,
                Some(b'\\') => {
                    has_escape = true;
                    self.advance(); // Skip escaped char
                }
                Some(_) => {}
                None => return Err(AsonError::UnclosedString { position: start }),
            }
        }

        let content_end = self.pos - 1; // Exclude closing quote

        if has_escape {
            // Need to allocate and decode escapes
            let mut result = String::new();
            let slice = &self.input[content_start..content_end];
            let mut chars = slice.chars();
            while let Some(c) = chars.next() {
                if c == '\\' {
                    match chars.next() {
                        Some('n') => result.push('\n'),
                        Some('t') => result.push('\t'),
                        Some('r') => result.push('\r'),
                        Some('\\') => result.push('\\'),
                        Some('"') => result.push('"'),
                        Some(other) => result.push(other),
                        None => break,
                    }
                } else {
                    result.push(c);
                }
            }
            Ok(Token::Str(Cow::Owned(result)))
        } else {
            // Zero-copy: borrow directly from input
            Ok(Token::Str(Cow::Borrowed(
                &self.input[content_start..content_end],
            )))
        }
    }

    fn parse_identifier(&mut self, start: usize, first: u8) -> Result<Token<'a>> {
        // Identifier in schema context
        let _ = first; // Already consumed
        while let Some(b) = self.peek() {
            if b == b'{' || b == b'}' || b == b'[' || b == b']' || b == b',' || b == b':' {
                break;
            }
            self.advance();
        }
        let s = &self.input[start..self.pos];
        Ok(Token::Ident(Cow::Borrowed(s)))
    }

    fn parse_value(&mut self, start: usize, _first: u8) -> Result<Token<'a>> {
        let mut has_escape = false;

        // Scan until delimiter
        while let Some(b) = self.peek() {
            match b {
                b',' | b')' | b']' | b'}' => break,
                b'\\' => {
                    has_escape = true;
                    self.advance();
                    self.advance(); // Skip escaped char
                }
                _ => {
                    self.advance();
                }
            }
        }

        let raw = &self.input[start..self.pos];

        if has_escape {
            // Decode escapes
            let mut result = String::new();
            let mut chars = raw.chars();
            while let Some(c) = chars.next() {
                if c == '\\' {
                    match chars.next() {
                        Some('n') => result.push('\n'),
                        Some('t') => result.push('\t'),
                        Some('r') => result.push('\r'),
                        Some('\\') => result.push('\\'),
                        Some(other) => result.push(other),
                        None => break,
                    }
                } else {
                    result.push(c);
                }
            }
            return Ok(Token::Str(Cow::Owned(result)));
        }

        // Try to parse as number or keyword
        match raw {
            "true" => Ok(Token::True),
            "false" => Ok(Token::False),
            _ => {
                if let Ok(n) = raw.parse::<i64>() {
                    Ok(Token::Integer(n))
                } else if let Ok(n) = raw.parse::<f64>() {
                    Ok(Token::Float(n))
                } else {
                    // Zero-copy string
                    Ok(Token::Str(Cow::Borrowed(raw)))
                }
            }
        }
    }
}
