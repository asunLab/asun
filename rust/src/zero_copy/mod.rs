//! Zero-copy parsing module for ASON.
//!
//! This module provides zero-copy variants of ASON types that borrow from the input
//! string instead of allocating new strings. This is more efficient but requires
//! the parsed values to live no longer than the input string.
//!
//! # Example
//!
//! ```
//! use ason::zero_copy::{Value, parse};
//!
//! let input = "{name,age}:(Alice,30)";
//! let value = parse(input).unwrap();
//! assert_eq!(value.get("name").unwrap().as_str(), Some("Alice"));
//! ```

mod value;
mod lexer;
mod parser;

pub use value::Value;
pub use parser::parse;

