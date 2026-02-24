# ASON Java — 高性能数组模式对象表示法 (High-Performance Array-Schema Object Notation)

零拷贝、SIMD 加速的 ASON 序列化库，适用于 Java 25+。
基准测试对比对象为 **Gson** (Google)，它是 JVM 生态中使用最广泛的 JSON 库。

## 特性

- **ClassMeta + MethodHandle invokeExact**: 预计算每类元数据并使用预适配的 MethodHandle。原始类型字段使用特定类型的 `invokeExact`（例如 `(Object)->int`），实现零装箱/适配开销。
- **SIMD 加速**: 使用 `jdk.incubator.vector` (ByteVector 256位/128位) 进行快速字符扫描 —— 特殊字符检测、转义扫描、分隔符搜索。
- **ThreadLocal 缓冲池**: `ByteBuffer` 使用 ThreadLocal 复用（最高 1MB），2 倍增长因子 —— 消除重复编码调用的分配开销。
- **ISO-8859-1 快速字符串构建**: 编码器和解码器均可检测仅含 ASCII 的内容，并使用 `ISO_8859_1` 字符集直接进行字节拷贝构建字符串（对比 UTF-8 校验开销）。
- **单阶段融合写入 (Single-Pass Fused Write)**: `writeString` 在单次扫描中结合了分类和写入，遇到特殊字符时可回滚 —— 无需单独的分析阶段。
- **CHAR_CLASS 查找表**: 128 字节分类表取代了 ASON 分隔符检测中每字符 6 次以上的比较。
- **POW10 直接浮点解析**: 解码器解析 `整数部分 + 小数部分 / POW10[小数位数]`，避免了简单浮点数的 `String` 分配。
- **类型标签 Switch 调度**: 整数标记的字段类型 (T_BOOLEAN=0 .. T_STRUCT=12)，在编码和解码中实现 O(1) 调度。
- **DEC_DIGITS 快速格式化**: 200 字节查找表实现两位数整数格式化，消除每位除法开销。
- **二进制格式**: 小端序线格式，零解析开销 —— 原始类型直接内存读取。
- **模式驱动**: `{field1,field2}:(val1,val2)` 格式消除数组中重复键名。

## API

```java
import io.ason.Ason;

// 文本编解码
String text = Ason.encode(obj);                    // 无类型模式
String typed = Ason.encodeTyped(obj);              // 带类型标注
T obj = Ason.decode(text, MyClass.class);          // 单个结构体
List<T> list = Ason.decodeList(text, MyClass.class); // 结构体列表

// byte[] 解码 (避免 String→byte[] 转换)
T obj = Ason.decode(bytes, MyClass.class);
List<T> list = Ason.decodeList(bytes, MyClass.class);

// 二进制编解码
byte[] bin = Ason.encodeBinary(obj);
T obj = Ason.decodeBinary(bin, MyClass.class);
List<T> list = Ason.decodeBinaryList(bin, MyClass.class);
```

## ASON 格式

单个结构体：

```text
{name,age,active}:(Alice,30,true)
```

数组（模式驱动 —— 模式只写一次）：

```text
[{name,age,active}]:(Alice,30,true),(Bob,25,false)
```

带类型模式：

```text
{name:str,age:int,active:bool}:(Alice,30,true)
```

嵌套结构体：

```text
{id,dept}:(1,(Engineering))
```

## 性能对比 FastJSON2

在 JDK 25 (aarch64) 上对比 FastJSON2 2.0.53 —— Java 生态中最快的 JSON 库。
倍率 > 1.0 表示 ASON 更快。

### 序列化（ASON 在所有规模下胜出）

| 测试            | JSON (FastJSON2) | ASON    | 倍率        |
| --------------- | ---------------- | ------- | ----------- |
| 平面结构 100×   | 1.10ms           | 0.91ms  | **1.22x** ✓ |
| 平面结构 500×   | 4.91ms           | 4.76ms  | **1.03x** ✓ |
| 平面结构 1000×  | 18.35ms          | 8.69ms  | **2.11x** ✓ |
| 平面结构 5000×  | 56.69ms          | 42.22ms | **1.34x** ✓ |
| 深层嵌套 10×    | 4.26ms           | 3.64ms  | **1.17x** ✓ |
| 深层嵌套 50×    | 17.65ms          | 15.49ms | **1.14x** ✓ |
| 深层嵌套 100×   | 32.60ms          | 29.36ms | **1.11x** ✓ |
| 单个平面 10000× | 9.53ms           | 4.50ms  | **2.12x** ✓ |

### 反序列化（与 FastJSON2 持平）

| 测试           | JSON (FastJSON2) | ASON    | 倍率        |
| -------------- | ---------------- | ------- | ----------- |
| 平面结构 100×  | 1.31ms           | 1.25ms  | **1.05x** ✓ |
| 平面结构 500×  | 6.42ms           | 6.73ms  | 0.95x       |
| 平面结构 1000× | 9.59ms           | 10.51ms | 0.91x       |
| 深层嵌套 50×   | 16.63ms          | 17.89ms | 0.93x       |
| 深层嵌套 100×  | 34.15ms          | 33.89ms | **1.01x** ✓ |

### 吞吐量（1000 条记录 × 100 次迭代）

| 方向     | JSON (FastJSON2) | ASON           | 倍率                       |
| -------- | ---------------- | -------------- | -------------------------- |
| 序列化   | ~13M 条记录/s    | ~9M 条记录/s   | 0.7x (ASON 输出缩小了 53%) |
| 反序列化 | ~9.3M 条记录/s   | ~9.6M 条记录/s | **1.03x** ✓                |

### 体积缩减

| 数据      | JSON   | ASON 文本 | 节省    | ASON 二进制 | 节省    |
| --------- | ------ | --------- | ------- | ----------- | ------- |
| 平面 1000 | 121 KB | 55 KB     | **53%** | 72 KB       | **39%** |
| 深层 100  | 427 KB | 166 KB    | **61%** | 220 KB      | **49%** |

## 为什么 ASON 比 Gson 更快

Gson 使用 Java 反射进行字段访问并构建基于树的中间表示。ASON 通过格式优势和 JIT 友好设计实现了 2–5 倍的性能提升：

1. **无键名重复**: JSON 为每个对象重复写入键名。ASON 只写一次模式，之后仅有值 — 输出缩小 53% 意味着更低的内存带宽占用。
2. **无引号开销**: ASON 仅对包含特殊字符的字符串加引号 — 大多数字符串直接输出，节省了 `"` 分隔符开销。
3. **MethodHandle invokeExact**: 预适配的句柄匹配 JIT 优化的直接字段访问。原始类型无装箱，调用点无需类型适配 — 对比 Gson 基于反射的 `Field.get()`/`Field.set()`。
4. **SIMD 扫描**: 字符分类使用 256 位向量操作，每周期处理 32 字节。
5. **ThreadLocal 缓冲池**: 复用高达 1MB 的字节缓冲，消除编码热点路径的分配压力。
6. **ISO-8859-1 字符串快径**: 仅含 ASCII 的字符串（常见情况）使用 `ISO_8859_1` 字符集直接进行字节拷贝构建 — 比 UTF-8 校验快 2-3 倍。
7. **直接浮点解析**: POW10 查找表通过 `整数部分 + 小数部分 / 10^n` 进行解析，无需 String 分配 — 避免了简单浮点数的 `Double.parseDouble()` 开销。
8. **无中间树**: Gson 在解析时构建 `JsonElement` 树。ASON 直接解码到目标对象，零中间分配。

## 支持类型

| Java 类型                      | ASON 文本           | ASON 二进制         |
| ------------------------------ | ------------------- | ------------------- |
| `boolean`                      | `true`/`false`      | 1 字节 (0/1)        |
| `int`, `long`, `short`, `byte` | 十进制              | 4/8/2/1 字节 LE     |
| `float`, `double`              | 十进制              | 4/8 字节 IEEE754 LE |
| `String`                       | 原文或 `"带引号"`   | u32 长度 + UTF-8    |
| `char`                         | 单字符字符串        | 2 字节 LE           |
| `Optional<T>`                  | 值或空              | u8 标记 + 载荷      |
| `List<T>`                      | `[v1,v2,...]`       | u32 计数 + 元素     |
| `Map<K,V>`                     | `[(k1,v1),(k2,v2)]` | u32 计数 + 键值对   |
| 嵌套结构体                     | `(f1,f2,...)`       | 按字段顺序          |

## 构建与运行

```bash
# 要求：JDK 25+，Gradle 9+
./gradlew test
./gradlew runBasicExample
./gradlew runComplexExample
./gradlew runBenchExample
```

## 项目结构

```text
src/main/java/io/ason/
├── Ason.java          — 公共 API + 文本编码器 (CHAR_CLASS, DEC_DIGITS, 单阶段 writeString)
├── ClassMeta.java     — 每类元数据缓存 (FieldMeta, MethodHandle invokeExact, 类型标签)
├── AsonDecoder.java   — SIMD 加速文本解码器 (POW10, skipWs, ISO-8859-1 快径)
├── AsonBinary.java    — 二进制编解码器 (LE 线格式)
├── ByteBuffer.java    — ThreadLocal 字节缓冲池 (最大 1MB, 2倍增长, ASCII 追踪)
├── SimdUtils.java     — SIMD 工具 (ByteVector 256/128)
├── AsonException.java — 运行时异常
└── examples/
    ├── BasicExample.java    — 12 个基础示例
    ├── ComplexExample.java  — 14 个复杂示例
    └── BenchExample.java    — 完整基准测试套件 (对比 Gson)
```
