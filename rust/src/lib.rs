//! # ASON - Array-Schema Object Notation
//!
//! A token-efficient data format that separates schema from data.
//!
//! ## Example
//!
//! ```text
//! {name,age}:(Alice,30),(Bob,25)
//! ```
//!
//! ## Zero-Copy Parsing
//!
//! For high-performance scenarios, use the `zero_copy` module which borrows
//! string data directly from the input instead of allocating:
//!
//! ```rust
//! use ason::zero_copy;
//!
//! let input = "{name,age}:(Alice,30)";
//! let value = zero_copy::parse(input).unwrap();
//! assert_eq!(value.get("name").unwrap().as_str(), Some("Alice"));
//! ```

pub mod ast;
pub mod error;
pub mod lexer;
pub mod parser;
pub mod serde_impl;
pub mod serializer;
pub mod zero_copy;

pub use ast::{Schema, SchemaField, Value};
pub use error::{AsonError, Result};
pub use lexer::{Lexer, Spanned, Token};
pub use parser::{Parser, parse};
pub use serde_impl::{FromValueError, ToValueError, from_value, to_value};
pub use serializer::{SerializeConfig, to_string, to_string_pretty, to_string_with_config};
