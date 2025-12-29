//! ASON Basic Usage Examples
//!
//! Run: cargo run --example basic

use ason::{from_str, to_string, to_string_pretty};
use serde::{Deserialize, Serialize};

#[derive(Debug, Serialize, Deserialize, PartialEq)]
struct User {
    name: String,
    age: i32,
}

#[derive(Debug, Serialize, Deserialize, PartialEq)]
struct Address {
    city: String,
    zip: i32,
}

#[derive(Debug, Serialize, Deserialize, PartialEq)]
struct Person {
    name: String,
    addr: Address,
}

fn main() {
    println!("=== ASON Basic Examples ===\n");

    // 1. Simple struct
    println!("1. Simple struct");
    let user = User {
        name: "Alice".into(),
        age: 30,
    };
    let s = to_string(&user).unwrap();
    println!("   Serialized: {}", s);

    let parsed: User = from_str(&s).unwrap();
    println!("   Parsed: {:?}", parsed);
    assert_eq!(user, parsed);
    println!();

    // 2. Nested struct
    println!("2. Nested struct");
    let person = Person {
        name: "Alice".into(),
        addr: Address {
            city: "NYC".into(),
            zip: 10001,
        },
    };
    let s = to_string(&person).unwrap();
    println!("   Serialized: {}", s);

    let parsed: Person = from_str(&s).unwrap();
    println!("   Parsed: {:?}", parsed);
    println!();

    // 3. Array of structs
    println!("3. Array of structs");
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
    println!("   Serialized: {}", s);
    println!();

    // 4. Pretty print
    println!("4. Pretty print");
    let s = to_string_pretty(&person).unwrap();
    println!("   Formatted:\n{}", s);
    println!();

    // 5. Pretty print
    println!("5. Pretty print");
    let persons = vec![
        Person {
            name: "Alice".into(),
            addr: Address {
                city: "NYC".into(),
                zip: 10001,
            },
        },
        Person {
            name: "Bob".into(),
            addr: Address {
                city: "LA".into(),
                zip: 90001,
            },
        },
    ];
    let s = to_string_pretty(&persons).unwrap();
    println!("   Formatted:\n{}", s);
    println!();

    // 6. Parse without schema (using struct field names)
    println!("6. Parse tuple format");
    let input = "(Alice,30)";
    let user: User = from_str(input).unwrap();
    println!("   Input: {}", input);
    println!("   Parsed: {:?}", user);
    println!();

    // 7. Struct with array field
    println!("7. Struct with array field");
    #[derive(Debug, Serialize, Deserialize)]
    struct Company {
        name: String,
        employees: Vec<User>,
        headquarters: Address,
    }
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
    println!("   Serialized: {}", s);
    println!();

    // 8. HashMap
    println!("8. HashMap");
    use std::collections::HashMap;
    let mut map: HashMap<String, i32> = HashMap::new();
    map.insert("x".into(), 10);
    map.insert("y".into(), 20);
    let s = to_string(&map).unwrap();
    println!("   Serialized: {}", s);
    println!();

    println!("=== Examples Complete ===");
}
