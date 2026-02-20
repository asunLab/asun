use ason::{
    Result, StructSchema, from_str, from_str_vec, to_string, to_string_typed, to_string_vec,
    to_string_vec_typed,
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

    // 8. Roundtrip
    println!("\nRoundtrip test:");
    let original = User {
        id: 42,
        name: "Test User".into(),
        active: true,
    };
    let serialized = to_string(&original).unwrap();
    let deserialized: User = from_str(&serialized).unwrap();
    println!("  original:     {:?}", original);
    println!("  serialized:   {}", serialized);
    println!("  deserialized: {:?}", deserialized);
    assert_eq!(original, deserialized);
    println!("  ✓ roundtrip OK");

    // 9. Optional fields
    println!("\nOptional fields:");
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

    // 10. Array fields
    println!("\nArray fields:");
    #[derive(Debug, Deserialize)]
    struct Tagged {
        name: String,
        tags: Vec<String>,
    }
    let input = "{name,tags}:(Alice,[rust,go,python])";
    let t: Tagged = from_str(input).unwrap();
    println!("  {:?}", t);

    // 11. Comments
    println!("\nWith comments:");
    let input = "/* user list */ {id,name,active}:(1,Alice,true)";
    let user: User = from_str(input).unwrap();
    println!("  {:?}", user);

    println!("\n=== All examples passed! ===");
}
