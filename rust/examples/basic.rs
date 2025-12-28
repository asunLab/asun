//! ASON Basic Usage Examples
//!
//! Run: cargo run --example basic

use ason::{Value, from_value, parse, to_string, to_string_pretty, to_value};
use indexmap::IndexMap;
use serde::{Deserialize, Serialize};

fn main() {
    println!("=== ASON Basic Examples ===\n");

    // 1. Parse a single object
    println!("1. Parse a single object");
    let input = "{name,age}:(Alice,30)";
    println!("   Input: {}", input);

    let value = parse(input).unwrap();
    println!("   name = {:?}", value.get("name").unwrap().as_str());
    println!("   age  = {:?}", value.get("age").unwrap().as_i64());
    println!();

    // 2. Parse multiple records
    println!("2. Parse multiple records");
    let input = "{name,age}:(Alice,30),(Bob,25),(Charlie,35)";
    println!("   Input: {}", input);

    let users = parse(input).unwrap();
    let arr = users.as_array().unwrap();
    println!("   Total {} records:", arr.len());
    for (i, user) in arr.iter().enumerate() {
        let name = user.get("name").unwrap().as_str().unwrap();
        let age = user.get("age").unwrap().as_i64().unwrap();
        println!("   [{}] {} - {} years old", i, name, age);
    }
    println!();

    // 3. Nested objects
    println!("3. Nested objects");
    let input = "{name,addr{city,zip}}:(Alice,(NYC,10001))";
    println!("   Input: {}", input);

    let value = parse(input).unwrap();
    let addr = value.get("addr").unwrap();
    println!("   City: {}", addr.get("city").unwrap().as_str().unwrap());
    println!("   Zip:  {}", addr.get("zip").unwrap().as_i64().unwrap());
    println!();

    // 4. Array fields
    println!("4. Array fields");
    let input = "{name,scores[]}:(Alice,[90,85,92])";
    println!("   Input: {}", input);

    let value = parse(input).unwrap();
    let scores = value.get("scores").unwrap().as_array().unwrap();
    println!("   Scores: {:?}", scores);
    println!();

    // 5. Build and serialize
    println!("5. Build and serialize");
    let mut obj = IndexMap::new();
    obj.insert("name".to_string(), Value::String("Bob".to_string()));
    obj.insert("age".to_string(), Value::Integer(25));
    obj.insert("active".to_string(), Value::Bool(true));
    let value = Value::Object(obj);

    println!("   Compact: {}", to_string(&value));
    println!();

    // 6. Pretty print
    println!("6. Pretty print (nested structure)");
    let input = "{users[{name,age}]}:([(Alice,30),(Bob,25)])";
    let value = parse(input).unwrap();
    println!("   Compact: {}", to_string(&value));
    println!("   Formatted:\n{}", to_string_pretty(&value));

    // 7. Serde integration
    println!("7. Serde integration");

    #[derive(Debug, Serialize, Deserialize)]
    struct User {
        name: String,
        age: i64,
    }

    // Parse then deserialize
    let value = parse("{name,age}:(Alice,30)").unwrap();
    let user: User = from_value(&value).unwrap();
    println!("   Deserialized: {:?}", user);

    // Serialize to Value
    let user = User {
        name: "Charlie".to_string(),
        age: 35,
    };
    let value = to_value(&user).unwrap();
    println!("   Serialized: {}", to_string(&value));
    println!();

    // 8. Null handling
    println!("8. Null handling");
    let input = "{name,email,age}:(Alice,,30)";
    println!("   Input: {}", input);

    let value = parse(input).unwrap();
    println!("   email = {:?}", value.get("email").unwrap());
    println!();

    // 9. Zero-copy parsing
    println!("9. Zero-copy parsing");
    let input = "{msg}:(Hello World)";
    println!("   Input: {}", input);

    let value = ason::zero_copy::parse(input).unwrap();
    println!(
        "   Message: {}",
        value.get("msg").unwrap().as_str().unwrap()
    );

    // Convert to owned
    let owned = value.into_owned();
    println!("   Converted to owned: {:?}", owned);
    println!();

    // 10. Comment support
    println!("10. Comment support");
    let input = "/* user data */ {name,age}:(Alice /* test */, 30)";
    println!("   Input: {}", input);

    let value = parse(input).unwrap();
    println!("   Parsed successfully: {}", to_string(&value));

    println!("\n=== Examples Complete ===");
}
