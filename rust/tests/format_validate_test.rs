#[cfg(test)]
mod format_validation {
    use ason::{decode, error::Error};
    use serde::Deserialize;

    #[derive(Debug, Deserialize, PartialEq)]
    struct FmtUser {
        id: i64,
        name: String,
    }

    const BAD_FMT: &str = "{id:int, name:str}:\n  (1, Alice),\n  (2, Bob),\n  (3, Carol)";
    const GOOD_FMT: &str = "[{id:int, name:str}]:\n  (1, Alice),\n  (2, Bob),\n  (3, Carol)";

    #[test]
    fn test_bad_format_as_vec() {
        println!("=== Rust: decode Vec<FmtUser> with {{schema}}: format ===");
        let result = decode::<Vec<FmtUser>>(BAD_FMT);
        match &result {
            Ok(v) => println!("  SUCCESS (unexpected!): {:?}", v),
            Err(e) => println!("  ERROR (expected): {:?}", e),
        }
        // Should be an error — {}: format is NOT valid for Vec decode
        assert!(result.is_err(), "should reject {{schema}}: format for Vec");
    }

    #[test]
    fn test_bad_format_as_single_trailing_rows() {
        println!("=== Rust: decode single FmtUser with trailing rows ===");
        let result = decode::<FmtUser>(BAD_FMT);
        match &result {
            Ok(u) => println!("  SUCCESS: {:?}  (trailing rows silently dropped?)", u),
            Err(e) => println!("  ERROR: {:?}", e),
        }
        // Trailing characters error expected
        assert!(result.is_err(), "should reject trailing rows");
    }

    #[test]
    fn test_good_format_as_vec() {
        println!("=== Rust: decode Vec<FmtUser> with [{{schema}}]: format ===");
        let result = decode::<Vec<FmtUser>>(GOOD_FMT);
        match &result {
            Ok(v) => println!("  SUCCESS: {:?}", v),
            Err(e) => println!("  ERROR: {:?}", e),
        }
        assert!(result.is_ok(), "should accept [{{}}]: format for Vec");
        let users = result.unwrap();
        assert_eq!(users.len(), 3);
        assert_eq!(users[0].name, "Alice");
        assert_eq!(users[2].name, "Carol");
    }
}
