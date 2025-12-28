//! Lexer (tokenizer) for ASON format.
//!
//! The lexer converts an input string into a stream of tokens.

use crate::error::{AsonError, Result};

/// A token in the ASON format.
#[derive(Debug, Clone, PartialEq)]
pub enum Token {
    /// `{` - Start of schema or nested object
    LBrace,
    /// `}` - End of schema or nested object
    RBrace,
    /// `(` - Start of object data
    LParen,
    /// `)` - End of object data
    RParen,
    /// `[` - Start of array
    LBracket,
    /// `]` - End of array
    RBracket,
    /// `,` - Separator
    Comma,
    /// `:` - Schema-data separator
    Colon,
    /// An identifier (field name in schema)
    Ident(String),
    /// A string value (quoted or unquoted)
    Str(String),
    /// An integer value
    Integer(i64),
    /// A floating-point value
    Float(f64),
    /// `true` literal
    True,
    /// `false` literal
    False,
    /// End of input
    Eof,
}

/// A token with its position in the source.
#[derive(Debug, Clone, PartialEq)]
pub struct Spanned {
    pub token: Token,
    pub start: usize,
    pub end: usize,
}

impl Spanned {
    pub fn new(token: Token, start: usize, end: usize) -> Self {
        Self { token, start, end }
    }
}

/// Lexer for tokenizing ASON input.
pub struct Lexer<'a> {
    input: &'a str,
    chars: std::iter::Peekable<std::str::CharIndices<'a>>,
    /// Current position in the input
    pos: usize,
    /// Whether we're currently inside a schema (between { and })
    in_schema: bool,
    /// Brace depth for tracking nested schemas
    brace_depth: usize,
}

impl<'a> Lexer<'a> {
    /// Create a new lexer for the given input.
    pub fn new(input: &'a str) -> Self {
        Self {
            input,
            chars: input.char_indices().peekable(),
            pos: 0,
            in_schema: false,
            brace_depth: 0,
        }
    }

    /// Get the current position.
    pub fn position(&self) -> usize {
        self.pos
    }

    /// Peek at the next character without consuming it.
    fn peek_char(&mut self) -> Option<char> {
        self.chars.peek().map(|(_, c)| *c)
    }

    /// Consume and return the next character.
    fn next_char(&mut self) -> Option<(usize, char)> {
        let result = self.chars.next();
        if let Some((pos, _)) = result {
            self.pos = pos;
        }
        result
    }

    /// Skip whitespace characters.
    fn skip_whitespace(&mut self) {
        while let Some(c) = self.peek_char() {
            if c.is_whitespace() {
                self.next_char();
            } else {
                break;
            }
        }
    }

    /// Skip a block comment /* ... */
    fn skip_comment(&mut self) -> Result<()> {
        let start = self.pos;
        // Already consumed '/'
        if self.peek_char() != Some('*') {
            return Ok(());
        }
        self.next_char(); // consume '*'

        loop {
            match self.next_char() {
                Some((_, '*')) => {
                    if self.peek_char() == Some('/') {
                        self.next_char(); // consume '/'
                        return Ok(());
                    }
                }
                Some(_) => continue,
                None => {
                    return Err(AsonError::UnclosedComment { position: start });
                }
            }
        }
    }

    /// Skip whitespace and comments.
    fn skip_ws_and_comments(&mut self) -> Result<()> {
        loop {
            self.skip_whitespace();
            if self.peek_char() == Some('/') {
                let start = self.pos;
                self.next_char(); // consume '/'
                if self.peek_char() == Some('*') {
                    self.skip_comment()?;
                } else {
                    // Not a comment, restore position
                    self.chars = self.input[start..].char_indices().peekable();
                    self.pos = start;
                    break;
                }
            } else {
                break;
            }
        }
        Ok(())
    }

    /// Parse a quoted string "..."
    fn parse_quoted_string(&mut self) -> Result<String> {
        let start = self.pos;
        let mut s = String::new();

        loop {
            match self.next_char() {
                Some((_, '\\')) => {
                    // Escape sequence
                    match self.next_char() {
                        Some((_, 'n')) => s.push('\n'),
                        Some((_, 't')) => s.push('\t'),
                        Some((_, '\\')) => s.push('\\'),
                        Some((_, '"')) => s.push('"'),
                        Some((_, ',')) => s.push(','),
                        Some((_, '(')) => s.push('('),
                        Some((_, ')')) => s.push(')'),
                        Some((_, '[')) => s.push('['),
                        Some((_, ']')) => s.push(']'),
                        Some((pos, ch)) => {
                            return Err(AsonError::InvalidEscape { ch, position: pos });
                        }
                        None => {
                            return Err(AsonError::UnexpectedEof { position: self.pos });
                        }
                    }
                }
                Some((_, '"')) => {
                    // End of string
                    return Ok(s);
                }
                Some((_, ch)) => {
                    s.push(ch);
                }
                None => {
                    return Err(AsonError::UnclosedString { position: start });
                }
            }
        }
    }

    /// Parse an unquoted string (plain_str)
    /// Stops at delimiters: , ( ) [ ] { } :
    fn parse_plain_string(&mut self, first_char: char) -> Result<String> {
        let mut s = String::new();
        s.push(first_char);

        loop {
            match self.peek_char() {
                Some(c) if is_delimiter(c) => {
                    // Stop at delimiter
                    break;
                }
                Some('\\') => {
                    self.next_char(); // consume '\'
                    match self.next_char() {
                        Some((_, 'n')) => s.push('\n'),
                        Some((_, 't')) => s.push('\t'),
                        Some((_, '\\')) => s.push('\\'),
                        Some((_, '"')) => s.push('"'),
                        Some((_, ',')) => s.push(','),
                        Some((_, '(')) => s.push('('),
                        Some((_, ')')) => s.push(')'),
                        Some((_, '[')) => s.push('['),
                        Some((_, ']')) => s.push(']'),
                        Some((pos, ch)) => {
                            return Err(AsonError::InvalidEscape { ch, position: pos });
                        }
                        None => {
                            return Err(AsonError::UnexpectedEof { position: self.pos });
                        }
                    }
                }
                Some(c) => {
                    self.next_char();
                    s.push(c);
                }
                None => break,
            }
        }

        // Trim whitespace for plain strings
        Ok(s.trim().to_string())
    }

    /// Parse an identifier (field name in schema)
    fn parse_ident(&mut self, first_char: char) -> String {
        let mut s = String::new();
        s.push(first_char);

        while let Some(c) = self.peek_char() {
            if c.is_alphanumeric() || c == '_' {
                self.next_char();
                s.push(c);
            } else {
                break;
            }
        }
        s
    }

    /// Get the next token.
    pub fn next_token(&mut self) -> Result<Spanned> {
        self.skip_ws_and_comments()?;

        let start = self.pos;

        let Some((pos, ch)) = self.next_char() else {
            return Ok(Spanned::new(Token::Eof, start, start));
        };

        let token = match ch {
            '{' => {
                self.in_schema = true;
                self.brace_depth += 1;
                Token::LBrace
            }
            '}' => {
                if self.brace_depth > 0 {
                    self.brace_depth -= 1;
                }
                if self.brace_depth == 0 {
                    self.in_schema = false;
                }
                Token::RBrace
            }
            '(' => Token::LParen,
            ')' => Token::RParen,
            '[' => Token::LBracket,
            ']' => Token::RBracket,
            ',' => Token::Comma,
            ':' => Token::Colon,
            '"' => {
                let s = self.parse_quoted_string()?;
                Token::Str(s)
            }
            _ => {
                if self.in_schema {
                    // In schema mode, parse as identifier
                    if ch.is_alphabetic() || ch == '_' {
                        let ident = self.parse_ident(ch);
                        Token::Ident(ident)
                    } else {
                        return Err(AsonError::UnexpectedChar { ch, position: pos });
                    }
                } else {
                    // In data mode, parse as value
                    self.parse_value(ch, pos)?
                }
            }
        };

        let end = self.pos + 1;
        Ok(Spanned::new(token, start, end))
    }

    /// Parse a value (number, boolean, or string)
    fn parse_value(&mut self, first_char: char, _start_pos: usize) -> Result<Token> {
        // Check for negative number
        if first_char == '-'
            && let Some(c) = self.peek_char()
            && c.is_ascii_digit()
        {
            return self.parse_number(first_char);
        }

        // Check for number
        if first_char.is_ascii_digit() {
            return self.parse_number(first_char);
        }

        // Check for boolean or string
        let s = self.parse_plain_string(first_char)?;

        match s.as_str() {
            "true" => Ok(Token::True),
            "false" => Ok(Token::False),
            "" => Ok(Token::Str(String::new())), // Empty value
            _ => {
                // Try to parse as number
                if let Ok(n) = s.parse::<i64>() {
                    Ok(Token::Integer(n))
                } else if let Ok(n) = s.parse::<f64>() {
                    Ok(Token::Float(n))
                } else {
                    Ok(Token::Str(s))
                }
            }
        }
    }

    /// Parse a number (integer or float)
    fn parse_number(&mut self, first_char: char) -> Result<Token> {
        let mut s = String::new();
        s.push(first_char);

        let mut has_dot = false;

        while let Some(c) = self.peek_char() {
            if c.is_ascii_digit() {
                self.next_char();
                s.push(c);
            } else if c == '.' && !has_dot {
                // Check if next char is a digit
                self.next_char();
                s.push('.');
                has_dot = true;
            } else if is_delimiter(c) || c.is_whitespace() {
                break;
            } else {
                // Part of a string, continue as plain string
                return self.continue_as_string(s);
            }
        }

        if has_dot {
            s.parse::<f64>()
                .map(Token::Float)
                .map_err(|_| AsonError::ParseError {
                    position: self.pos,
                    message: format!("invalid float: {}", s),
                })
        } else {
            s.parse::<i64>()
                .map(Token::Integer)
                .map_err(|_| AsonError::ParseError {
                    position: self.pos,
                    message: format!("invalid integer: {}", s),
                })
        }
    }

    /// Continue parsing as a string after starting with digits
    fn continue_as_string(&mut self, prefix: String) -> Result<Token> {
        let mut s = prefix;

        loop {
            match self.peek_char() {
                Some(c) if is_delimiter(c) => break,
                Some('\\') => {
                    self.next_char();
                    match self.next_char() {
                        Some((_, 'n')) => s.push('\n'),
                        Some((_, 't')) => s.push('\t'),
                        Some((_, '\\')) => s.push('\\'),
                        Some((_, '"')) => s.push('"'),
                        Some((_, ',')) => s.push(','),
                        Some((_, '(')) => s.push('('),
                        Some((_, ')')) => s.push(')'),
                        Some((_, '[')) => s.push('['),
                        Some((_, ']')) => s.push(']'),
                        Some((pos, ch)) => {
                            return Err(AsonError::InvalidEscape { ch, position: pos });
                        }
                        None => break,
                    }
                }
                Some(c) => {
                    self.next_char();
                    s.push(c);
                }
                None => break,
            }
        }

        Ok(Token::Str(s.trim().to_string()))
    }

    /// Tokenize the entire input and return all tokens.
    pub fn tokenize(&mut self) -> Result<Vec<Spanned>> {
        let mut tokens = Vec::new();
        loop {
            let spanned = self.next_token()?;
            let is_eof = spanned.token == Token::Eof;
            tokens.push(spanned);
            if is_eof {
                break;
            }
        }
        Ok(tokens)
    }
}

/// Check if a character is a delimiter
fn is_delimiter(c: char) -> bool {
    matches!(c, ',' | '(' | ')' | '[' | ']' | '{' | '}' | ':')
}

#[cfg(test)]
mod tests {
    use super::*;

    fn tokenize(input: &str) -> Vec<Token> {
        Lexer::new(input)
            .tokenize()
            .unwrap()
            .into_iter()
            .map(|s| s.token)
            .collect()
    }

    #[test]
    fn test_simple_schema() {
        let tokens = tokenize("{name,age}");
        assert_eq!(
            tokens,
            vec![
                Token::LBrace,
                Token::Ident("name".to_string()),
                Token::Comma,
                Token::Ident("age".to_string()),
                Token::RBrace,
                Token::Eof,
            ]
        );
    }

    #[test]
    fn test_schema_with_data() {
        let tokens = tokenize("{name,age}:(Alice,30)");
        assert_eq!(
            tokens,
            vec![
                Token::LBrace,
                Token::Ident("name".to_string()),
                Token::Comma,
                Token::Ident("age".to_string()),
                Token::RBrace,
                Token::Colon,
                Token::LParen,
                Token::Str("Alice".to_string()),
                Token::Comma,
                Token::Integer(30),
                Token::RParen,
                Token::Eof,
            ]
        );
    }

    #[test]
    fn test_quoted_string() {
        let tokens = tokenize(r#"("hello world","  spaces  ")"#);
        assert_eq!(
            tokens,
            vec![
                Token::LParen,
                Token::Str("hello world".to_string()),
                Token::Comma,
                Token::Str("  spaces  ".to_string()),
                Token::RParen,
                Token::Eof,
            ]
        );
    }

    #[test]
    fn test_escape_sequences() {
        let tokens = tokenize(r#"("say \"hi\"","line1\nline2")"#);
        assert_eq!(
            tokens,
            vec![
                Token::LParen,
                Token::Str("say \"hi\"".to_string()),
                Token::Comma,
                Token::Str("line1\nline2".to_string()),
                Token::RParen,
                Token::Eof,
            ]
        );
    }

    #[test]
    fn test_numbers() {
        let tokens = tokenize("(42,-100,3.14,-0.5)");
        assert_eq!(
            tokens,
            vec![
                Token::LParen,
                Token::Integer(42),
                Token::Comma,
                Token::Integer(-100),
                Token::Comma,
                Token::Float(3.14),
                Token::Comma,
                Token::Float(-0.5),
                Token::RParen,
                Token::Eof,
            ]
        );
    }

    #[test]
    fn test_booleans() {
        let tokens = tokenize("(true,false)");
        assert_eq!(
            tokens,
            vec![
                Token::LParen,
                Token::True,
                Token::Comma,
                Token::False,
                Token::RParen,
                Token::Eof,
            ]
        );
    }

    #[test]
    fn test_array() {
        let tokens = tokenize("[1,2,3]");
        assert_eq!(
            tokens,
            vec![
                Token::LBracket,
                Token::Integer(1),
                Token::Comma,
                Token::Integer(2),
                Token::Comma,
                Token::Integer(3),
                Token::RBracket,
                Token::Eof,
            ]
        );
    }

    #[test]
    fn test_comments() {
        let tokens = tokenize("/* comment */ {name} /* another */");
        assert_eq!(
            tokens,
            vec![
                Token::LBrace,
                Token::Ident("name".to_string()),
                Token::RBrace,
                Token::Eof,
            ]
        );
    }

    #[test]
    fn test_plain_string_trim() {
        let tokens = tokenize("(  hello world  ,  foo  )");
        assert_eq!(
            tokens,
            vec![
                Token::LParen,
                Token::Str("hello world".to_string()),
                Token::Comma,
                Token::Str("foo".to_string()),
                Token::RParen,
                Token::Eof,
            ]
        );
    }

    #[test]
    fn test_nested_schema() {
        let tokens = tokenize("{name,addr{city,zip}}");
        assert_eq!(
            tokens,
            vec![
                Token::LBrace,
                Token::Ident("name".to_string()),
                Token::Comma,
                Token::Ident("addr".to_string()),
                Token::LBrace,
                Token::Ident("city".to_string()),
                Token::Comma,
                Token::Ident("zip".to_string()),
                Token::RBrace,
                Token::RBrace,
                Token::Eof,
            ]
        );
    }

    #[test]
    fn test_array_schema() {
        let tokens = tokenize("{scores[]}");
        assert_eq!(
            tokens,
            vec![
                Token::LBrace,
                Token::Ident("scores".to_string()),
                Token::LBracket,
                Token::RBracket,
                Token::RBrace,
                Token::Eof,
            ]
        );
    }

    #[test]
    fn test_unclosed_string_error() {
        let result = Lexer::new(r#"("hello)"#).tokenize();
        assert!(result.is_err());
        assert!(matches!(
            result.unwrap_err(),
            AsonError::UnclosedString { .. }
        ));
    }

    #[test]
    fn test_unclosed_comment_error() {
        let result = Lexer::new("/* unclosed").tokenize();
        assert!(result.is_err());
        assert!(matches!(
            result.unwrap_err(),
            AsonError::UnclosedComment { .. }
        ));
    }

    #[test]
    fn test_empty_string_value() {
        let tokens = tokenize(r#"("","")"#);
        assert_eq!(
            tokens,
            vec![
                Token::LParen,
                Token::Str("".to_string()),
                Token::Comma,
                Token::Str("".to_string()),
                Token::RParen,
                Token::Eof,
            ]
        );
    }

    // ========================================================================
    // Edge case tests
    // ========================================================================

    #[test]
    fn test_unicode_identifier() {
        let tokens = tokenize("{名前,年齢}");
        assert_eq!(
            tokens,
            vec![
                Token::LBrace,
                Token::Ident("名前".to_string()),
                Token::Comma,
                Token::Ident("年齢".to_string()),
                Token::RBrace,
                Token::Eof,
            ]
        );
    }

    #[test]
    fn test_unicode_value() {
        let tokens = tokenize("(北京,上海)");
        assert_eq!(
            tokens,
            vec![
                Token::LParen,
                Token::Str("北京".to_string()),
                Token::Comma,
                Token::Str("上海".to_string()),
                Token::RParen,
                Token::Eof,
            ]
        );
    }

    #[test]
    fn test_emoji_in_string() {
        let tokens = tokenize("(Hello 🌍!)");
        assert_eq!(
            tokens,
            vec![
                Token::LParen,
                Token::Str("Hello 🌍!".to_string()),
                Token::RParen,
                Token::Eof,
            ]
        );
    }

    #[test]
    fn test_negative_number() {
        let tokens = tokenize("(-42,-3.14)");
        assert_eq!(
            tokens,
            vec![
                Token::LParen,
                Token::Integer(-42),
                Token::Comma,
                Token::Float(-3.14),
                Token::RParen,
                Token::Eof,
            ]
        );
    }

    #[test]
    fn test_large_integer() {
        let tokens = tokenize("(9223372036854775807)");
        assert_eq!(
            tokens,
            vec![
                Token::LParen,
                Token::Integer(i64::MAX),
                Token::RParen,
                Token::Eof,
            ]
        );
    }

    #[test]
    fn test_consecutive_commas() {
        let tokens = tokenize("(a,,b)");
        assert_eq!(
            tokens,
            vec![
                Token::LParen,
                Token::Str("a".to_string()),
                Token::Comma,
                Token::Comma,
                Token::Str("b".to_string()),
                Token::RParen,
                Token::Eof,
            ]
        );
    }

    #[test]
    fn test_backslash_escapes() {
        let tokens = tokenize(r#"(hello\, world)"#);
        assert_eq!(
            tokens,
            vec![
                Token::LParen,
                Token::Str("hello, world".to_string()),
                Token::RParen,
                Token::Eof,
            ]
        );
    }

    #[test]
    fn test_quoted_with_all_escapes() {
        let tokens = tokenize(r#"("tab\there\nnewline\\backslash\"quote")"#);
        assert_eq!(
            tokens,
            vec![
                Token::LParen,
                Token::Str("tab\there\nnewline\\backslash\"quote".to_string()),
                Token::RParen,
                Token::Eof,
            ]
        );
    }

    #[test]
    fn test_multiline_comment() {
        let tokens = tokenize("/* line1\nline2\nline3 */ (a)");
        assert_eq!(
            tokens,
            vec![
                Token::LParen,
                Token::Str("a".to_string()),
                Token::RParen,
                Token::Eof,
            ]
        );
    }

    #[test]
    fn test_adjacent_comments() {
        let tokens = tokenize("/* a *//* b */(c)");
        assert_eq!(
            tokens,
            vec![
                Token::LParen,
                Token::Str("c".to_string()),
                Token::RParen,
                Token::Eof,
            ]
        );
    }

    #[test]
    fn test_deeply_nested_brackets() {
        let tokens = tokenize("[[[1]]]");
        assert_eq!(
            tokens,
            vec![
                Token::LBracket,
                Token::LBracket,
                Token::LBracket,
                Token::Integer(1),
                Token::RBracket,
                Token::RBracket,
                Token::RBracket,
                Token::Eof,
            ]
        );
    }

    #[test]
    fn test_zero_values() {
        let tokens = tokenize("(0,0.0)");
        assert_eq!(
            tokens,
            vec![
                Token::LParen,
                Token::Integer(0),
                Token::Comma,
                Token::Float(0.0),
                Token::RParen,
                Token::Eof,
            ]
        );
    }
}
