# 性能基准

所有基准测试均在 **Apple M 系列**硬件上以 `--release` 构建（Rust）运行，其他语言的结果比例相近。

## 总结

| 指标 | ASON Text vs JSON | ASON-BIN vs JSON |
|------|------------------|-----------------|
| 序列化速度 | **2–4×** 更快 | **7–10×** 更快 |
| 反序列化速度 | **1.2–2.5×** 更快 | **2–2.5×** 更快 |
| 载荷体积 | **53–61%** 更小 | **38–49%** 更小 |

## 平面结构体（8 字段，× 1000 次）

```
struct User { id, name, email, age, active, score, role, created_at }
```

| 测试项 | JSON | ASON Text | ASON-BIN | Text vs JSON | BIN vs JSON |
|--------|------|-----------|----------|-------------|-------------|
| 序列化 | 2.19 ms | 1.07 ms | **0.28 ms** | 2.0× | **7.7×** |
| 反序列化 | 6.05 ms | 5.10 ms | **2.96 ms** | 1.2× | **2.0×** |
| 载荷 | 121 675 B | 56 716 B | **74 454 B** | 53% 更小 | 39% 更小 |

## 深层嵌套结构体（5 层，× 100 次）

```
Organization > Department > Team > Member > Contact
```

| 测试项 | JSON | ASON Text | ASON-BIN | Text vs JSON | BIN vs JSON |
|--------|------|-----------|----------|-------------|-------------|
| 序列化 | 4.19 ms | 2.67 ms | **0.50 ms** | 1.6× | **8.4×** |
| 反序列化 | 9.37 ms | 8.55 ms | **3.72 ms** | 1.1× | **2.5×** |
| 载荷 | 438 112 B | 174 611 B | **225 434 B** | 60% 更小 | 49% 更小 |

## 单次往返（User，× 100 000 次）

| 格式 | 每次时间 | vs JSON |
|------|---------|---------|
| ASON-BIN | ~182 ns | **2.1× 更快** |
| JSON | ~375 ns | — |
| ASON Text | ~552 ns | 0.68× |

> ASON Text 单结构体往返比 JSON 慢，因为 Schema/数据分离对单个对象有额外开销。ASON Text 的优势在**数组 ≥ 10 行**时才体现出来。

## ASON-BIN 为什么快

### 1. 无动态类型推断
二进制布局与结构体声明顺序一致，反序列化器直接调用类型化的解码函数，无需前瞻或类型分支。

### 2. 零哈希匹配
字段按位置索引，无字符串读取、无哈希计算、无映射查找。

### 3. 长度前缀字符串
字符串边界在读取前已知，无需扫描分隔符。

### 4. 零拷贝反序列化
带 `'de` 生命周期的字符串字段直接借用输入缓冲区，无 `String` 分配。

### 5. SIMD 字符串拷贝
≥ 32 字节的字符串以 16 字节 SIMD 块批量拷贝：
- **ARM**：NEON
- **x86-64**：SSE2

## 自行运行基准测试

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
