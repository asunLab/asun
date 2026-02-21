pub mod binary;
pub mod deserialize;
pub mod error;
pub mod serialize;
pub mod simd;

pub use binary::{from_bin, from_bin_vec, to_bin, to_bin_vec};
pub use deserialize::{from_str, from_str_vec};
pub use error::{Error, Result};
pub use serialize::{StructSchema, to_string, to_string_typed, to_string_vec, to_string_vec_typed};

#[cfg(test)]
mod tests {
    use super::*;
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
        fn serialize_fields(&self, ser: &mut serialize::Serializer) -> Result<()> {
            use serde::Serialize;
            self.id.serialize(&mut *ser)?;
            self.name.serialize(&mut *ser)?;
            self.active.serialize(&mut *ser)?;
            Ok(())
        }
        // field_types not overridden — default returns &[]
    }

    #[test]
    fn test_serialize_struct() {
        let user = User {
            id: 1,
            name: "Alice".into(),
            active: true,
        };
        let s = to_string(&user).unwrap();
        assert_eq!(s, "{id,name,active}:(1,Alice,true)");
    }

    #[test]
    fn test_deserialize_struct_with_schema() {
        let input = "{id,name,active}:(1,Alice,true)";
        let user: User = from_str(input).unwrap();
        assert_eq!(user.id, 1);
        assert_eq!(user.name, "Alice");
        assert!(user.active);
    }

    #[test]
    fn test_deserialize_struct_with_typed_schema() {
        let input = "{id:int,name:str,active:bool}:(1,Alice,true)";
        let user: User = from_str(input).unwrap();
        assert_eq!(user.id, 1);
        assert_eq!(user.name, "Alice");
        assert!(user.active);
    }

    #[test]
    fn test_roundtrip() {
        let user = User {
            id: 42,
            name: "Bob".into(),
            active: false,
        };
        let s = to_string(&user).unwrap();
        let user2: User = from_str(&s).unwrap();
        assert_eq!(user, user2);
    }

    #[test]
    fn test_vec_deserialize() {
        let input = "{id:int,name:str,active:bool}:(1,Alice,true),(2,Bob,false)";
        let users: Vec<User> = from_str_vec(input).unwrap();
        assert_eq!(users.len(), 2);
        assert_eq!(users[0].name, "Alice");
        assert_eq!(users[1].name, "Bob");
    }

    #[test]
    fn test_multiline() {
        let input = "{id:int,name:str,active:bool}:
  (1, Alice, true),
  (2, Bob, false)";
        let users: Vec<User> = from_str_vec(input).unwrap();
        assert_eq!(users.len(), 2);
    }

    #[test]
    fn test_quoted_string() {
        let input = "{id,name,active}:(1,\"Carol Smith\",true)";
        let user: User = from_str(input).unwrap();
        assert_eq!(user.name, "Carol Smith");
    }

    #[test]
    fn test_optional_field() {
        #[derive(Debug, Deserialize, PartialEq)]
        struct Item {
            id: i64,
            label: Option<String>,
        }
        let input = "{id,label}:(1,)";
        let item: Item = from_str(input).unwrap();
        assert_eq!(item.id, 1);
        assert_eq!(item.label, None);
    }

    #[test]
    fn test_array_field() {
        #[derive(Debug, Deserialize, PartialEq)]
        struct Tagged {
            name: String,
            tags: Vec<String>,
        }
        let input = "{name,tags}:(Alice,[rust,go])";
        let t: Tagged = from_str(input).unwrap();
        assert_eq!(t.tags, vec!["rust", "go"]);
    }

    #[test]
    fn test_nested_struct() {
        #[derive(Debug, Deserialize, PartialEq)]
        struct Dept {
            title: String,
        }
        #[derive(Debug, Deserialize, PartialEq)]
        struct Employee {
            name: String,
            dept: Dept,
        }
        let input = "{name,dept:{title}}:(Alice,(Manager))";
        let e: Employee = from_str(input).unwrap();
        assert_eq!(e.name, "Alice");
        assert_eq!(e.dept.title, "Manager");
    }

    #[test]
    fn test_serialize_vec() {
        #[derive(Debug, Serialize)]
        struct Row {
            id: i64,
            name: String,
        }
        impl StructSchema for Row {
            fn field_names() -> &'static [&'static str] {
                &["id", "name"]
            }
            fn serialize_fields(&self, ser: &mut serialize::Serializer) -> Result<()> {
                use serde::Serialize;
                self.id.serialize(&mut *ser)?;
                self.name.serialize(&mut *ser)?;
                Ok(())
            }
        }
        let rows = vec![
            Row {
                id: 1,
                name: "Alice".into(),
            },
            Row {
                id: 2,
                name: "Bob".into(),
            },
        ];
        let s = to_string_vec(&rows).unwrap();
        assert_eq!(s, "{id,name}:(1,Alice),(2,Bob)");
    }

    #[test]
    fn test_escape_roundtrip() {
        #[derive(Debug, Serialize, Deserialize, PartialEq)]
        struct Note {
            text: String,
        }
        let note = Note {
            text: "hello, world (test)".into(),
        };
        let s = to_string(&note).unwrap();
        let note2: Note = from_str(&s).unwrap();
        assert_eq!(note, note2);
    }

    #[test]
    fn test_trailing_comma() {
        let input = "{id:int,name:str,active:bool}:(1,Alice,true),(2,Bob,false),";
        let users: Vec<User> = from_str_vec(input).unwrap();
        assert_eq!(users.len(), 2);
    }

    #[test]
    fn test_comment_stripping() {
        let input = "/* users */ {id,name,active}:(1,Alice,true)";
        let user: User = from_str(input).unwrap();
        assert_eq!(user.id, 1);
    }

    #[test]
    fn test_float_field() {
        #[derive(Debug, Deserialize, PartialEq)]
        struct Score {
            id: i64,
            value: f64,
        }
        let input = "{id,value}:(1,95.5)";
        let s: Score = from_str(input).unwrap();
        assert_eq!(s.value, 95.5);
    }

    #[test]
    fn test_map_field() {
        use std::collections::HashMap;
        #[derive(Debug, Deserialize, PartialEq)]
        struct Item {
            name: String,
            attrs: HashMap<String, i64>,
        }
        let input = "{name,attrs}:(Alice,[(age,30),(score,95)])";
        let item: Item = from_str(input).unwrap();
        assert_eq!(item.attrs["age"], 30);
        assert_eq!(item.attrs["score"], 95);
    }

    // ===================================================================
    // Annotated vs Unannotated schema tests
    // ===================================================================

    #[test]
    fn test_annotated_simple_struct() {
        let typed = "{id:int,name:str,active:bool}:(42,Bob,false)";
        let untyped = "{id,name,active}:(42,Bob,false)";
        let u1: User = from_str(typed).unwrap();
        let u2: User = from_str(untyped).unwrap();
        assert_eq!(u1, u2);
        assert_eq!(u1.id, 42);
        assert_eq!(u1.name, "Bob");
        assert!(!u1.active);
    }

    #[test]
    fn test_annotated_vec() {
        let typed = "{id:int,name:str,active:bool}:(1,Alice,true),(2,Bob,false)";
        let untyped = "{id,name,active}:(1,Alice,true),(2,Bob,false)";
        let v1: Vec<User> = from_str_vec(typed).unwrap();
        let v2: Vec<User> = from_str_vec(untyped).unwrap();
        assert_eq!(v1, v2);
        assert_eq!(v1.len(), 2);
    }

    #[test]
    fn test_annotated_nested_struct() {
        #[derive(Debug, Deserialize, PartialEq)]
        struct Dept {
            title: String,
            budget: f64,
        }
        #[derive(Debug, Deserialize, PartialEq)]
        struct Employee {
            name: String,
            age: i64,
            dept: Dept,
            active: bool,
        }

        let typed = "{name:str,age:int,dept:{title:str,budget:float},active:bool}:(Alice,30,(Engineering,50000.5),true)";
        let untyped = "{name,age,dept:{title,budget},active}:(Alice,30,(Engineering,50000.5),true)";
        let e1: Employee = from_str(typed).unwrap();
        let e2: Employee = from_str(untyped).unwrap();
        assert_eq!(e1, e2);
        assert_eq!(e1.name, "Alice");
        assert_eq!(e1.dept.title, "Engineering");
        assert_eq!(e1.dept.budget, 50000.5);
    }

    #[test]
    fn test_annotated_with_arrays() {
        #[derive(Debug, Deserialize, PartialEq)]
        struct Profile {
            name: String,
            scores: Vec<i64>,
            tags: Vec<String>,
        }

        let typed = "{name:str,scores:[int],tags:[str]}:(Alice,[90,85,92],[rust,go])";
        let untyped = "{name,scores,tags}:(Alice,[90,85,92],[rust,go])";
        let p1: Profile = from_str(typed).unwrap();
        let p2: Profile = from_str(untyped).unwrap();
        assert_eq!(p1, p2);
        assert_eq!(p1.scores, vec![90, 85, 92]);
        assert_eq!(p1.tags, vec!["rust", "go"]);
    }

    #[test]
    fn test_annotated_with_map() {
        use std::collections::HashMap;
        #[derive(Debug, Deserialize, PartialEq)]
        struct Config {
            name: String,
            attrs: HashMap<String, i64>,
        }

        let typed = "{name:str,attrs:map[str,int]}:(server,[(port,8080),(timeout,30)])";
        let untyped = "{name,attrs}:(server,[(port,8080),(timeout,30)])";
        let c1: Config = from_str(typed).unwrap();
        let c2: Config = from_str(untyped).unwrap();
        assert_eq!(c1, c2);
        assert_eq!(c1.attrs["port"], 8080);
    }

    #[test]
    fn test_annotated_with_optional() {
        #[derive(Debug, Deserialize, PartialEq)]
        struct Record {
            id: i64,
            label: Option<String>,
            score: Option<f64>,
        }

        let typed = "{id:int,label:str,score:float}:(1,hello,95.5)";
        let untyped = "{id,label,score}:(1,hello,95.5)";
        let r1: Record = from_str(typed).unwrap();
        let r2: Record = from_str(untyped).unwrap();
        assert_eq!(r1, r2);

        // Test with None values
        let typed_none = "{id:int,label:str,score:float}:(2,,)";
        let untyped_none = "{id,label,score}:(2,,)";
        let r3: Record = from_str(typed_none).unwrap();
        let r4: Record = from_str(untyped_none).unwrap();
        assert_eq!(r3, r4);
        assert_eq!(r3.label, None);
        assert_eq!(r3.score, None);
    }

    #[test]
    fn test_annotated_deep_nesting() {
        #[derive(Debug, Deserialize, PartialEq)]
        struct Task {
            title: String,
            done: bool,
        }
        #[derive(Debug, Deserialize, PartialEq)]
        struct Project {
            name: String,
            tasks: Vec<Task>,
        }
        #[derive(Debug, Deserialize, PartialEq)]
        struct Team {
            lead: String,
            projects: Vec<Project>,
        }
        #[derive(Debug, Deserialize, PartialEq)]
        struct Company {
            name: String,
            revenue: f64,
            team: Team,
        }

        let typed = "{name:str,revenue:float,team:{lead:str,projects:[{name:str,tasks:[{title:str,done:bool}]}]}}:(Acme,500.5,(Alice,[(API,[(Design,true),(Code,false)])]))";
        let untyped = "{name,revenue,team:{lead,projects:[{name,tasks:[{title,done}]}]}}:(Acme,500.5,(Alice,[(API,[(Design,true),(Code,false)])]))";
        let c1: Company = from_str(typed).unwrap();
        let c2: Company = from_str(untyped).unwrap();
        assert_eq!(c1, c2);
        assert_eq!(c1.name, "Acme");
        assert_eq!(c1.team.lead, "Alice");
        assert_eq!(c1.team.projects[0].name, "API");
        assert_eq!(c1.team.projects[0].tasks[0].title, "Design");
        assert!(c1.team.projects[0].tasks[0].done);
        assert!(!c1.team.projects[0].tasks[1].done);
    }

    #[test]
    fn test_annotated_mixed_partial() {
        // Only some fields have type annotations
        #[derive(Debug, Deserialize, PartialEq)]
        struct Mixed {
            id: i64,
            name: String,
            score: f64,
            active: bool,
        }

        let partial = "{id:int,name,score:float,active}:(1,Alice,95.5,true)";
        let full = "{id:int,name:str,score:float,active:bool}:(1,Alice,95.5,true)";
        let none = "{id,name,score,active}:(1,Alice,95.5,true)";
        let m1: Mixed = from_str(partial).unwrap();
        let m2: Mixed = from_str(full).unwrap();
        let m3: Mixed = from_str(none).unwrap();
        assert_eq!(m1, m2);
        assert_eq!(m2, m3);
    }

    #[test]
    fn test_serializer_output_is_unannotated() {
        // Verify that to_string outputs schema without type annotations
        let user = User {
            id: 1,
            name: "Alice".into(),
            active: true,
        };
        let s = to_string(&user).unwrap();
        assert_eq!(s, "{id,name,active}:(1,Alice,true)");
        // Confirm it does NOT contain type hints
        assert!(!s.contains(":int"));
        assert!(!s.contains(":str"));
        assert!(!s.contains(":bool"));
    }

    // ===================================================================
    // Typed serialization tests (to_string_typed / to_string_vec_typed)
    // ===================================================================

    #[test]
    fn test_to_string_typed_simple() {
        let user = User {
            id: 1,
            name: "Alice".into(),
            active: true,
        };
        let s = to_string_typed(&user).unwrap();
        assert_eq!(s, "{id:int,name:str,active:bool}:(1,Alice,true)");
    }

    #[test]
    fn test_to_string_typed_roundtrip() {
        let user = User {
            id: 42,
            name: "Bob".into(),
            active: false,
        };
        let s = to_string_typed(&user).unwrap();
        assert!(s.starts_with("{id:int,name:str,active:bool}:"));
        // Deserialize back
        let user2: User = from_str(&s).unwrap();
        assert_eq!(user, user2);
    }

    #[test]
    fn test_to_string_typed_floats() {
        #[derive(Debug, Serialize, Deserialize, PartialEq)]
        struct Score {
            id: i64,
            value: f64,
            label: String,
        }
        let s = Score {
            id: 1,
            value: 95.5,
            label: "good".into(),
        };
        let out = to_string_typed(&s).unwrap();
        assert_eq!(out, "{id:int,value:float,label:str}:(1,95.5,good)");
        let s2: Score = from_str(&out).unwrap();
        assert_eq!(s, s2);
    }

    #[test]
    fn test_to_string_typed_all_primitives() {
        #[derive(Debug, Serialize, Deserialize, PartialEq)]
        struct All {
            b: bool,
            i: i64,
            u: u32,
            f: f64,
            c: char,
            s: String,
        }
        let val = All {
            b: true,
            i: -42,
            u: 100,
            f: 3.15,
            c: 'A',
            s: "hello".into(),
        };
        let out = to_string_typed(&val).unwrap();
        assert_eq!(
            out,
            "{b:bool,i:int,u:int,f:float,c:str,s:str}:(true,-42,100,3.15,A,hello)"
        );
        let val2: All = from_str(&out).unwrap();
        assert_eq!(val, val2);
    }

    #[test]
    fn test_to_string_typed_optional() {
        #[derive(Debug, Serialize, Deserialize, PartialEq)]
        struct Opt {
            id: i64,
            label: Option<String>,
            score: Option<f64>,
        }

        // Some values
        let v1 = Opt {
            id: 1,
            label: Some("hello".into()),
            score: Some(95.5),
        };
        let out1 = to_string_typed(&v1).unwrap();
        assert_eq!(out1, "{id:int,label:str,score:float}:(1,hello,95.5)");

        // None values — type hint may not be emitted for None
        let v2 = Opt {
            id: 2,
            label: None,
            score: None,
        };
        let out2 = to_string_typed(&v2).unwrap();
        // None fields don't have type hints since no value is serialized
        assert_eq!(out2, "{id:int,label,score}:(2,,)");

        // Both roundtrip correctly
        let v1b: Opt = from_str(&out1).unwrap();
        assert_eq!(v1, v1b);
        let v2b: Opt = from_str(&out2).unwrap();
        assert_eq!(v2, v2b);
    }

    #[test]
    fn test_to_string_typed_nested_struct() {
        // Nested structs: inner struct becomes a tuple, no type hint on it
        #[derive(Debug, Serialize, Deserialize, PartialEq)]
        struct Dept {
            title: String,
        }
        #[derive(Debug, Serialize, Deserialize, PartialEq)]
        struct Employee {
            name: String,
            dept: Dept,
            active: bool,
        }
        let e = Employee {
            name: "Alice".into(),
            dept: Dept {
                title: "Engineering".into(),
            },
            active: true,
        };
        let out = to_string_typed(&e).unwrap();
        // dept is a nested struct — no simple type hint
        assert_eq!(
            out,
            "{name:str,dept,active:bool}:(Alice,(Engineering),true)"
        );
        let e2: Employee = from_str(&out).unwrap();
        assert_eq!(e, e2);
    }

    #[test]
    fn test_to_string_vec_typed() {
        #[derive(Debug, Serialize, Deserialize, PartialEq)]
        struct Row {
            id: i64,
            name: String,
            score: f64,
        }
        impl StructSchema for Row {
            fn field_names() -> &'static [&'static str] {
                &["id", "name", "score"]
            }
            fn field_types() -> &'static [&'static str] {
                &["int", "str", "float"]
            }
            fn serialize_fields(&self, ser: &mut serialize::Serializer) -> Result<()> {
                use serde::Serialize;
                self.id.serialize(&mut *ser)?;
                self.name.serialize(&mut *ser)?;
                self.score.serialize(&mut *ser)?;
                Ok(())
            }
        }
        let rows = vec![
            Row {
                id: 1,
                name: "Alice".into(),
                score: 95.5,
            },
            Row {
                id: 2,
                name: "Bob".into(),
                score: 87.0,
            },
        ];

        let untyped = to_string_vec(&rows).unwrap();
        assert_eq!(untyped, "{id,name,score}:(1,Alice,95.5),(2,Bob,87.0)");

        let typed = to_string_vec_typed(&rows).unwrap();
        assert_eq!(
            typed,
            "{id:int,name:str,score:float}:(1,Alice,95.5),(2,Bob,87.0)"
        );

        // Both deserialize identically
        let r1: Vec<Row> = from_str_vec(&untyped).unwrap();
        let r2: Vec<Row> = from_str_vec(&typed).unwrap();
        assert_eq!(r1, r2);
    }

    #[test]
    fn test_to_string_vec_typed_without_field_types() {
        // StructSchema without field_types — falls back to unannotated
        let users = vec![User {
            id: 1,
            name: "Alice".into(),
            active: true,
        }];

        let typed = to_string_vec_typed(&users).unwrap();
        // No types since field_types() returns empty
        assert_eq!(typed, "{id,name,active}:(1,Alice,true)");
    }
}
