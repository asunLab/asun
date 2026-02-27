//! WASM exports for browser / Node.js usage.
//! Compile with: zig build -Dtarget=wasm32-freestanding -Doptimize=ReleaseSmall

const std = @import("std");
const features = @import("features.zig");
const parser   = @import("parser.zig");

// ── Bump allocator — simple for WASM ─────────────────────────────────────────

var heap: [8 * 1024 * 1024]u8 = undefined; // 8 MiB
var fba = std.heap.FixedBufferAllocator.init(&heap);

fn alloc() std.mem.Allocator {
    return fba.allocator();
}

// Reset the bump allocator between calls.
export fn ason_reset() void {
    fba.reset();
}

// ── Shared output buffer ───────────────────────────────────────────────────────

var out_buf: [4 * 1024 * 1024]u8 = undefined;

fn writeOut(s: []const u8) usize {
    const n = @min(s.len, out_buf.len - 1);
    @memcpy(out_buf[0..n], s[0..n]);
    out_buf[n] = 0;
    return n;
}

export fn ason_out_ptr() [*]u8 {
    return &out_buf;
}

// ── Validate ──────────────────────────────────────────────────────────────────

/// Returns number of diagnostics (0 = valid).
export fn ason_validate(src_ptr: [*]const u8, src_len: usize) usize {
    const src = src_ptr[0..src_len];
    var result = parser.parse(src, alloc()) catch return 1;
    defer result.deinit();
    return result.diags.len;
}

// ── Format ────────────────────────────────────────────────────────────────────

/// Formats ASON, writes to out_buf, returns length.
export fn ason_format(src_ptr: [*]const u8, src_len: usize) usize {
    const src = src_ptr[0..src_len];
    const out = features.format(src, alloc()) catch return 0;
    return writeOut(out);
}

// ── Compress ──────────────────────────────────────────────────────────────────

export fn ason_compress(src_ptr: [*]const u8, src_len: usize) usize {
    const src = src_ptr[0..src_len];
    const out = features.compress(src, alloc()) catch return 0;
    return writeOut(out);
}

// ── ASON → JSON ────────────────────────────────────────────────────────────────

export fn ason_to_json(src_ptr: [*]const u8, src_len: usize) usize {
    const src = src_ptr[0..src_len];
    const out = features.asonToJson(src, alloc()) catch return 0;
    return writeOut(out);
}

// ── JSON → ASON ────────────────────────────────────────────────────────────────

export fn ason_from_json(src_ptr: [*]const u8, src_len: usize) usize {
    const src = src_ptr[0..src_len];
    const out = features.jsonToAson(src, alloc()) catch return 0;
    return writeOut(out);
}

// ── Complete (simple — returns newline-separated labels) ──────────────────────

export fn ason_complete(src_ptr: [*]const u8, src_len: usize, line: u32, col: u32) usize {
    const src = src_ptr[0..src_len];
    var result = parser.parse(src, alloc()) catch return 0;
    defer result.deinit();
    const items = features.complete(result.root, line, col, alloc()) catch return 0;
    var fbs = std.io.fixedBufferStream(&out_buf);
    const w = fbs.writer();
    for (items) |it| {
        w.print("{s}\n", .{it.label}) catch break;
    }
    return fbs.getWritten().len;
}
