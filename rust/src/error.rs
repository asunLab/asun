//! Error types for ASON serialization/deserialization.

use std::fmt::{self, Display};

/// Result type for ASON operations.
pub type Result<T> = std::result::Result<T, Error>;

/// Error type for ASON operations.
#[derive(Debug)]
pub struct Error {
    message: String,
}

impl Error {
    pub fn new(msg: impl Into<String>) -> Self {
        Error { message: msg.into() }
    }
}

impl Display for Error {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{}", self.message)
    }
}

impl std::error::Error for Error {}

impl serde::de::Error for Error {
    fn custom<T: Display>(msg: T) -> Self {
        Error::new(msg.to_string())
    }
}

impl serde::ser::Error for Error {
    fn custom<T: Display>(msg: T) -> Self {
        Error::new(msg.to_string())
    }
}

