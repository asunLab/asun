# ASON Java — 高性能数组模式对象表示法

零拷贝、SIMD 加速的 Java ASON 序列化库，要求 Java 25+。

## 特性

- **SIMD 加速**：使用 `jdk.incubator.vector`（ByteVector 256位/128位）进行快速字符扫描 —— 特殊字符检测、空白跳过、转义扫描、分隔符搜索
- **零拷贝架构**：自定义 `ByteBuffer` 直接写入字节数组，避免 `StringBuilder` 开销和中间层字符串转换
- **DEC_DIGITS 快速格式化**：200字节查找表实现两位数整数格式化（与 Rust 版本一致），消除逐位除法开销
- **反射缓存**：`WeakHashMap<Class<?>, Field[]>` 每类一次性字段解析，跳过 static/transient/synthetic 字段
- **二进制格式**：小端序线格式，零解析开销 —— 原始类型直接内存读取
- **模式驱动**：`{field1,field2}:(val1,val2)` 格式消除数组中重复键名

## API

```java
import io.ason.Ason;

// 文本编解码
String text = Ason.encode(obj);                    // 无类型模式
String typed = Ason.encodeTyped(obj);              // 带类型标注
T obj = Ason.decode(text, MyClass.class);          // 单个结构体
List<T> list = Ason.decodeList(text, MyClass.class); // 结构体列表

// 二进制编解码
byte[] bin = Ason.encodeBinary(obj);
T obj = Ason.decodeBinary(bin, MyClass.class);
List<T> list = Ason.decodeBinaryList(bin, MyClass.class);
```

## ASON 格式

单个结构体：
```
{name,age,active}:(Alice,30,true)
```

数组（模式驱动 —— 模式只写一次）：
```
[{name,age,active}]:(Alice,30,true),(Bob,25,false)
```

带类型模式：
```
{name:str,age:int,active:bool}:(Alice,30,true)
```

嵌套结构体：
```
{id,dept}:(1,(Engineering))
```

## 性能优势

### 对比 JSON (Gson) — 基准测试结果

| 测试 | 指标 | JSON (Gson) | ASON | 倍率 |
|------|------|-------------|------|------|
| 平面结构 1000× | 序列化 | 48.56ms | 20.08ms | **快 2.42 倍** |
| 平面结构 1000× | 反序列化 | 41.28ms | 26.02ms | **快 1.59 倍** |
| 平面结构 5000× | 序列化 | 215.06ms | 95.74ms | **快 2.25 倍** |
| 平面结构 5000× | 反序列化 | 200.89ms | 118.11ms | **快 1.70 倍** |
| 深层嵌套 100× | 序列化 | 161.21ms | 70.78ms | **快 2.28 倍** |
| 深层嵌套 100× | 反序列化 | 173.04ms | 97.69ms | **快 1.77 倍** |
| 单个平面 | 往返 10000× | 11.03ms | 5.86ms | **快 1.88 倍** |
| 单个深层 | 往返 10000× | 346.06ms | 151.55ms | **快 2.28 倍** |

### 吞吐量（1000 条记录 × 100 次迭代）

| 方向 | JSON (Gson) | ASON | 加速比 |
|------|-------------|------|--------|
| 序列化 | 240万 条/秒 | 530万 条/秒 | **2.21 倍** |
| 反序列化 | 250万 条/秒 | 430万 条/秒 | **1.73 倍** |

### 体积缩减

| 数据 | JSON | ASON 文本 | 节省 | ASON 二进制 | 节省 |
|------|------|-----------|------|-------------|------|
| 平面 1000 | 121 KB | 55 KB | **53%** | 72 KB | **39%** |
| 深层 100 | 427 KB | 166 KB | **61%** | 220 KB | **49%** |

### 二进制格式性能

| 测试 | 序列化 | 反序列化 |
|------|--------|----------|
| 平面 5000× | 55.35ms | 52.74ms |
| 深层 100× | 37.59ms | 40.39ms |

二进制格式是最快选项 —— 直接小端序内存写入，零文本解析。

## 为什么 ASON 更快

1. **无键名重复**：JSON 为每个对象重复写入所有键名。ASON 只写一次模式，之后只有值数据。
2. **无引号开销**：ASON 只对包含特殊字符的字符串加引号 —— 大多数字符串直接输出。
3. **SIMD 扫描**：字符分类使用 256 位向量操作，每周期处理 32 字节。
4. **零拷贝缓冲**：直接字节数组写入，摊销增长，无中间 StringBuilder 分配。
5. **快速整数路径**：DEC_DIGITS 查找表每次表访问输出两位数字，除法次数减半。
6. **浮点快速路径**：整数值浮点数（如 `95.0`）完全跳过 `Double.toString()`。

## 支持类型

| Java 类型 | ASON 文本 | ASON 二进制 |
|-----------|-----------|-------------|
| `boolean` | `true`/`false` | 1 字节 (0/1) |
| `int`, `long`, `short`, `byte` | 十进制 | 4/8/2/1 字节 LE |
| `float`, `double` | 十进制 | 4/8 字节 IEEE754 LE |
| `String` | 原文或 `"带引号"` | u32 长度 + UTF-8 |
| `char` | 单字符字符串 | 2 字节 LE |
| `Optional<T>` | 值或空 | u8 标记 + 载荷 |
| `List<T>` | `[v1,v2,...]` | u32 计数 + 元素 |
| `Map<K,V>` | `[(k1,v1),(k2,v2)]` | u32 计数 + 键值对 |
| 嵌套结构体 | `(f1,f2,...)` | 按字段顺序 |

## 构建与运行

```bash
# 要求：JDK 25+，Gradle 9+
./gradlew test
./gradlew runBasicExample
./gradlew runComplexExample
./gradlew runBenchExample
```

## 项目结构

```
src/main/java/io/ason/
├── Ason.java          — 公共 API（encode/decode/encodeBinary/decodeBinary）
├── AsonDecoder.java   — SIMD 加速文本解码器
├── AsonBinary.java    — 二进制编解码器（LE 线格式）
├── ByteBuffer.java    — 零拷贝字节缓冲
├── SimdUtils.java     — SIMD 工具（ByteVector 256/128）
├── AsonException.java — 运行时异常
└── examples/
    ├── BasicExample.java    — 12 个基础示例
    ├── ComplexExample.java  — 14 个复杂示例
    └── BenchExample.java    — 完整基准测试套件
```
