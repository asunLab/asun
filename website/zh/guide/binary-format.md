# 二进制格式（ASON-BIN）

ASON-BIN 是针对任意 `serde` 兼容类型的紧凑二进制编码，是速度最快的 ASON 格式，非常适合内部服务通信、缓存和持久化存储。

## 使用时机

| 场景 | 推荐格式 |
|------|----------|
| LLM 提示词 / 响应 | ASON Text |
| 人类可读配置 / 日志 | ASON Text |
| 内部微服务 RPC | **ASON-BIN** |
| Redis / 磁盘缓存 | **ASON-BIN** |
| 高频数据管道 | **ASON-BIN** |

## 性能数据

**平面结构体（8 字段，× 1000 次）**

| 测试项 | JSON | ASON Text | ASON-BIN | BIN vs JSON |
|--------|------|-----------|----------|-------------|
| 序列化 | 2.19 ms | 1.07 ms | **0.28 ms** | **7.7×** |
| 反序列化 | 6.05 ms | 5.10 ms | **2.96 ms** | **2.0×** |
| 载荷体积 | 121 675 B | 56 716 B | **74 454 B** | **39% 更小** |

**深层嵌套结构体（5 层，× 100 次）**

| 测试项 | JSON | ASON Text | ASON-BIN | BIN vs JSON |
|--------|------|-----------|----------|-------------|
| 序列化 | 4.19 ms | 2.67 ms | **0.50 ms** | **8.4×** |
| 反序列化 | 9.37 ms | 8.55 ms | **3.72 ms** | **2.5×** |
| 载荷体积 | 438 112 B | 174 611 B | **225 434 B** | **49% 更小** |

## 线格式

所有整数采用**小端序**，字符串采用长度前缀编码。

| 类型 | 编码 |
|------|------|
| `bool` | 1 字节（0 / 1） |
| `i8` / `u8` | 1 字节 |
| `i16` / `u16` | 2 字节 LE |
| `i32` / `u32` | 4 字节 LE |
| `i64` / `u64` | 8 字节 LE |
| `f32` | 4 字节 IEEE 754 LE |
| `f64` | 8 字节 IEEE 754 LE |
| `str` / `String` | `[u32 长度][UTF-8 字节]` |
| `Option<T>` | `[u8: 0=None / 1=Some][载荷]` |
| `Vec<T>` / 序列 | `[u32 个数][元素…]` |
| `struct` | 字段按声明顺序，无前缀 |
| `enum` | `[u32 变体索引][载荷]` |
| `unit` | 0 字节 |

## 零拷贝反序列化

当结构体字段使用 `'de` 生命周期借用输入缓冲区时，`from_bin` 直接返回指向输入字节的字符串切片 —— 无堆分配：

```rust
#[derive(Deserialize)]
struct BorrowedUser<'a> {
    id: i64,
    name: &'a str,  // 直接指向输入缓冲区，零 String 分配
    active: bool,
}

let bytes = to_bin(&user)?;
let u: BorrowedUser = from_bin(&bytes)?;
```

## SIMD 加速

`to_bin` 使用平台 SIMD 指令批量拷贝 ≥ 32 字节的字符串载荷：

- **ARM**：NEON 128 位 load/store
- **x86-64**：SSE2 128 位 load/store

字符串以 16 字节为单位批量拷贝（`simd_bulk_extend`），避免标量字节循环。

## Rust API

```rust
use ason::{to_bin, to_bin_vec, from_bin, from_bin_vec};

// 单个值
let bytes: Vec<u8> = to_bin(&value)?;
let value: T       = from_bin(&bytes)?;

// 切片
let bytes: Vec<u8>  = to_bin_vec(&values)?;
let values: Vec<T>  = from_bin_vec(&bytes)?;
```

## 互操作性说明

ASON-BIN 是**私有线格式** —— 不适用于跨语言交互。当不同语言运行时需要通信时，请使用 ASON Text；当需要人工检查时同样使用 ASON Text。

在同一语言内部（如 Rust 服务向 Redis 写入并读取），ASON-BIN 始终是最佳选择。
