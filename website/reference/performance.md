# Performance

All benchmarks were run on **Apple M-series** hardware with a `--release` build (Rust). Results for other languages are proportional.

## Summary

| Metric | ASON Text vs JSON | ASON-BIN vs JSON |
|--------|------------------|-----------------|
| Serialize speed | **2–4×** faster | **7–10×** faster |
| Deserialize speed | **1.2–2.5×** faster | **2–2.5×** faster |
| Payload size | **53–61%** smaller | **38–49%** smaller |

## Flat Struct (8 fields, × 1 000 iterations)

```
struct User { id, name, email, age, active, score, role, created_at }
```

| Test | JSON | ASON Text | ASON-BIN | Text vs JSON | BIN vs JSON |
|------|------|-----------|----------|-------------|-------------|
| Serialize | 2.19 ms | 1.07 ms | **0.28 ms** | 2.0× | **7.7×** |
| Deserialize | 6.05 ms | 5.10 ms | **2.96 ms** | 1.2× | **2.0×** |
| Payload | 121 675 B | 56 716 B | **74 454 B** | 53% smaller | 39% smaller |

## Deep Struct (5-level nested, × 100 iterations)

```
Organization > Department > Team > Member > Contact
```

| Test | JSON | ASON Text | ASON-BIN | Text vs JSON | BIN vs JSON |
|------|------|-----------|----------|-------------|-------------|
| Serialize | 4.19 ms | 2.67 ms | **0.50 ms** | 1.6× | **8.4×** |
| Deserialize | 9.37 ms | 8.55 ms | **3.72 ms** | 1.1× | **2.5×** |
| Payload | 438 112 B | 174 611 B | **225 434 B** | 60% smaller | 49% smaller |

## Single Roundtrip (User, × 100 000)

| Format | Time/op | vs JSON |
|--------|---------|---------|
| ASON-BIN | ~182 ns | **2.1× faster** |
| JSON | ~375 ns | — |
| ASON Text | ~552 ns | 0.68× |

> ASON Text single-struct is slower than JSON in the roundtrip benchmark because the schema/data split adds overhead for single objects. ASON Text's advantage appears with **arrays of 10+ rows**.

## Why ASON-BIN Is Fast

### 1. No dynamic type inference
The binary layout mirrors the struct declaration order. The deserializer calls typed decode functions  directly — no lookahead, no branching on type.

### 2. Zero key-hashing
Fields are positionally indexed. No string reads, no hash computation, no map lookups.

### 3. Length-prefixed strings
String boundaries are known before reading — no scanning for delimiters.

### 4. Zero-copy deserialize
String fields with `'de` lifetime borrow directly from the input buffer — no `String` allocation.

### 5. SIMD string copy
Strings ≥ 32 bytes are bulk-copied in 16-byte SIMD chunks:
- **ARM**: NEON
- **x86-64**: SSE2

## Running Benchmarks Yourself

```bash
# Rust
cd rust
cargo run --release --example bench

# Go
cd go
go test -bench=. -benchmem ./...

# C
cd c && cmake -B build -DCMAKE_BUILD_TYPE=Release && cmake --build build
./build/examples/bench
```
