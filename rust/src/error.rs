//! Error types for ASON parsing and serialization.

use std::fmt;

/// Result type alias for ASON operations.
pub type Result<T> = std::result::Result<T, AsonError>;

/// Errors that can occur during ASON parsing or serialization.
#[derive(Debug, Clone, PartialEq)]
pub enum AsonError {
    /// Unexpected end of input while parsing.
    UnexpectedEof { position: usize },

    /// Unexpected character encountered.
    UnexpectedChar { ch: char, position: usize },

    /// Invalid escape sequence.
    InvalidEscape { ch: char, position: usize },

    /// Unclosed string (missing closing quote).
    UnclosedString { position: usize },

    /// Unclosed comment (missing */).
    UnclosedComment { position: usize },

    /// Unclosed bracket or parenthesis.
    UnclosedBracket { bracket: char, position: usize },

    /// Field count mismatch between schema and data.
    FieldCountMismatch {
        expected: usize,
        actual: usize,
        position: usize,
    },

    /// Invalid schema syntax.
    InvalidSchema { position: usize, message: String },

    /// Invalid identifier (field name).
    InvalidIdentifier { name: String, position: usize },

    /// Expected a specific token but found something else.
    ExpectedToken {
        expected: String,
        found: String,
        position: usize,
    },

    /// Generic parse error with custom message.
    ParseError { position: usize, message: String },
}

impl fmt::Display for AsonError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            AsonError::UnexpectedEof { position } => {
                write!(f, "unexpected end of input at position {}", position)
            }
            AsonError::UnexpectedChar { ch, position } => {
                write!(f, "unexpected character '{}' at position {}", ch, position)
            }
            AsonError::InvalidEscape { ch, position } => {
                write!(
                    f,
                    "invalid escape sequence '\\{}' at position {}",
                    ch, position
                )
            }
            AsonError::UnclosedString { position } => {
                write!(f, "unclosed string starting at position {}", position)
            }
            AsonError::UnclosedComment { position } => {
                write!(f, "unclosed comment starting at position {}", position)
            }
            AsonError::UnclosedBracket { bracket, position } => {
                write!(
                    f,
                    "unclosed '{}' starting at position {}",
                    bracket, position
                )
            }
            AsonError::FieldCountMismatch {
                expected,
                actual,
                position,
            } => {
                write!(
                    f,
                    "field count mismatch: schema has {} fields, but data has {} values at position {}",
                    expected, actual, position
                )
            }
            AsonError::InvalidSchema { position, message } => {
                write!(
                    f,
                    "invalid schema syntax at position {}: {}",
                    position, message
                )
            }
            AsonError::InvalidIdentifier { name, position } => {
                write!(f, "invalid identifier '{}' at position {}", name, position)
            }
            AsonError::ExpectedToken {
                expected,
                found,
                position,
            } => {
                write!(
                    f,
                    "expected '{}' but found '{}' at position {}",
                    expected, found, position
                )
            }
            AsonError::ParseError { position, message } => {
                write!(f, "parse error at position {}: {}", position, message)
            }
        }
    }
}

impl std::error::Error for AsonError {}

impl AsonError {
    /// Get the position where the error occurred.
    pub fn position(&self) -> usize {
        match self {
            AsonError::UnexpectedEof { position } => *position,
            AsonError::UnexpectedChar { position, .. } => *position,
            AsonError::InvalidEscape { position, .. } => *position,
            AsonError::UnclosedString { position } => *position,
            AsonError::UnclosedComment { position } => *position,
            AsonError::UnclosedBracket { position, .. } => *position,
            AsonError::FieldCountMismatch { position, .. } => *position,
            AsonError::InvalidSchema { position, .. } => *position,
            AsonError::InvalidIdentifier { position, .. } => *position,
            AsonError::ExpectedToken { position, .. } => *position,
            AsonError::ParseError { position, .. } => *position,
        }
    }
}
