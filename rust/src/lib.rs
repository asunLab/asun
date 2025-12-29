//! ASON (Array-Schema Object Notation) - A token-efficient data format
//!
//! # Example
//! ```
//! use serde::{Deserialize, Serialize};
//! use ason::{from_str, to_string, to_string_pretty};
//!
//! #[derive(Serialize, Deserialize, Debug, PartialEq)]
//! struct User {
//!     name: String,
//!     age: i32,
//! }
//!
//! let user = User { name: "Alice".into(), age: 30 };
//!
//! // Serialize
//! let s = to_string(&user).unwrap();
//! assert_eq!(s, "{name,age}:(Alice,30)");
//!
//! // Deserialize
//! let parsed: User = from_str(&s).unwrap();
//! assert_eq!(parsed, user);
//! ```

mod de;
mod error;
mod ser;

pub use de::{Deserializer, from_str};
pub use error::{Error, Result};
pub use ser::{Serializer, to_string, to_string_pretty};

#[cfg(test)]
mod tests {
    use super::*;
    use serde::{Deserialize, Serialize};
    use std::collections::HashMap;

    #[derive(Serialize, Deserialize, Debug, PartialEq)]
    struct User {
        name: String,
        age: i32,
    }

    #[derive(Serialize, Deserialize, Debug, PartialEq)]
    struct Address {
        city: String,
        zip: i32,
    }

    #[derive(Serialize, Deserialize, Debug, PartialEq)]
    struct Person {
        name: String,
        addr: Address,
    }

    // Complex nested structure
    #[derive(Serialize, Deserialize, Debug, PartialEq)]
    struct Company {
        name: String,
        employees: Vec<User>,
        headquarters: Address,
    }

    // Struct with array field
    #[derive(Serialize, Deserialize, Debug, PartialEq)]
    struct Team {
        name: String,
        scores: Vec<i32>,
    }

    // Struct with optional and boolean
    #[derive(Serialize, Deserialize, Debug, PartialEq)]
    struct Profile {
        username: String,
        verified: bool,
        bio: Option<String>,
    }

    // ========== Basic Tests ==========

    #[test]
    fn test_serialize_simple() {
        let user = User {
            name: "Alice".into(),
            age: 30,
        };
        let s = to_string(&user).unwrap();
        assert_eq!(s, "{name,age}:(Alice,30)");
    }

    #[test]
    fn test_serialize_nested() {
        let person = Person {
            name: "Alice".into(),
            addr: Address {
                city: "NYC".into(),
                zip: 10001,
            },
        };
        let s = to_string(&person).unwrap();
        assert_eq!(s, "{name,addr{city,zip}}:(Alice,(NYC,10001))");
    }

    #[test]
    fn test_serialize_array() {
        let users = vec![
            User {
                name: "Alice".into(),
                age: 30,
            },
            User {
                name: "Bob".into(),
                age: 25,
            },
        ];
        let s = to_string(&users).unwrap();
        assert_eq!(s, "{name,age}:(Alice,30),(Bob,25)");
    }

    #[test]
    fn test_deserialize_simple() {
        let s = "{name,age}:(Alice,30)";
        let user: User = from_str(s).unwrap();
        assert_eq!(
            user,
            User {
                name: "Alice".into(),
                age: 30
            }
        );
    }

    #[test]
    fn test_deserialize_nested() {
        let s = "{name,addr{city,zip}}:(Alice,(NYC,10001))";
        let person: Person = from_str(s).unwrap();
        assert_eq!(person.name, "Alice");
        assert_eq!(person.addr.city, "NYC");
        assert_eq!(person.addr.zip, 10001);
    }

    #[test]
    fn test_deserialize_array() {
        let s = "{name,age}:(Alice,30),(Bob,25)";
        let users: Vec<User> = from_str(s).unwrap();
        assert_eq!(users.len(), 2);
        assert_eq!(users[0].name, "Alice");
        assert_eq!(users[1].name, "Bob");
    }

    #[test]
    fn test_roundtrip() {
        let user = User {
            name: "Charlie".into(),
            age: 35,
        };
        let s = to_string(&user).unwrap();
        let parsed: User = from_str(&s).unwrap();
        assert_eq!(parsed, user);
    }

    #[test]
    fn test_tuple_format() {
        let s = "(Alice,30)";
        let user: User = from_str(s).unwrap();
        assert_eq!(
            user,
            User {
                name: "Alice".into(),
                age: 30
            }
        );
    }

    // ========== Primitive Arrays ==========

    #[test]
    fn test_simple_int_array() {
        let arr = vec![1, 2, 3, 4, 5];
        let s = to_string(&arr).unwrap();
        assert_eq!(s, "[1,2,3,4,5]");
        let parsed: Vec<i32> = from_str(&s).unwrap();
        assert_eq!(parsed, arr);
    }

    #[test]
    fn test_simple_string_array() {
        let arr = vec!["hello", "world"];
        let s = to_string(&arr).unwrap();
        assert_eq!(s, "[hello,world]");
    }

    #[test]
    fn test_nested_int_array() {
        let arr = vec![vec![1, 2, 3], vec![4, 5, 6]];
        let s = to_string(&arr).unwrap();
        assert_eq!(s, "[[1,2,3],[4,5,6]]");
        let parsed: Vec<Vec<i32>> = from_str(&s).unwrap();
        assert_eq!(parsed, arr);
    }

    #[test]
    fn test_triple_nested_array() {
        let arr = vec![vec![vec![1, 2], vec![3, 4]], vec![vec![5, 6], vec![7, 8]]];
        let s = to_string(&arr).unwrap();
        assert_eq!(s, "[[[1,2],[3,4]],[[5,6],[7,8]]]");
        let parsed: Vec<Vec<Vec<i32>>> = from_str(&s).unwrap();
        assert_eq!(parsed, arr);
    }

    // ========== Struct with Array Field ==========

    #[test]
    fn test_struct_with_array_field() {
        let team = Team {
            name: "Alpha".into(),
            scores: vec![100, 95, 88],
        };
        let s = to_string(&team).unwrap();
        assert_eq!(s, "{name,scores[]}:(Alpha,[100,95,88])");
        let parsed: Team = from_str(&s).unwrap();
        assert_eq!(parsed, team);
    }

    // ========== HashMap ==========

    #[test]
    fn test_hashmap_serialize() {
        let mut map: HashMap<String, i32> = HashMap::new();
        map.insert("a".into(), 1);
        let s = to_string(&map).unwrap();
        // HashMap order is not guaranteed, just check format
        assert!(s.contains("a") && s.contains("1"));
    }

    #[test]
    fn test_hashmap_roundtrip() {
        let mut map: HashMap<String, i32> = HashMap::new();
        map.insert("x".into(), 10);
        map.insert("y".into(), 20);
        let s = to_string(&map).unwrap();
        let parsed: HashMap<String, i32> = from_str(&s).unwrap();
        assert_eq!(parsed, map);
    }

    // ========== Complex Nested Structures ==========

    #[test]
    fn test_company_structure() {
        let company = Company {
            name: "TechCorp".into(),
            employees: vec![
                User {
                    name: "Alice".into(),
                    age: 30,
                },
                User {
                    name: "Bob".into(),
                    age: 25,
                },
            ],
            headquarters: Address {
                city: "SF".into(),
                zip: 94102,
            },
        };
        let s = to_string(&company).unwrap();
        // Should have nested schema for employees and headquarters
        assert!(s.contains("{name,employees"));
        assert!(s.contains("headquarters{city,zip}"));

        let parsed: Company = from_str(&s).unwrap();
        assert_eq!(parsed, company);
    }

    // ========== Boolean and Option ==========

    #[test]
    fn test_boolean() {
        let profile = Profile {
            username: "alice".into(),
            verified: true,
            bio: None,
        };
        let s = to_string(&profile).unwrap();
        assert!(s.contains("true"));
        assert!(s.contains("null"));
    }

    #[test]
    fn test_option_some() {
        let profile = Profile {
            username: "bob".into(),
            verified: false,
            bio: Some("Hello world".into()),
        };
        let s = to_string(&profile).unwrap();
        let parsed: Profile = from_str(&s).unwrap();
        assert_eq!(parsed, profile);
    }

    // ========== Special Characters ==========

    #[test]
    fn test_string_with_comma() {
        let user = User {
            name: "Alice, Bob".into(),
            age: 30,
        };
        let s = to_string(&user).unwrap();
        // Should be quoted
        assert!(s.contains("\"Alice, Bob\""));
        let parsed: User = from_str(&s).unwrap();
        assert_eq!(parsed, user);
    }

    #[test]
    fn test_string_with_parentheses() {
        let user = User {
            name: "Test (1)".into(),
            age: 25,
        };
        let s = to_string(&user).unwrap();
        let parsed: User = from_str(&s).unwrap();
        assert_eq!(parsed, user);
    }

    #[test]
    fn test_empty_string() {
        let user = User {
            name: "".into(),
            age: 0,
        };
        let s = to_string(&user).unwrap();
        assert!(s.contains("\"\""));
        let parsed: User = from_str(&s).unwrap();
        assert_eq!(parsed, user);
    }

    // ========== Numeric Edge Cases ==========

    #[test]
    fn test_negative_numbers() {
        let user = User {
            name: "Test".into(),
            age: -5,
        };
        let s = to_string(&user).unwrap();
        let parsed: User = from_str(&s).unwrap();
        assert_eq!(parsed, user);
    }

    #[test]
    fn test_float_array() {
        let arr = vec![1.5, 2.7, -3.5];
        let s = to_string(&arr).unwrap();
        let parsed: Vec<f64> = from_str(&s).unwrap();
        assert_eq!(parsed.len(), 3);
        assert!((parsed[0] - 1.5).abs() < 0.001);
        assert!((parsed[2] - (-3.5)).abs() < 0.001);
    }

    // ========== Empty Collections ==========

    #[test]
    fn test_empty_array() {
        let arr: Vec<i32> = vec![];
        let s = to_string(&arr).unwrap();
        assert_eq!(s, "[]");
        let parsed: Vec<i32> = from_str(&s).unwrap();
        assert_eq!(parsed, arr);
    }

    #[test]
    fn test_struct_with_empty_array() {
        let team = Team {
            name: "Empty".into(),
            scores: vec![],
        };
        let s = to_string(&team).unwrap();
        let parsed: Team = from_str(&s).unwrap();
        assert_eq!(parsed, team);
    }
}
