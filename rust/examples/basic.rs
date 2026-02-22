use ason::{
    Result, StructSchema, from_bin, from_bin_vec, from_str, from_str_vec, to_bin, to_bin_vec,
    to_string, to_string_typed, to_string_vec, to_string_vec_typed,
};
use serde::{Deserialize, Serialize};

#[derive(Debug, Serialize, Deserialize, PartialEq)]
struct User {
    id: i64,
    name: String,
    active: bool,
}

impl StructSchema for User {
    fn field_names() -> &'static [&'static str] {
        &["id", "name", "active"]
    }
    fn field_types() -> &'static [&'static str] {
        &["int", "str", "bool"]
    }
    fn serialize_fields(&self, ser: &mut ason::serialize::Serializer) -> Result<()> {
        use serde::Serialize;
        self.id.serialize(&mut *ser)?;
        self.name.serialize(&mut *ser)?;
        self.active.serialize(&mut *ser)?;
        Ok(())
    }
}

fn main() {
    println!("=== ASON Basic Examples ===\n");

    // 1. Serialize a single struct
    let user = User {
        id: 1,
        name: "Alice".into(),
        active: true,
    };
    let ason_str = to_string(&user).unwrap();
    println!("Serialize single struct:");
    println!("  {}\n", ason_str);

    // 2. Serialize with type annotations (to_string_typed)
    let typed_str = to_string_typed(&user).unwrap();
    println!("Serialize with type annotations:");
    println!("  {}\n", typed_str);
    assert!(typed_str.starts_with("{id:int,name:str,active:bool}:"));

    // 3. Deserialize from ASON (accepts both annotated and unannotated)
    let input = "{id:int,name:str,active:bool}:(1,Alice,true)";
    let user: User = from_str(input).unwrap();
    println!("Deserialize single struct:");
    println!("  {:?}\n", user);

    // 4. Serialize a vec of structs (schema-driven)
    let users = vec![
        User {
            id: 1,
            name: "Alice".into(),
            active: true,
        },
        User {
            id: 2,
            name: "Bob".into(),
            active: false,
        },
        User {
            id: 3,
            name: "Carol Smith".into(),
            active: true,
        },
    ];
    let ason_vec = to_string_vec(&users).unwrap();
    println!("Serialize vec (schema-driven):");
    println!("  {}\n", ason_vec);

    // 5. Serialize vec with type annotations (to_string_vec_typed)
    let typed_vec = to_string_vec_typed(&users).unwrap();
    println!("Serialize vec with type annotations:");
    println!("  {}\n", typed_vec);
    assert!(typed_vec.starts_with("{id:int,name:str,active:bool}:"));

    // 6. Deserialize vec
    let input =
        "{id:int,name:str,active:bool}:(1,Alice,true),(2,Bob,false),(3,\"Carol Smith\",true)";
    let users: Vec<User> = from_str_vec(input).unwrap();
    println!("Deserialize vec:");
    for u in &users {
        println!("  {:?}", u);
    }

    // 7. Multiline format
    println!("\nMultiline format:");
    let multiline = "{id:int, name:str, active:bool}:
  (1, Alice, true),
  (2, Bob, false),
  (3, \"Carol Smith\", true)";
    let users: Vec<User> = from_str_vec(multiline).unwrap();
    for u in &users {
        println!("  {:?}", u);
    }

    // 8. Roundtrip (ASON-text + ASON-bin + JSON)
    println!("\n8. Roundtrip (ASON-text vs ASON-bin vs JSON):");
    let original = User {
        id: 42,
        name: "Test User".into(),
        active: true,
    };
    // ASON text
    let ason_str = to_string(&original).unwrap();
    let from_ason: User = from_str(&ason_str).unwrap();
    assert_eq!(original, from_ason);
    // ASON binary
    let ason_bin = to_bin(&original).unwrap();
    let from_bin_val: User = from_bin(&ason_bin).unwrap();
    assert_eq!(original, from_bin_val);
    // JSON
    let json_str = serde_json::to_string(&original).unwrap();
    let from_json: User = serde_json::from_str(&json_str).unwrap();
    assert_eq!(original, from_json);
    println!("  original:     {:?}", original);
    println!("  ASON text:    {} ({} B)", ason_str, ason_str.len());
    println!("  ASON binary:  {} B", ason_bin.len());
    println!("  JSON:         {} ({} B)", json_str, json_str.len());
    println!("  ✓ all 3 formats roundtrip OK");

    // 9. Vec roundtrip (ASON-text + ASON-bin + JSON)
    println!("\n9. Vec roundtrip (ASON-text vs ASON-bin vs JSON):");
    let vec_ason = to_string_vec(&users).unwrap();
    let vec_bin = to_bin_vec(&users).unwrap();
    let vec_json = serde_json::to_string(&users).unwrap();
    let v1: Vec<User> = from_str_vec(&vec_ason).unwrap();
    let v2: Vec<User> = from_bin_vec(&vec_bin).unwrap();
    let v3: Vec<User> = serde_json::from_str(&vec_json).unwrap();
    assert_eq!(users, v1);
    assert_eq!(users, v2);
    assert_eq!(users, v3);
    println!("  ASON text:   {} B", vec_ason.len());
    println!("  ASON binary: {} B", vec_bin.len());
    println!("  JSON:        {} B", vec_json.len());
    println!(
        "  BIN vs JSON: {:.0}% smaller",
        (1.0 - vec_bin.len() as f64 / vec_json.len() as f64) * 100.0
    );
    println!("  ✓ vec roundtrip OK (all 3 formats)");

    // 10. Optional fields
    println!("\n10. Optional fields:");
    #[derive(Debug, Deserialize)]
    struct Item {
        id: i64,
        label: Option<String>,
    }
    let input = "{id,label}:(1,hello)";
    let item: Item = from_str(input).unwrap();
    println!("  with value: {:?}", item);

    let input = "{id,label}:(2,)";
    let item: Item = from_str(input).unwrap();
    println!("  with null:  {:?}", item);

    // 11. Array fields
    println!("\n11. Array fields:");
    #[derive(Debug, Deserialize)]
    struct Tagged {
        name: String,
        tags: Vec<String>,
    }
    let input = "{name,tags}:(Alice,[rust,go,python])";
    let t: Tagged = from_str(input).unwrap();
    println!("  {:?}", t);

    // 12. Comments
    println!("\n12. With comments:");
    let input = "/* user list */ {id,name,active}:(1,Alice,true)";
    let user: User = from_str(input).unwrap();
    println!("  {:?}", user);

    println!("\n=== All examples passed! ===");
}
