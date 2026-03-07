# Zig 指南

Zig ASON 库充分利用 Zig 的 comptime 特性，提供零开销、完全类型安全的序列化。

## 安装

在 `build.zig.zon` 中添加依赖：

```zig
.dependencies = .{
    .ason = .{
        .url = "https://github.com/your-org/ason/archive/refs/heads/main.tar.gz",
    },
},
```

在 `build.zig` 中引入：

```zig
const ason = b.dependency("ason", .{ .target = target, .optimize = optimize });
exe.root_module.addImport("ason", ason.module("ason"));
```

## 使用示例

```zig
const std = @import("std");
const ason = @import("ason");

const User = struct {
    id: i64,
    name: []const u8,
    active: bool,
};

pub fn main() !void {
    var gpa = std.heap.GeneralPurposeAllocator(.{}){};
    defer _ = gpa.deinit();
    const allocator = gpa.allocator();

    const users = [_]User{
        .{ .id = 1, .name = "Alice", .active = true },
        .{ .id = 2, .name = "Bob",   .active = false },
    };

    // 序列化
    const text = try ason.encodeSlice(allocator, &users);
    defer allocator.free(text);
    std.debug.print("{s}\n", .{text});

    // 反序列化
    const restored = try ason.decodeSlice(allocator, User, text);
    defer allocator.free(restored);

    // 二进制格式
    const bytes = try ason.encodeBinSlice(allocator, &users);
    defer allocator.free(bytes);
    const restored2 = try ason.decodeBinSlice(allocator, User, bytes);
    defer allocator.free(restored2);
}
```

## Comptime Schema 生成

Zig 的 `comptime` 允许在编译期生成 Schema，零运行时开销：

```zig
// Schema 在编译期从结构体字段自动生成
const schema = ason.Schema(User);
// schema.field_names → .{"id", "name", "active"}
// schema.field_types → .{"int", "str", "bool"}
```

## 构建示例

```bash
cd zig
zig build run-basic
zig build run-bench
zig build run-complex
```

## 运行测试

```bash
cd zig && zig build test
```
