# Zig Guide

The Zig ASON library leverages Zig's comptime features for zero-overhead serialization with full type safety.

## Installation

Add to `build.zig.zon`:

```zig
.dependencies = .{
    .ason = .{
        .url = "https://github.com/your-org/ason/archive/refs/heads/main.tar.gz",
    },
},
```

Then in `build.zig`:

```zig
const ason = b.dependency("ason", .{ .target = target, .optimize = optimize });
exe.root_module.addImport("ason", ason.module("ason"));
```

## Usage

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

    // Serialize
    const text = try ason.encodeSlice(allocator, &users);
    defer allocator.free(text);
    std.debug.print("{s}\n", .{text});

    // Deserialize
    const restored = try ason.decodeSlice(allocator, User, text);
    defer allocator.free(restored);

    // Binary
    const bytes = try ason.encodeBinSlice(allocator, &users);
    defer allocator.free(bytes);
    const restored2 = try ason.decodeBinSlice(allocator, User, bytes);
    defer allocator.free(restored2);
}
```

## Comptime Schema Generation

Zig's `comptime` allows the schema to be generated at compile time with zero runtime overhead:

```zig
// Schema is computed at comptime from the struct fields
const schema = ason.Schema(User);
// schema.field_names → .{"id", "name", "active"}
// schema.field_types → .{"int", "str", "bool"}
```

## Building Examples

```bash
cd zig
zig build run-basic
zig build run-bench
zig build run-complex
```

## Running Tests

```bash
cd zig && zig build test
```
