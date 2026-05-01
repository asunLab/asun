const std = @import("std");
const asun = @import("asun");

const Failure = struct {
    id: []const u8,
    msg: []const u8,
};

pub fn main() !void {
    var gpa_impl: std.heap.DebugAllocator(.{}) = .{};
    defer _ = gpa_impl.deinit();
    const gpa = gpa_impl.allocator();

    var arena_impl = std.heap.ArenaAllocator.init(gpa);
    defer arena_impl.deinit();
    const arena = arena_impl.allocator();

    var io_impl: std.Io.Threaded = .init(gpa, .{});
    defer io_impl.deinit();

    // ===== cases.json (untyped decode) =========================================
    const decode_json = try std.Io.Dir.cwd().readFileAlloc(
        io_impl.io(),
        "../../cases.json",
        arena,
        .limited(16 * 1024 * 1024),
    );
    const decode_parsed = try std.json.parseFromSlice(std.json.Value, arena, decode_json, .{});
    const decode_root = decode_parsed.value.object;
    const decode_cases = decode_root.get("cases").?.array.items;
    std.debug.print("loaded {d} cases from conformance/cases.json\n", .{decode_cases.len});

    var d_total: usize = 0;
    var d_ok_pass: usize = 0;
    var d_ok_fail: usize = 0;
    var d_err_pass: usize = 0;
    var d_err_fail: usize = 0;
    var d_skipped: usize = 0;
    var d_failures = std.array_list.Managed(Failure).init(arena);

    for (decode_cases) |case| {
        d_total += 1;
        const obj = case.object;
        const id = obj.get("id").?.string;
        if (obj.get("schemaDriven")) |sd| {
            if (sd == .bool and sd.bool) {
                d_skipped += 1;
                continue;
            }
        }
        const input = obj.get("input").?.string;
        const kind = obj.get("kind").?.string;

        var got_holder = asun.decodeZerocopy(std.json.Value, input, gpa) catch |err| {
            if (std.mem.eql(u8, kind, "error")) {
                d_err_pass += 1;
            } else {
                d_ok_fail += 1;
                if (d_failures.items.len < 25) {
                    try d_failures.append(.{
                        .id = id,
                        .msg = try std.fmt.allocPrint(arena, "expected ok, got error: {}\n    input: {s}", .{ err, input }),
                    });
                }
            }
            continue;
        };
        defer got_holder.deinit();
        const got = got_holder.value;

        if (std.mem.eql(u8, kind, "ok")) {
            const expected = obj.get("expected").?;
            if (jsonEqual(expected, got)) {
                d_ok_pass += 1;
            } else {
                d_ok_fail += 1;
                if (d_failures.items.len < 25) {
                    const exp_s = try std.json.Stringify.valueAlloc(arena, expected, .{});
                    const got_s = try std.json.Stringify.valueAlloc(arena, got, .{});
                    try d_failures.append(.{
                        .id = id,
                        .msg = try std.fmt.allocPrint(arena, "value mismatch\n    input:    {s}\n    expected: {s}\n    actual:   {s}", .{ input, exp_s, got_s }),
                    });
                }
            }
        } else {
            d_err_fail += 1;
            if (d_failures.items.len < 25) {
                const got_s = try std.json.Stringify.valueAlloc(arena, got, .{});
                try d_failures.append(.{
                    .id = id,
                    .msg = try std.fmt.allocPrint(arena, "expected error, got ok: {s}\n    input: {s}", .{ got_s, input }),
                });
            }
        }
    }

    std.debug.print("\n================ ASUN-ZIG conformance ================\n", .{});
    std.debug.print("total                : {d}\n", .{d_total});
    std.debug.print("untyped ok-cases pass: {d}\n", .{d_ok_pass});
    std.debug.print("untyped ok-cases fail: {d}\n", .{d_ok_fail});
    std.debug.print("untyped err-cases pass: {d}\n", .{d_err_pass});
    std.debug.print("untyped err-cases fail: {d}\n", .{d_err_fail});
    std.debug.print("skipped (needs typed): {d}\n", .{d_skipped});
    const executed = d_total - d_skipped;
    const dpct = if (executed > 0) @as(f64, @floatFromInt(d_ok_pass + d_err_pass)) / @as(f64, @floatFromInt(executed)) * 100.0 else 0.0;
    std.debug.print("untyped pass rate    : {d}/{d} ({d:.1}%)\n", .{ d_ok_pass + d_err_pass, executed, dpct });
    std.debug.print("=======================================================\n", .{});
    for (d_failures.items) |failure| {
        std.debug.print("\n[{s}]\n    {s}\n", .{ failure.id, failure.msg });
    }

    // ===== encode-cases.json (round-trip) ======================================
    const cases_json = try std.Io.Dir.cwd().readFileAlloc(
        io_impl.io(),
        "../../encode-cases.json",
        arena,
        .limited(16 * 1024 * 1024),
    );

    const parsed = try std.json.parseFromSlice(std.json.Value, arena, cases_json, .{});
    const manifest = parsed.value;
    const root = manifest.object;
    const cases = root.get("cases").?.array.items;
    const declared_count = root.get("count").?.integer;

    std.debug.print("loaded {d} encode cases from conformance/encode-cases.json\n", .{declared_count});

    var passed: usize = 0;
    var failed: usize = 0;
    var failures = std.array_list.Managed(Failure).init(arena);

    for (cases) |case| {
        const obj = case.object;
        const id = obj.get("id").?.string;
        const value = obj.get("value").?;

        const encoded = asun.encode(std.json.Value, value, arena) catch |err| {
            failed += 1;
            if (failures.items.len < 25) {
                try failures.append(.{
                    .id = id,
                    .msg = try std.fmt.allocPrint(arena, "encode error: {}", .{err}),
                });
            }
            continue;
        };

        var decoded_holder = asun.decodeZerocopy(std.json.Value, encoded, gpa) catch |err| {
            failed += 1;
            if (failures.items.len < 25) {
                try failures.append(.{
                    .id = id,
                    .msg = try std.fmt.allocPrint(arena, "decode error after encode: {}\n    encoded: {s}", .{ err, encoded }),
                });
            }
            continue;
        };
        const decoded = decoded_holder.value;

        if (!jsonEqual(value, decoded)) {
            failed += 1;
            if (failures.items.len < 25) {
                const expected = try std.json.Stringify.valueAlloc(arena, value, .{});
                const actual = try std.json.Stringify.valueAlloc(arena, decoded, .{});
                try failures.append(.{
                    .id = id,
                    .msg = try std.fmt.allocPrint(
                        arena,
                        "round-trip mismatch\n    encoded:  {s}\n    expected: {s}\n    actual:   {s}",
                        .{ encoded, expected, actual },
                    ),
                });
            }
            decoded_holder.deinit();
            continue;
        }

        decoded_holder.deinit();
        passed += 1;
    }

    std.debug.print("\n================ ASUN-ZIG encode conformance ================\n", .{});
    std.debug.print("total : {d}\n", .{cases.len});
    std.debug.print("pass  : {d}\n", .{passed});
    std.debug.print("fail  : {d}\n", .{failed});
    const pct = if (cases.len > 0) @as(f64, @floatFromInt(passed)) / @as(f64, @floatFromInt(cases.len)) * 100.0 else 0.0;
    std.debug.print("rate  : {d}/{d} ({d:.1}%)\n", .{ passed, cases.len, pct });
    std.debug.print("==============================================================\n", .{});

    for (failures.items) |failure| {
        std.debug.print("\n[{s}]\n    {s}\n", .{ failure.id, failure.msg });
    }

    if (failed > 0 or d_ok_fail > 0 or d_err_fail > 0) std.process.exit(1);
}

fn jsonEqual(a: std.json.Value, b: std.json.Value) bool {
    if (a == .integer and b == .float) return numericEqual(@floatFromInt(a.integer), b.float);
    if (a == .float and b == .integer) return numericEqual(a.float, @floatFromInt(b.integer));
    if (std.meta.activeTag(a) != std.meta.activeTag(b)) return false;
    return switch (a) {
        .null => true,
        .bool => |av| av == b.bool,
        .integer => |av| av == b.integer,
        .float => |av| numericEqual(av, b.float),
        .number_string => |av| std.mem.eql(u8, av, b.number_string),
        .string => |av| std.mem.eql(u8, av, b.string),
        .array => |av| blk: {
            const bv = b.array;
            if (av.items.len != bv.items.len) break :blk false;
            for (av.items, bv.items) |ai, bi| {
                if (!jsonEqual(ai, bi)) break :blk false;
            }
            break :blk true;
        },
        .object => |av| blk: {
            const bv = b.object;
            if (av.count() != bv.count()) break :blk false;
            var it = av.iterator();
            while (it.next()) |entry| {
                const other = bv.get(entry.key_ptr.*) orelse break :blk false;
                if (!jsonEqual(entry.value_ptr.*, other)) break :blk false;
            }
            break :blk true;
        },
    };
}

fn numericEqual(a: f64, b: f64) bool {
    if (a == b) return true;
    const diff = @abs(a - b);
    const scale = @max(@abs(a), @abs(b));
    return diff <= @max(1e-12, scale * 1e-12);
}
