# ASON Java — High-Performance Array-Schema Object Notation

A zero-copy, SIMD-accelerated ASON serialization library for Java 25+.

## Features

- **SIMD Acceleration**: Uses `jdk.incubator.vector` (ByteVector 256-bit/128-bit) for fast character scanning — special char detection, whitespace skipping, escape scanning, delimiter search
- **Zero-Copy Architecture**: Custom `ByteBuffer` writes directly to byte arrays, avoids `StringBuilder` overhead and intermediate String conversions
- **DEC_DIGITS Fast Formatting**: 200-byte lookup table for two-digit integer formatting (same as Rust), eliminates division-per-digit overhead
- **Reflection Cache**: `WeakHashMap<Class<?>, Field[]>` with one-time field resolution per class, skips static/transient/synthetic fields
- **Binary Format**: Little-endian wire format with zero parsing overhead — direct memory reads for primitives
- **Schema-Driven**: `{field1,field2}:(val1,val2)` format eliminates redundant key repetition in arrays

## API

```java
import io.ason.Ason;

// Text encode/decode
String text = Ason.encode(obj);                    // untyped schema
String typed = Ason.encodeTyped(obj);              // with type annotations
T obj = Ason.decode(text, MyClass.class);          // single struct
List<T> list = Ason.decodeList(text, MyClass.class); // list of structs

// Binary encode/decode
byte[] bin = Ason.encodeBinary(obj);
T obj = Ason.decodeBinary(bin, MyClass.class);
List<T> list = Ason.decodeBinaryList(bin, MyClass.class);
```

## ASON Format

Single struct:
```
{name,age,active}:(Alice,30,true)
```

Array (schema-driven — schema written once):
```
[{name,age,active}]:(Alice,30,true),(Bob,25,false)
```

Typed schema:
```
{name:str,age:int,active:bool}:(Alice,30,true)
```

Nested struct:
```
{id,dept}:(1,(Engineering))
```

## Performance Advantages

### vs JSON (Gson) — Benchmark Results

| Test | Metric | JSON (Gson) | ASON | Ratio |
|------|--------|-------------|------|-------|
| Flat 1000× | Serialize | 48.56ms | 20.08ms | **2.42x faster** |
| Flat 1000× | Deserialize | 41.28ms | 26.02ms | **1.59x faster** |
| Flat 5000× | Serialize | 215.06ms | 95.74ms | **2.25x faster** |
| Flat 5000× | Deserialize | 200.89ms | 118.11ms | **1.70x faster** |
| Deep 100× | Serialize | 161.21ms | 70.78ms | **2.28x faster** |
| Deep 100× | Deserialize | 173.04ms | 97.69ms | **1.77x faster** |
| Single Flat | Roundtrip 10000× | 11.03ms | 5.86ms | **1.88x faster** |
| Single Deep | Roundtrip 10000× | 346.06ms | 151.55ms | **2.28x faster** |

### Throughput (1000 records × 100 iterations)

| Direction | JSON (Gson) | ASON | Speedup |
|-----------|-------------|------|---------|
| Serialize | 2.4M records/s | 5.3M records/s | **2.21x** |
| Deserialize | 2.5M records/s | 4.3M records/s | **1.73x** |

### Size Reduction

| Data | JSON | ASON Text | Saving | ASON Binary | Saving |
|------|------|-----------|--------|-------------|--------|
| Flat 1000 | 121 KB | 55 KB | **53%** | 72 KB | **39%** |
| Deep 100 | 427 KB | 166 KB | **61%** | 220 KB | **49%** |

### Binary Format Performance

| Test | Serialize | Deserialize |
|------|-----------|-------------|
| Flat 5000× | 55.35ms | 52.74ms |
| Deep 100× | 37.59ms | 40.39ms |

Binary format is the fastest option — direct LE memory writes with zero text parsing.

## Why ASON is Faster

1. **No Key Repetition**: JSON repeats every key for every object. ASON writes the schema once, then only values.
2. **No Quoting Overhead**: ASON only quotes strings that contain special characters — most strings are emitted raw.
3. **SIMD Scanning**: Character classification uses 256-bit vector operations, processing 32 bytes per cycle.
4. **Zero-Copy Buffer**: Direct byte array writes with amortized growth, no intermediate StringBuilder allocation.
5. **Fast Integer Path**: DEC_DIGITS lookup table emits two digits per table access, halving the division count.
6. **Float Fast Path**: Integer-valued floats (e.g., `95.0`) skip `Double.toString()` entirely.

## Supported Types

| Java Type | ASON Text | ASON Binary |
|-----------|-----------|-------------|
| `boolean` | `true`/`false` | 1 byte (0/1) |
| `int`, `long`, `short`, `byte` | decimal | 4/8/2/1 bytes LE |
| `float`, `double` | decimal | 4/8 bytes IEEE754 LE |
| `String` | plain or `"quoted"` | u32 length + UTF-8 |
| `char` | single char string | 2 bytes LE |
| `Optional<T>` | value or empty | u8 tag + payload |
| `List<T>` | `[v1,v2,...]` | u32 count + elements |
| `Map<K,V>` | `[(k1,v1),(k2,v2)]` | u32 count + pairs |
| Nested struct | `(f1,f2,...)` | fields in order |

## Build & Run

```bash
# Requirements: JDK 25+, Gradle 9+
./gradlew test
./gradlew runBasicExample
./gradlew runComplexExample
./gradlew runBenchExample
```

## Project Structure

```
src/main/java/io/ason/
├── Ason.java          — Public API (encode/decode/encodeBinary/decodeBinary)
├── AsonDecoder.java   — SIMD-accelerated text decoder
├── AsonBinary.java    — Binary codec (LE wire format)
├── ByteBuffer.java    — Zero-copy byte buffer
├── SimdUtils.java     — SIMD utilities (ByteVector 256/128)
├── AsonException.java — Runtime exception
└── examples/
    ├── BasicExample.java    — 12 basic examples
    ├── ComplexExample.java  — 14 complex examples
    └── BenchExample.java    — Full benchmark suite
```
