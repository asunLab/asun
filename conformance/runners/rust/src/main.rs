//! Conformance runner for asun-rs.
//!
//! Reads `conformance/cases.json` and executes every case against
//! `asun::decode::<serde_json::Value>`. Reports pass / fail / error counts
//! and prints details for the first N failures.

use serde::Deserialize;
use std::path::PathBuf;

#[derive(Debug, Deserialize)]
struct Manifest {
    #[allow(dead_code)]
    version: u32,
    count: usize,
    cases: Vec<Case>,
}

#[derive(Debug, Deserialize)]
struct Case {
    id: String,
    #[allow(dead_code)]
    category: String,
    #[allow(dead_code)]
    desc: String,
    input: String,
    kind: String, // "ok" | "error"
    #[serde(default, rename = "schemaDriven")]
    schema_driven: bool,
    #[serde(default)]
    expected: Option<serde_json::Value>,
    #[serde(default, rename = "errorHint")]
    #[allow(dead_code)]
    error_hint: Option<String>,
}

fn manifest_path() -> PathBuf {
    // CARGO_MANIFEST_DIR = .../conformance/runners/rust
    // ../../cases.json   = .../conformance/cases.json
    let mut p = PathBuf::from(env!("CARGO_MANIFEST_DIR"));
    p.pop();
    p.pop();
    p.push("cases.json");
    p
}

fn values_equal(a: &serde_json::Value, b: &serde_json::Value) -> bool {
    use serde_json::Value as V;
    match (a, b) {
        (V::Null, V::Null) => true,
        (V::Bool(x), V::Bool(y)) => x == y,
        (V::String(x), V::String(y)) => x == y,
        (V::Number(x), V::Number(y)) => {
            // Tolerant numeric compare:
            // - if both can be represented as i64, compare as i64
            // - else compare as f64 with tight epsilon
            if let (Some(xi), Some(yi)) = (x.as_i64(), y.as_i64()) {
                return xi == yi;
            }
            if let (Some(xu), Some(yu)) = (x.as_u64(), y.as_u64()) {
                return xu == yu;
            }
            let xf = x.as_f64().unwrap_or(f64::NAN);
            let yf = y.as_f64().unwrap_or(f64::NAN);
            (xf - yf).abs() <= 1e-12 || xf == yf
        }
        (V::Array(x), V::Array(y)) => {
            x.len() == y.len() && x.iter().zip(y).all(|(a, b)| values_equal(a, b))
        }
        (V::Object(x), V::Object(y)) => {
            x.len() == y.len()
                && x.iter()
                    .all(|(k, v)| y.get(k).map(|w| values_equal(v, w)).unwrap_or(false))
        }
        _ => false,
    }
}

#[derive(Default)]
struct Stats {
    total: usize,
    passed: usize,
    failed: usize,
    error_passed: usize,
    error_failed: usize,
    skipped_typed: usize,
}

fn main() {
    let path = manifest_path();
    let raw = std::fs::read_to_string(&path).expect("read cases.json");
    let manifest: Manifest = serde_json::from_str(&raw).expect("parse cases.json");
    eprintln!("loaded {} cases from {}", manifest.count, path.display());

    let mut stats = Stats::default();
    let mut first_failures: Vec<(String, String)> = Vec::new();
    const MAX_REPORT: usize = 25;

    for case in &manifest.cases {
        stats.total += 1;

        // Untyped harness can only handle non-schema-driven inputs.
        // Schema-driven inputs (`{schema}:` / `[{schema}]:`) require a typed
        // Rust target, which the generic runner cannot provide.
        if case.schema_driven {
            stats.skipped_typed += 1;
            continue;
        }

        let result: Result<serde_json::Value, asun::Error> = asun::decode(&case.input);

        let pass = match (case.kind.as_str(), &result) {
            ("ok", Ok(v)) => {
                let expected = case.expected.as_ref().expect("ok case must have expected");
                let eq = values_equal(v, expected);
                if eq {
                    stats.passed += 1;
                } else {
                    stats.failed += 1;
                    if first_failures.len() < MAX_REPORT {
                        first_failures.push((
                            case.id.clone(),
                            format!(
                                "value mismatch\n    input:    {}\n    expected: {}\n    actual:   {}",
                                show(&case.input),
                                serde_json::to_string(expected).unwrap_or_default(),
                                serde_json::to_string(v).unwrap_or_default(),
                            ),
                        ));
                    }
                }
                eq
            }
            ("ok", Err(e)) => {
                stats.failed += 1;
                if first_failures.len() < MAX_REPORT {
                    first_failures.push((
                        case.id.clone(),
                        format!(
                            "expected ok, got error\n    input: {}\n    err:   {}",
                            show(&case.input),
                            e
                        ),
                    ));
                }
                false
            }
            ("error", Err(_)) => {
                stats.error_passed += 1;
                true
            }
            ("error", Ok(v)) => {
                stats.error_failed += 1;
                if first_failures.len() < MAX_REPORT {
                    first_failures.push((
                        case.id.clone(),
                        format!(
                            "expected error, got ok\n    input:    {}\n    accepted: {}",
                            show(&case.input),
                            serde_json::to_string(v).unwrap_or_default(),
                        ),
                    ));
                }
                false
            }
            _ => unreachable!(),
        };
        let _ = pass;
    }

    println!();
    println!("================ ASUN-RS conformance ================");
    println!("total                : {}", stats.total);
    println!("untyped ok-cases pass: {}", stats.passed);
    println!("untyped ok-cases fail: {}", stats.failed);
    println!("untyped err-cases pass: {}", stats.error_passed);
    println!("untyped err-cases fail: {}", stats.error_failed);
    println!("skipped (needs typed): {}", stats.skipped_typed);
    let executed = stats.total - stats.skipped_typed;
    let pass = stats.passed + stats.error_passed;
    let pct = if executed > 0 {
        (pass as f64) / (executed as f64) * 100.0
    } else {
        0.0
    };
    println!(
        "untyped pass rate    : {}/{} ({:.1}%)  [excluding skipped]",
        pass, executed, pct
    );
    println!("=====================================================");

    if !first_failures.is_empty() {
        println!("\nFirst {} failures:\n", first_failures.len());
        for (id, msg) in &first_failures {
            println!("[{id}]\n    {msg}\n");
        }
    }

    // ---------- Encode (round-trip) suite ----------
    let mut enc_path = PathBuf::from(env!("CARGO_MANIFEST_DIR"));
    enc_path.pop();
    enc_path.pop();
    enc_path.push("encode-cases.json");

    let mut enc_failed: usize = 0;
    let mut enc_passed: usize = 0;
    let mut enc_failures: Vec<(String, String)> = Vec::new();

    if let Ok(raw) = std::fs::read_to_string(&enc_path) {
        let m: serde_json::Value = serde_json::from_str(&raw).expect("parse encode-cases.json");
        let cases = m.get("cases").and_then(|c| c.as_array()).expect("cases array");
        eprintln!("loaded {} encode cases from {}", cases.len(), enc_path.display());

        for case in cases {
            let id = case.get("id").and_then(|v| v.as_str()).unwrap_or("?");
            let value = case.get("value").cloned().unwrap_or(serde_json::Value::Null);
            match asun::encode(&value) {
                Ok(text) => match asun::decode::<serde_json::Value>(&text) {
                    Ok(decoded) => {
                        if values_equal(&decoded, &value) {
                            enc_passed += 1;
                        } else {
                            enc_failed += 1;
                            if enc_failures.len() < MAX_REPORT {
                                enc_failures.push((
                                    id.to_string(),
                                    format!(
                                        "round-trip mismatch\n    value:   {}\n    encoded: {}\n    decoded: {}",
                                        serde_json::to_string(&value).unwrap_or_default(),
                                        show(&text),
                                        serde_json::to_string(&decoded).unwrap_or_default(),
                                    ),
                                ));
                            }
                        }
                    }
                    Err(e) => {
                        enc_failed += 1;
                        if enc_failures.len() < MAX_REPORT {
                            enc_failures.push((
                                id.to_string(),
                                format!(
                                    "decode-after-encode failed\n    value:   {}\n    encoded: {}\n    err:     {}",
                                    serde_json::to_string(&value).unwrap_or_default(),
                                    show(&text),
                                    e,
                                ),
                            ));
                        }
                    }
                },
                Err(e) => {
                    enc_failed += 1;
                    if enc_failures.len() < MAX_REPORT {
                        enc_failures.push((
                            id.to_string(),
                            format!(
                                "encode failed\n    value: {}\n    err:   {}",
                                serde_json::to_string(&value).unwrap_or_default(),
                                e,
                            ),
                        ));
                    }
                }
            }
        }

        let total = enc_passed + enc_failed;
        let pct = if total > 0 {
            (enc_passed as f64) / (total as f64) * 100.0
        } else {
            0.0
        };
        println!();
        println!("============ ASUN-RS encode round-trip ==============");
        println!("total : {}", total);
        println!("pass  : {}", enc_passed);
        println!("fail  : {}", enc_failed);
        println!("rate  : {}/{} ({:.1}%)", enc_passed, total, pct);
        println!("=====================================================");
        if !enc_failures.is_empty() {
            println!("\nFirst {} encode failures:\n", enc_failures.len());
            for (id, msg) in &enc_failures {
                println!("[{id}]\n    {msg}\n");
            }
        }
    }

    if stats.failed > 0 || stats.error_failed > 0 || enc_failed > 0 {
        std::process::exit(1);
    }
}

fn show(s: &str) -> String {
    if s.len() > 120 {
        format!("{}…(+{} bytes)", &s[..120].replace('\n', "\\n"), s.len() - 120)
    } else {
        s.replace('\n', "\\n")
    }
}
