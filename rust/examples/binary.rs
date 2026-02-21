//! ASON Binary Format — usage examples and performance benchmark.
//!
//! Run:  cargo run --release --example binary

use ason::{StructSchema, from_bin, from_bin_vec, to_bin, to_bin_vec};
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::time::Instant;

// ===========================================================================
// Example structs
// ===========================================================================

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
struct User {
    id: i64,
    name: String,
    email: String,
    age: i64,
    score: f64,
    active: bool,
    role: String,
    city: String,
}

impl StructSchema for User {
    fn field_names() -> &'static [&'static str] {
        &[
            "id", "name", "email", "age", "score", "active", "role", "city",
        ]
    }
    fn serialize_fields(&self, ser: &mut ason::serialize::Serializer) -> ason::Result<()> {
        use serde::Serialize;
        self.id.serialize(&mut *ser)?;
        self.name.serialize(&mut *ser)?;
        self.email.serialize(&mut *ser)?;
        self.age.serialize(&mut *ser)?;
        self.score.serialize(&mut *ser)?;
        self.active.serialize(&mut *ser)?;
        self.role.serialize(&mut *ser)?;
        self.city.serialize(&mut *ser)?;
        Ok(())
    }
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
struct AllTypes {
    b: bool,
    i8v: i8,
    i16v: i16,
    i32v: i32,
    i64v: i64,
    u8v: u8,
    u16v: u16,
    u32v: u32,
    u64v: u64,
    f32v: f32,
    f64v: f64,
    s: String,
    opt_some: Option<i64>,
    opt_none: Option<i64>,
    vec_int: Vec<i64>,
    vec_str: Vec<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
struct Task {
    id: i64,
    title: String,
    priority: i64,
    done: bool,
    hours: f64,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
struct Project {
    name: String,
    budget: f64,
    active: bool,
    tasks: Vec<Task>,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
struct Team {
    name: String,
    lead: String,
    size: i64,
    projects: Vec<Project>,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
struct Division {
    name: String,
    location: String,
    headcount: i64,
    teams: Vec<Team>,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
struct Company {
    name: String,
    founded: i64,
    revenue_m: f64,
    public: bool,
    divisions: Vec<Division>,
    tags: Vec<String>,
}

// ===========================================================================
// Data generators
// ===========================================================================

fn make_user(i: usize) -> User {
    let names = [
        "Alice", "Bob", "Carol", "David", "Eve", "Frank", "Grace", "Hank",
    ];
    let roles = ["engineer", "designer", "manager", "analyst"];
    let cities = ["NYC", "LA", "Chicago", "Houston", "Phoenix"];
    User {
        id: i as i64,
        name: names[i % names.len()].into(),
        email: format!("{}@example.com", names[i % names.len()].to_lowercase()),
        age: 25 + (i % 40) as i64,
        score: 50.0 + (i % 50) as f64 + 0.5,
        active: i % 3 != 0,
        role: roles[i % roles.len()].into(),
        city: cities[i % cities.len()].into(),
    }
}

fn make_company(i: usize) -> Company {
    Company {
        name: format!("Corp_{}", i),
        founded: 1990 + (i % 35) as i64,
        revenue_m: 10.0 + i as f64 * 5.5,
        public: i % 2 == 0,
        divisions: (0..2)
            .map(|d| Division {
                name: format!("Div_{}_{}", i, d),
                location: ["NYC", "London", "Tokyo", "Berlin"][d % 4].into(),
                headcount: 50 + (d * 20) as i64,
                teams: (0..2)
                    .map(|t| Team {
                        name: format!("Team_{}_{}_{}", i, d, t),
                        lead: ["Alice", "Bob", "Carol", "David"][t % 4].into(),
                        size: 5 + (t * 2) as i64,
                        projects: (0..3)
                            .map(|p| Project {
                                name: format!("Proj_{}_{}", t, p),
                                budget: 100.0 + p as f64 * 50.5,
                                active: p % 2 == 0,
                                tasks: (0..4)
                                    .map(|tk| Task {
                                        id: (i * 100 + d * 10 + t * 5 + tk) as i64,
                                        title: format!("Task_{}", tk),
                                        priority: (tk % 3 + 1) as i64,
                                        done: tk % 2 == 0,
                                        hours: 2.0 + tk as f64 * 1.5,
                                    })
                                    .collect(),
                            })
                            .collect(),
                    })
                    .collect(),
            })
            .collect(),
        tags: vec![
            "enterprise".into(),
            "tech".into(),
            format!("sector_{}", i % 5),
        ],
    }
}

fn format_bytes(b: usize) -> String {
    if b >= 1_048_576 {
        format!("{:.1} MB", b as f64 / 1_048_576.0)
    } else if b >= 1024 {
        format!("{:.1} KB", b as f64 / 1024.0)
    } else {
        format!("{} B", b)
    }
}

// ===========================================================================
// Part 1 — Basic usage examples
// ===========================================================================

fn section_basic() {
    println!("\n┌─────────────────────────────────────────────┐");
    println!("│  Part 1: Basic to_bin / from_bin usage       │");
    println!("└─────────────────────────────────────────────┘");

    // --- Single struct ---
    let user = make_user(0);
    let bytes = to_bin(&user).unwrap();
    let user2: User = from_bin(&bytes).unwrap();
    assert_eq!(user, user2);
    println!(
        "  Single User roundtrip: {} bytes (JSON: {} B)",
        bytes.len(),
        serde_json::to_string(&user).unwrap().len()
    );

    // --- All primitive types ---
    let at = AllTypes {
        b: true,
        i8v: -1,
        i16v: -300,
        i32v: -70000,
        i64v: i64::MIN,
        u8v: 255,
        u16v: 65535,
        u32v: u32::MAX,
        u64v: u64::MAX,
        f32v: 3.15,
        f64v: 2.718281828459045,
        s: "hello world".into(),
        opt_some: Some(42),
        opt_none: None,
        vec_int: vec![1, 2, 3, 4, 5],
        vec_str: vec!["rust".into(), "fast".into()],
    };
    let at_bytes = to_bin(&at).unwrap();
    let at2: AllTypes = from_bin(&at_bytes).unwrap();
    assert_eq!(at, at2);
    println!(
        "  AllTypes roundtrip:   {} bytes (JSON: {} B)",
        at_bytes.len(),
        serde_json::to_string(&at).unwrap().len()
    );

    // --- Vec<T> via to_bin_vec / from_bin_vec ---
    let users: Vec<User> = (0..5).map(make_user).collect();
    let vec_bytes = to_bin_vec(&users).unwrap();
    let users2: Vec<User> = from_bin_vec(&vec_bytes).unwrap();
    assert_eq!(users, users2);
    println!(
        "  Vec<User> × 5:        {} bytes (JSON: {} B)",
        vec_bytes.len(),
        serde_json::to_string(&users).unwrap().len()
    );

    // --- HashMap ---
    #[derive(Debug, Serialize, Deserialize, PartialEq)]
    struct Profile {
        name: String,
        attrs: HashMap<String, i64>,
    }
    let mut p = Profile {
        name: "Alice".into(),
        attrs: HashMap::new(),
    };
    p.attrs.insert("age".into(), 30);
    p.attrs.insert("score".into(), 95);
    let p_bytes = to_bin(&p).unwrap();
    let p2: Profile = from_bin(&p_bytes).unwrap();
    assert_eq!(p, p2);
    println!(
        "  Profile+HashMap:      {} bytes (JSON: {} B)",
        p_bytes.len(),
        serde_json::to_string(&p).unwrap().len()
    );

    // --- Enums ---
    #[derive(Debug, Serialize, Deserialize, PartialEq)]
    enum Status {
        Active,
        Inactive,
        Custom(i64, String),
    }
    for s in [
        Status::Active,
        Status::Inactive,
        Status::Custom(42, "x".into()),
    ] {
        let b = to_bin(&s).unwrap();
        let s2: Status = from_bin(&b).unwrap();
        assert_eq!(s, s2);
    }
    println!("  Enum variants:        OK");

    // --- 5-level deep struct ---
    let company = make_company(0);
    let co_bytes = to_bin(&company).unwrap();
    let company2: Company = from_bin(&co_bytes).unwrap();
    assert_eq!(company, company2);
    println!(
        "  5-level deep Company: {} bytes (JSON: {} B)",
        co_bytes.len(),
        serde_json::to_string(&company).unwrap().len()
    );

    println!("\n  ✓ All roundtrips correct");
}

// ===========================================================================
// Part 2 — Zero-copy string demonstration
// ===========================================================================

fn section_zerocopy() {
    println!("\n┌─────────────────────────────────────────────────┐");
    println!("│  Part 2: Zero-copy &'de str borrowing            │");
    println!("└─────────────────────────────────────────────────┘");

    // Structs with borrowed string fields (&'de str) get TRUE zero-copy:
    // deserializing never allocates for string fields.
    #[derive(Debug, Deserialize, PartialEq)]
    struct BorrowedUser<'a> {
        id: i64,
        name: &'a str,  // ← zero-copy: borrows from the input bytes
        email: &'a str, // ← zero-copy
        age: i64,
        score: f64,
        active: bool,
        role: &'a str, // ← zero-copy
        city: &'a str, // ← zero-copy
    }

    let user = make_user(0);
    let bytes = to_bin(&user).unwrap();

    let borrowed: BorrowedUser = from_bin(&bytes).unwrap();
    assert_eq!(borrowed.id, user.id);
    assert_eq!(borrowed.name, user.name.as_str());
    assert_eq!(borrowed.role, user.role.as_str());

    println!("  Deserialized BorrowedUser with 4 &str fields");
    println!(
        "  → name = {:?} (borrowed from input buf, 0 allocs)",
        borrowed.name
    );
    println!(
        "  → role = {:?} (borrowed from input buf, 0 allocs)",
        borrowed.role
    );
    println!("  ✓ Zero-copy str fields confirmed");
}

// ===========================================================================
// Part 3 — Binary format layout visualization
// ===========================================================================

fn section_layout() {
    println!("\n┌─────────────────────────────────────────────────┐");
    println!("│  Part 3: Wire format layout                      │");
    println!("└─────────────────────────────────────────────────┘");

    #[derive(Serialize)]
    struct Mini {
        id: i64,
        name: String,
        active: bool,
    }

    let v = Mini {
        id: 42,
        name: "Bob".into(),
        active: true,
    };
    let bytes = to_bin(&v).unwrap();

    print!("  Mini{{id:42, name:\"Bob\", active:true}} → bytes: ");
    for (i, b) in bytes.iter().enumerate() {
        if i > 0 {
            print!(" ");
        }
        print!("{:02x}", b);
    }
    println!();
    println!("  Layout:");
    println!(
        "    [{:02x} {:02x} {:02x} {:02x} {:02x} {:02x} {:02x} {:02x}]  ← i64 42 (LE)",
        bytes[0], bytes[1], bytes[2], bytes[3], bytes[4], bytes[5], bytes[6], bytes[7]
    );
    println!(
        "    [{:02x} {:02x} {:02x} {:02x}]                     ← str len=3 (LE)",
        bytes[8], bytes[9], bytes[10], bytes[11]
    );
    println!(
        "    [{:02x} {:02x} {:02x}]                         ← 'Bob'",
        bytes[12], bytes[13], bytes[14]
    );
    println!(
        "    [{:02x}]                              ← bool true",
        bytes[15]
    );
    println!(
        "  Total: {} bytes (JSON: {} B)",
        bytes.len(),
        serde_json::to_string(&serde_json::json!({"id":42,"name":"Bob","active":true}))
            .unwrap()
            .len()
    );
}

// ===========================================================================
// Part 4 — Performance benchmark vs JSON and text ASON
// ===========================================================================

struct BenchRow {
    name: String,
    json_ser_ms: f64,
    ason_ser_ms: f64,
    bin_ser_ms: f64,
    json_de_ms: f64,
    ason_de_ms: f64,
    bin_de_ms: f64,
    json_bytes: usize,
    ason_bytes: usize,
    bin_bytes: usize,
}

impl BenchRow {
    fn print(&self) {
        let ser_vs_json = self.json_ser_ms / self.bin_ser_ms;
        let ser_vs_ason = self.ason_ser_ms / self.bin_ser_ms;
        let de_vs_json = self.json_de_ms / self.bin_de_ms;
        let de_vs_ason = self.ason_de_ms / self.bin_de_ms;
        let save_vs_json = (1.0 - self.bin_bytes as f64 / self.json_bytes as f64) * 100.0;
        let save_vs_ason = (1.0 - self.bin_bytes as f64 / self.ason_bytes as f64) * 100.0;

        println!("  {}", self.name);
        println!(
            "    Serialize:   JSON {:>8.2}ms | ASON {:>8.2}ms | BIN {:>8.2}ms",
            self.json_ser_ms, self.ason_ser_ms, self.bin_ser_ms
        );
        println!(
            "                 BIN vs JSON: {:.1}x faster  |  BIN vs ASON: {:.1}x faster",
            ser_vs_json, ser_vs_ason
        );
        println!(
            "    Deserialize: JSON {:>8.2}ms | ASON {:>8.2}ms | BIN {:>8.2}ms",
            self.json_de_ms, self.ason_de_ms, self.bin_de_ms
        );
        println!(
            "                 BIN vs JSON: {:.1}x faster  |  BIN vs ASON: {:.1}x faster",
            de_vs_json, de_vs_ason
        );
        println!(
            "    Size:        JSON {:>8} | ASON {:>8} | BIN {:>8}",
            format_bytes(self.json_bytes),
            format_bytes(self.ason_bytes),
            format_bytes(self.bin_bytes)
        );
        println!(
            "                 BIN vs JSON: {:.0}% smaller  |  BIN vs ASON: {:.0}% smaller",
            save_vs_json, save_vs_ason
        );
    }
}

fn bench_users(count: usize, iters: u32) -> BenchRow {
    let users: Vec<User> = (0..count).map(make_user).collect();

    // JSON serialize
    let mut json_str = String::new();
    let t = Instant::now();
    for _ in 0..iters {
        json_str = serde_json::to_string(&users).unwrap();
    }
    let json_ser = t.elapsed();

    // ASON text serialize
    let mut ason_str = String::new();
    let t = Instant::now();
    for _ in 0..iters {
        ason_str = ason::to_string_vec(&users).unwrap();
    }
    let ason_ser = t.elapsed();

    // Binary serialize
    let mut bin_buf = vec![];
    let t = Instant::now();
    for _ in 0..iters {
        bin_buf = to_bin_vec(&users).unwrap();
    }
    let bin_ser = t.elapsed();

    // JSON deserialize
    let t = Instant::now();
    for _ in 0..iters {
        let _: Vec<User> = serde_json::from_str(&json_str).unwrap();
    }
    let json_de = t.elapsed();

    // ASON text deserialize
    let t = Instant::now();
    for _ in 0..iters {
        let _: Vec<User> = ason::from_str_vec(&ason_str).unwrap();
    }
    let ason_de = t.elapsed();

    // Binary deserialize
    let t = Instant::now();
    for _ in 0..iters {
        let _: Vec<User> = from_bin_vec(&bin_buf).unwrap();
    }
    let bin_de = t.elapsed();

    // Verify
    let users2: Vec<User> = from_bin_vec(&bin_buf).unwrap();
    assert_eq!(users, users2);

    BenchRow {
        name: format!("Flat User × {} (8 fields)", count),
        json_ser_ms: json_ser.as_secs_f64() * 1000.0,
        ason_ser_ms: ason_ser.as_secs_f64() * 1000.0,
        bin_ser_ms: bin_ser.as_secs_f64() * 1000.0,
        json_de_ms: json_de.as_secs_f64() * 1000.0,
        ason_de_ms: ason_de.as_secs_f64() * 1000.0,
        bin_de_ms: bin_de.as_secs_f64() * 1000.0,
        json_bytes: json_str.len(),
        ason_bytes: ason_str.len(),
        bin_bytes: bin_buf.len(),
    }
}

fn bench_companies(count: usize, iters: u32) -> BenchRow {
    let companies: Vec<Company> = (0..count).map(make_company).collect();

    let mut json_str = String::new();
    let t = Instant::now();
    for _ in 0..iters {
        json_str = serde_json::to_string(&companies).unwrap();
    }
    let json_ser = t.elapsed();

    let mut ason_strs: Vec<String> = vec![];
    let t = Instant::now();
    for _ in 0..iters {
        ason_strs = companies
            .iter()
            .map(|c| ason::to_string(c).unwrap())
            .collect();
    }
    let ason_ser = t.elapsed();
    let ason_total: String = ason_strs.join("\n");

    let mut bin_buf = vec![];
    let t = Instant::now();
    for _ in 0..iters {
        bin_buf = to_bin_vec(&companies).unwrap();
    }
    let bin_ser = t.elapsed();

    let t = Instant::now();
    for _ in 0..iters {
        let _: Vec<Company> = serde_json::from_str(&json_str).unwrap();
    }
    let json_de = t.elapsed();

    let t = Instant::now();
    for _ in 0..iters {
        for s in &ason_strs {
            let _: Company = ason::from_str(s).unwrap();
        }
    }
    let ason_de = t.elapsed();

    let t = Instant::now();
    for _ in 0..iters {
        let _: Vec<Company> = from_bin_vec(&bin_buf).unwrap();
    }
    let bin_de = t.elapsed();

    let companies2: Vec<Company> = from_bin_vec(&bin_buf).unwrap();
    assert_eq!(companies, companies2);

    BenchRow {
        name: format!("5-level deep Company × {} (~48 nodes each)", count),
        json_ser_ms: json_ser.as_secs_f64() * 1000.0,
        ason_ser_ms: ason_ser.as_secs_f64() * 1000.0,
        bin_ser_ms: bin_ser.as_secs_f64() * 1000.0,
        json_de_ms: json_de.as_secs_f64() * 1000.0,
        ason_de_ms: ason_de.as_secs_f64() * 1000.0,
        bin_de_ms: bin_de.as_secs_f64() * 1000.0,
        json_bytes: json_str.len(),
        ason_bytes: ason_total.len(),
        bin_bytes: bin_buf.len(),
    }
}

fn bench_single_roundtrip(iters: u32) {
    println!("\n  Single struct roundtrip × {} iters", iters);

    let user = make_user(0);

    let t = Instant::now();
    for _ in 0..iters {
        let b = to_bin(&user).unwrap();
        let _: User = from_bin(&b).unwrap();
    }
    let bin_ms = t.elapsed().as_secs_f64() * 1000.0;

    let t = Instant::now();
    for _ in 0..iters {
        let s = serde_json::to_string(&user).unwrap();
        let _: User = serde_json::from_str(&s).unwrap();
    }
    let json_ms = t.elapsed().as_secs_f64() * 1000.0;

    let t = Instant::now();
    for _ in 0..iters {
        let s = ason::to_string(&user).unwrap();
        let _: User = ason::from_str(&s).unwrap();
    }
    let ason_ms = t.elapsed().as_secs_f64() * 1000.0;

    println!("    JSON:  {:>7.2}ms  (1.00x baseline)", json_ms);
    println!(
        "    ASON:  {:>7.2}ms  ({:.2}x vs JSON)",
        ason_ms,
        json_ms / ason_ms
    );
    println!(
        "    BIN:   {:>7.2}ms  ({:.2}x vs JSON, {:.2}x vs ASON)",
        bin_ms,
        json_ms / bin_ms,
        ason_ms / bin_ms
    );
}

fn bench_throughput(iters: u32) {
    println!("\n  Throughput (1000 Users × {} iters):", iters);

    let users: Vec<User> = (0..1000).map(make_user).collect();
    let json_str = serde_json::to_string(&users).unwrap();
    let ason_str = ason::to_string_vec(&users).unwrap();
    let bin_buf = to_bin_vec(&users).unwrap();

    let t = Instant::now();
    for _ in 0..iters {
        let _ = serde_json::to_string(&users).unwrap();
    }
    let json_ser_dur = t.elapsed();

    let t = Instant::now();
    for _ in 0..iters {
        let _ = ason::to_string_vec(&users).unwrap();
    }
    let ason_ser_dur = t.elapsed();

    let t = Instant::now();
    for _ in 0..iters {
        let _ = to_bin_vec(&users).unwrap();
    }
    let bin_ser_dur = t.elapsed();

    let t = Instant::now();
    for _ in 0..iters {
        let _: Vec<User> = serde_json::from_str(&json_str).unwrap();
    }
    let json_de_dur = t.elapsed();

    let t = Instant::now();
    for _ in 0..iters {
        let _: Vec<User> = ason::from_str_vec(&ason_str).unwrap();
    }
    let ason_de_dur = t.elapsed();

    let t = Instant::now();
    for _ in 0..iters {
        let _: Vec<User> = from_bin_vec(&bin_buf).unwrap();
    }
    let bin_de_dur = t.elapsed();

    let total = 1000.0 * iters as f64;
    let json_ser_rps = total / json_ser_dur.as_secs_f64();
    let ason_ser_rps = total / ason_ser_dur.as_secs_f64();
    let bin_ser_rps = total / bin_ser_dur.as_secs_f64();
    let json_de_rps = total / json_de_dur.as_secs_f64();
    let ason_de_rps = total / ason_de_dur.as_secs_f64();
    let bin_de_rps = total / bin_de_dur.as_secs_f64();

    println!("    Serialize throughput (records/sec):");
    println!("      JSON: {:>12.0}  (1.00x)", json_ser_rps);
    println!(
        "      ASON: {:>12.0}  ({:.2}x)",
        ason_ser_rps,
        ason_ser_rps / json_ser_rps
    );
    println!(
        "      BIN:  {:>12.0}  ({:.2}x vs JSON, {:.2}x vs ASON)",
        bin_ser_rps,
        bin_ser_rps / json_ser_rps,
        bin_ser_rps / ason_ser_rps
    );

    println!("    Deserialize throughput (records/sec):");
    println!("      JSON: {:>12.0}  (1.00x)", json_de_rps);
    println!(
        "      ASON: {:>12.0}  ({:.2}x)",
        ason_de_rps,
        ason_de_rps / json_de_rps
    );
    println!(
        "      BIN:  {:>12.0}  ({:.2}x vs JSON, {:.2}x vs ASON)",
        bin_de_rps,
        bin_de_rps / json_de_rps,
        bin_de_rps / ason_de_rps
    );
}

fn section_bench() {
    println!("\n┌─────────────────────────────────────────────────────────────┐");
    println!("│  Part 4: Performance benchmark (Binary vs ASON vs JSON)     │");
    println!("└─────────────────────────────────────────────────────────────┘");

    let iters = 100u32;
    println!("\n  Iterations per test: {}", iters);

    println!("\n  --- Flat struct (8 fields) ---");
    for count in [100, 1000, 5000] {
        let r = bench_users(count, iters);
        r.print();
        println!();
    }

    println!("  --- 5-Level Deep Nesting ---");
    for count in [10, 50, 100] {
        let r = bench_companies(count, iters);
        r.print();
        println!();
    }

    bench_single_roundtrip(10_000);
    bench_throughput(100);
}

// ===========================================================================
// Main
// ===========================================================================

fn main() {
    println!("╔══════════════════════════════════════════════════════════════╗");
    println!("║          ASON Binary Format — Examples & Benchmark          ║");
    println!("╚══════════════════════════════════════════════════════════════╝");
    println!(
        "\nSystem: {} {}",
        std::env::consts::OS,
        std::env::consts::ARCH
    );

    section_basic();
    section_zerocopy();
    section_layout();
    section_bench();

    println!("\n╔══════════════════════════════════════════════════════════════╗");
    println!("║                        Complete                             ║");
    println!("╚══════════════════════════════════════════════════════════════╝");
}
