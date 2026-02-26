# ASON Java — High-Performance Array-Schema Object Notation

A zero-copy, SIMD-accelerated ASON serialization library for Java 25+.
Benchmarked against **Gson** (Google), the most widely-used JSON library in the JVM ecosystem.

## Features

- **ClassMeta + MethodHandle invokeExact**: Pre-computed per-class metadata with pre-adapted MethodHandles. Primitive fields use type-specific `invokeExact` (e.g., `(Object)->int`) with zero boxing/adaptation overhead
- **SIMD Acceleration**: Uses `jdk.incubator.vector` (ByteVector 256-bit/128-bit) for fast character scanning — special char detection, escape scanning, delimiter search
- **ThreadLocal Buffer Pool**: `ByteBuffer` with ThreadLocal reuse (up to 1MB), 2× growth factor — eliminates allocation for repeated encode calls
- **ISO-8859-1 Fast String Construction**: Both encoder and decoder detect ASCII-only content and use `ISO_8859_1` charset for direct byte-copy String construction (vs UTF-8 validation overhead)
- **Single-Pass Fused Write**: `writeString` combines classify + write in one pass with rollback on special chars — no separate analysis pass
- **CHAR_CLASS Lookup Table**: 128-byte classification table replaces 6+ per-character comparisons for ASON delimiter detection
- **POW10 Direct Double Parsing**: Decoder parses `intPart + fracVal / POW10[fracDigits]` avoiding `String` allocation for simple decimals
- **Type-Tag Switch Dispatch**: Integer-tagged field types (T_BOOLEAN=0 .. T_STRUCT=12) for O(1) dispatch in both encode and decode
- **DEC_DIGITS Fast Formatting**: 200-byte lookup table for two-digit integer formatting, eliminates division-per-digit overhead
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

// byte[] decode (avoids String→byte[] conversion)
T obj = Ason.decode(bytes, MyClass.class);
List<T> list = Ason.decodeList(bytes, MyClass.class);

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

## Performance vs Gson

Benchmarked on JDK 25 (aarch64) against Gson 2.12.1 — the most widely-used JSON library in Java.
Ratio > 1.0 means ASON is faster.

### Serialization (ASON 4–5x faster)

| Test               | JSON (Gson) | ASON    | Ratio       |
| ------------------ | ----------- | ------- | ----------- |
| Flat 100×          | 6.81ms      | 1.52ms  | **4.48x** ✓ |
| Flat 500×          | 38.52ms     | 8.08ms  | **4.77x** ✓ |
| Flat 1000×         | 70.46ms     | 14.13ms | **4.99x** ✓ |
| Flat 5000×         | 355.77ms    | 71.44ms | **4.98x** ✓ |
| Deep 10×           | 27.52ms     | 4.93ms  | **5.58x** ✓ |
| Deep 50×           | 134.50ms    | 24.82ms | **5.42x** ✓ |
| Deep 100×          | 269.97ms    | 48.73ms | **5.54x** ✓ |
| Single Flat 10000× | 18.20ms     | 10.98ms | **1.66x** ✓ |

### Deserialization (ASON 1.7–2.1x faster)

| Test       | JSON (Gson) | ASON     | Ratio       |
| ---------- | ----------- | -------- | ----------- |
| Flat 100×  | 5.70ms      | 2.95ms   | **1.93x** ✓ |
| Flat 500×  | 29.20ms     | 13.83ms  | **2.11x** ✓ |
| Flat 1000× | 60.38ms     | 32.11ms  | **1.88x** ✓ |
| Flat 5000× | 286.69ms    | 143.87ms | **1.99x** ✓ |
| Deep 10×   | 25.00ms     | 14.92ms  | **1.68x** ✓ |
| Deep 50×   | 138.94ms    | 77.25ms  | **1.80x** ✓ |
| Deep 100×  | 246.83ms    | 138.38ms | **1.78x** ✓ |

### Throughput (1000 records × 100 iterations)

| Direction   | JSON (Gson)      | ASON            | Ratio       |
| ----------- | ---------------- | --------------- | ----------- |
| Serialize   | ~1.4M records/s  | ~7.1M records/s | **4.90x** ✓ |
| Deserialize | ~1.8M records/s  | ~3.7M records/s | **2.08x** ✓ |

### Size Reduction

| Data      | JSON   | ASON Text | Saving  | ASON Binary | Saving  |
| --------- | ------ | --------- | ------- | ----------- | ------- |
| Flat 1000 | 122 KB | 57 KB     | **53%** | 74 KB       | **39%** |
| Deep 100  | 438 KB | 170 KB    | **61%** | 225 KB      | **49%** |

## Why ASON Beats Gson

Gson uses Java reflection for field access and tree-based intermediate representations. ASON achieves 2–5x better performance through format advantages and JIT-friendly design:

1. **No Key Repetition**: JSON repeats every key for every object. ASON writes the schema once, then only values — 53% smaller output means less memory bandwidth.
2. **No Quoting Overhead**: ASON only quotes strings that contain special characters — most strings are emitted raw, saving `"` delimiter overhead.
3. **MethodHandle invokeExact**: Pre-adapted handles match JIT-optimized direct field access. No boxing for primitives, no type adaptation at call site — vs Gson's reflection-based `Field.get()`/`Field.set()`.
4. **SIMD Scanning**: Character classification uses 256-bit vector operations, processing 32 bytes per cycle.
5. **ThreadLocal Buffer Pool**: Reuses up to 1MB byte buffers, eliminating allocation pressure in the encode hot path.
6. **ISO-8859-1 String Fast Path**: ASCII-only strings (the common case) use `ISO_8859_1` charset for direct byte-copy construction — 2-3x faster than UTF-8 validation.
7. **Direct Double Parsing**: POW10 lookup table parses `intPart + fracVal / 10^n` without String allocation — avoids `Double.parseDouble()` for simple decimals.
8. **No Intermediate Tree**: Gson builds `JsonElement` trees during parsing. ASON decodes directly into target objects with zero intermediate allocations.

## Supported Types

| Java Type                      | ASON Text           | ASON Binary          |
| ------------------------------ | ------------------- | -------------------- |
| `boolean`                      | `true`/`false`      | 1 byte (0/1)         |
| `int`, `long`, `short`, `byte` | decimal             | 4/8/2/1 bytes LE     |
| `float`, `double`              | decimal             | 4/8 bytes IEEE754 LE |
| `String`                       | plain or `"quoted"` | u32 length + UTF-8   |
| `char`                         | single char string  | 2 bytes LE           |
| `Optional<T>`                  | value or empty      | u8 tag + payload     |
| `List<T>`                      | `[v1,v2,...]`       | u32 count + elements |
| `Map<K,V>`                     | `[(k1,v1),(k2,v2)]` | u32 count + pairs    |
| Nested struct                  | `(f1,f2,...)`       | fields in order      |

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
├── Ason.java          — Public API + text encoder (CHAR_CLASS, DEC_DIGITS, single-pass writeString)
├── ClassMeta.java     — Per-class metadata cache (FieldMeta, MethodHandle invokeExact, type tags)
├── AsonDecoder.java   — SIMD-accelerated text decoder (POW10, skipWs, ISO-8859-1 fast path)
├── AsonBinary.java    — Binary codec (LE wire format)
├── ByteBuffer.java    — ThreadLocal byte buffer pool (1MB max, 2× growth, ASCII tracking)
├── SimdUtils.java     — SIMD utilities (ByteVector 256/128)
├── AsonException.java — Runtime exception
└── examples/
    ├── BasicExample.java    — 12 basic examples
    ├── ComplexExample.java  — 14 complex examples
    └── BenchExample.java    — Full benchmark suite (vs Gson)
```
