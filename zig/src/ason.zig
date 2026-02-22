//! ASON (Array-Schema Object Notation) — Zig implementation
//!
//! High-performance, zero-copy serialization/deserialization with SIMD acceleration.
//!
//! Public API:
//!   encode(T, value, allocator) -> []const u8
//!   decode(T, input, allocator) -> T
//!   encodeBinary(T, value, allocator) -> []u8
//!   decodeBinary(T, input) -> T           (zero-copy where possible)
//!
//! Wire format (binary): little-endian, no field names, positional.
//!   bool  → 1 byte (0/1)
//!   i8    → 1 byte
//!   i16   → 2 bytes LE
//!   i32   → 4 bytes LE
//!   i64   → 8 bytes LE
//!   u8..u64 → same widths LE
//!   f32   → 4 bytes LE (IEEE 754 bitcast)
//!   f64   → 8 bytes LE (IEEE 754 bitcast)
//!   []const u8 (string) → u32 LE length + UTF-8 bytes (zero-copy on decode)
//!   ?T    → u8 tag (0=null,1=some) + [T if some]
//!   []T   → u32 LE count + [T × count]
//!   struct → fields in declaration order

const std = @import("std");
const mem = std.mem;
const math = std.math;
const Allocator = mem.Allocator;

pub const AsonError = error{
    Eof,
    InvalidFormat,
    InvalidNumber,
    InvalidBool,
    InvalidEscape,
    UnclosedString,
    UnclosedComment,
    UnclosedParen,
    UnclosedBracket,
    ExpectedColon,
    ExpectedOpenBrace,
    ExpectedOpenParen,
    ExpectedCloseParen,
    TrailingCharacters,
    OutOfMemory,
    BufferOverflow,
    Utf8Error,
    UnsupportedType,
};

// ============================================================================
// Writer: wraps unmanaged ArrayList(u8) + stored allocator
// ============================================================================

const Writer = struct {
    buf: std.ArrayList(u8) = .{},
    gpa: Allocator,

    fn init(gpa: Allocator) Writer {
        return .{ .gpa = gpa };
    }

    fn deinit(self: *Writer) void {
        self.buf.deinit(self.gpa);
    }

    fn append(self: *Writer, byte: u8) Allocator.Error!void {
        return self.buf.append(self.gpa, byte);
    }

    fn appendSlice(self: *Writer, bytes: []const u8) Allocator.Error!void {
        return self.buf.appendSlice(self.gpa, bytes);
    }

    fn ensureUnusedCapacity(self: *Writer, n: usize) Allocator.Error!void {
        return self.buf.ensureUnusedCapacity(self.gpa, n);
    }

    fn toOwnedSlice(self: *Writer) Allocator.Error![]u8 {
        return self.buf.toOwnedSlice(self.gpa);
    }
};

// ============================================================================
// SIMD helpers
// ============================================================================

const LANES = 16;

/// SIMD-accelerated: check if bytes contain any ASON special char
/// Special chars: control (<= 0x1f), comma, parens, brackets, quote, backslash
pub inline fn simdHasSpecialChars(bytes: []const u8) bool {
    var i: usize = 0;

    while (i + LANES <= bytes.len) : (i += LANES) {
        const chunk = simdLoad(bytes.ptr + i);
        const mask = simdSpecialMask(chunk);
        if (mask != 0) return true;
    }

    while (i < bytes.len) : (i += 1) {
        if (needsQuoteLut(bytes[i])) return true;
    }
    return false;
}

/// SIMD-accelerated: find first byte needing escape in a string (" \ or control)
pub inline fn simdFindEscape(bytes: []const u8, start: usize) usize {
    var i = start;

    while (i + LANES <= bytes.len) : (i += LANES) {
        const chunk = simdLoad(bytes.ptr + i);
        const mask = simdEscapeMask(chunk);
        if (mask != 0) return i + @ctz(mask);
    }

    while (i < bytes.len) : (i += 1) {
        const b = bytes[i];
        if (b <= 0x1f or b == '"' or b == '\\') return i;
    }
    return bytes.len;
}

/// SIMD-accelerated: find first quote or backslash
pub inline fn simdFindQuoteOrBackslash(bytes: []const u8, start: usize) usize {
    var i = start;

    while (i + LANES <= bytes.len) : (i += LANES) {
        const chunk = simdLoad(bytes.ptr + i);
        const mask = simdQuoteBackslashMask(chunk);
        if (mask != 0) return i + @ctz(mask);
    }

    while (i < bytes.len) : (i += 1) {
        const b = bytes[i];
        if (b == '"' or b == '\\') return i;
    }
    return bytes.len;
}

/// SIMD-accelerated: find first delimiter (, ) ] \)
pub inline fn simdFindPlainDelimiter(bytes: []const u8, start: usize) usize {
    var i = start;

    while (i + LANES <= bytes.len) : (i += LANES) {
        const chunk = simdLoad(bytes.ptr + i);
        const mask = simdDelimiterMask(chunk);
        if (mask != 0) return i + @ctz(mask);
    }

    while (i < bytes.len) : (i += 1) {
        switch (bytes[i]) {
            ',', ')', ']', '\\' => return i,
            else => {},
        }
    }
    return bytes.len;
}

/// SIMD-accelerated: skip whitespace
pub inline fn simdSkipWhitespace(bytes: []const u8, start: usize) usize {
    var i = start;

    while (i + LANES <= bytes.len) : (i += LANES) {
        const chunk = simdLoad(bytes.ptr + i);
        const ws_mask = simdWhitespaceMask(chunk);
        if (ws_mask == 0xFFFF) continue;
        if (ws_mask == 0) return i;
        const non_ws = ~ws_mask & 0xFFFF;
        return i + @ctz(non_ws);
    }

    while (i < bytes.len) : (i += 1) {
        switch (bytes[i]) {
            ' ', '\t', '\n', '\r' => {},
            else => return i,
        }
    }
    return i;
}

/// SIMD bulk memory copy for binary serialization
pub inline fn simdBulkExtend(w: *Writer, src: []const u8) !void {
    if (src.len == 0) return;
    if (src.len < 32) {
        try w.appendSlice(src);
        return;
    }
    try w.ensureUnusedCapacity(src.len);
    const dst_start = w.buf.items.len;
    var i: usize = 0;
    while (i + LANES <= src.len) : (i += LANES) {
        const chunk = simdLoad(src.ptr + i);
        simdStore(w.buf.items.ptr + dst_start + i, chunk);
    }
    if (i < src.len) {
        @memcpy(w.buf.items.ptr[dst_start + i .. dst_start + src.len], src[i..]);
    }
    w.buf.items.len = dst_start + src.len;
}

// ---- Platform SIMD primitives ----

const SimdVec = @Vector(LANES, u8);

inline fn simdLoad(ptr: [*]const u8) SimdVec {
    return ptr[0..LANES].*;
}

inline fn simdStore(ptr: [*]u8, v: SimdVec) void {
    ptr[0..LANES].* = v;
}

inline fn simdSplat(b: u8) SimdVec {
    return @splat(b);
}

inline fn simdCmpeq(a: SimdVec, b: SimdVec) u16 {
    const cmp: @Vector(LANES, bool) = a == b;
    return @bitCast(cmp);
}

inline fn simdCmple(a: SimdVec, b: SimdVec) u16 {
    const cmp: @Vector(LANES, bool) = a <= b;
    return @bitCast(cmp);
}

inline fn simdSpecialMask(chunk: SimdVec) u16 {
    const ctrl = simdCmple(chunk, simdSplat(0x1f));
    const comma = simdCmpeq(chunk, simdSplat(','));
    const lparen = simdCmpeq(chunk, simdSplat('('));
    const rparen = simdCmpeq(chunk, simdSplat(')'));
    const lbracket = simdCmpeq(chunk, simdSplat('['));
    const rbracket = simdCmpeq(chunk, simdSplat(']'));
    const quote = simdCmpeq(chunk, simdSplat('"'));
    const backslash = simdCmpeq(chunk, simdSplat('\\'));
    return ctrl | comma | lparen | rparen | lbracket | rbracket | quote | backslash;
}

inline fn simdEscapeMask(chunk: SimdVec) u16 {
    const ctrl = simdCmple(chunk, simdSplat(0x1f));
    const quote = simdCmpeq(chunk, simdSplat('"'));
    const backslash = simdCmpeq(chunk, simdSplat('\\'));
    return ctrl | quote | backslash;
}

inline fn simdQuoteBackslashMask(chunk: SimdVec) u16 {
    const quote = simdCmpeq(chunk, simdSplat('"'));
    const backslash = simdCmpeq(chunk, simdSplat('\\'));
    return quote | backslash;
}

inline fn simdDelimiterMask(chunk: SimdVec) u16 {
    const comma = simdCmpeq(chunk, simdSplat(','));
    const rparen = simdCmpeq(chunk, simdSplat(')'));
    const rbracket = simdCmpeq(chunk, simdSplat(']'));
    const backslash = simdCmpeq(chunk, simdSplat('\\'));
    return comma | rparen | rbracket | backslash;
}

inline fn simdWhitespaceMask(chunk: SimdVec) u16 {
    const sp = simdCmpeq(chunk, simdSplat(' '));
    const tab = simdCmpeq(chunk, simdSplat('\t'));
    const nl = simdCmpeq(chunk, simdSplat('\n'));
    const cr = simdCmpeq(chunk, simdSplat('\r'));
    return sp | tab | nl | cr;
}

fn needsQuoteLut(b: u8) bool {
    return b <= 0x1f or b == ',' or b == '(' or b == ')' or b == '[' or b == ']' or b == '"' or b == '\\';
}

// ============================================================================
// Two-digit lookup for fast integer formatting
// ============================================================================

const DEC_DIGITS: *const [200]u8 = "00010203040506070809101112131415161718192021222324252627282930313233343536373839404142434445464748495051525354555657585960616263646566676869707172737475767778798081828384858687888990919293949596979899";

// ============================================================================
// Text Serializer (encode)
// ============================================================================

pub fn encode(comptime T: type, value: T, allocator: Allocator) ![]const u8 {
    var w = Writer.init(allocator);
    errdefer w.deinit();
    try serializeValue(T, value, &w, true);
    return w.toOwnedSlice();
}

pub fn encodeVec(comptime T: type, values: []const T, allocator: Allocator) ![]const u8 {
    var w = Writer.init(allocator);
    errdefer w.deinit();

    try writeSchema(T, &w, false);
    try w.append(':');

    for (values, 0..) |v, i| {
        if (i > 0) try w.append(',');
        try writeTupleData(T, v, &w);
    }
    return w.toOwnedSlice();
}

pub fn encodeVecTyped(comptime T: type, values: []const T, allocator: Allocator) ![]const u8 {
    var w = Writer.init(allocator);
    errdefer w.deinit();

    try writeSchema(T, &w, true);
    try w.append(':');

    for (values, 0..) |v, i| {
        if (i > 0) try w.append(',');
        try writeTupleData(T, v, &w);
    }
    return w.toOwnedSlice();
}

pub fn encodeTyped(comptime T: type, value: T, allocator: Allocator) ![]const u8 {
    var w = Writer.init(allocator);
    errdefer w.deinit();
    try serializeValue(T, value, &w, true);
    return w.toOwnedSlice();
}

fn writeSchema(comptime T: type, w: *Writer, typed: bool) !void {
    const info = @typeInfo(T);
    switch (info) {
        .@"struct" => |s| {
            try w.append('{');
            inline for (s.fields, 0..) |field, i| {
                if (i > 0) try w.append(',');
                try w.appendSlice(field.name);
                if (typed) {
                    try w.append(':');
                    try writeTypeHint(field.type, w);
                }
            }
            try w.append('}');
        },
        else => return error.UnsupportedType,
    }
}

fn writeTypeHint(comptime T: type, w: *Writer) !void {
    const info = @typeInfo(T);
    switch (info) {
        .bool => try w.appendSlice("bool"),
        .int => {
            try w.appendSlice("int");
        },
        .float => try w.appendSlice("float"),
        .pointer => |ptr| {
            if (ptr.size == .slice and ptr.child == u8) {
                try w.appendSlice("str");
            } else if (ptr.size == .slice) {
                try w.append('[');
                try writeTypeHint(ptr.child, w);
                try w.append(']');
            } else {
                try w.appendSlice("str");
            }
        },
        .optional => |opt| try writeTypeHint(opt.child, w),
        .@"struct" => |s| {
            try w.append('{');
            inline for (s.fields, 0..) |field, i| {
                if (i > 0) try w.append(',');
                try w.appendSlice(field.name);
                try w.append(':');
                try writeTypeHint(field.type, w);
            }
            try w.append('}');
        },
        else => {
            if (comptime isSlice(T)) {
                try w.append('[');
                try writeTypeHint(SliceChild(T), w);
                try w.append(']');
            } else {
                try w.appendSlice("str");
            }
        },
    }
}

fn writeTupleData(comptime T: type, value: T, w: *Writer) !void {
    const info = @typeInfo(T);
    switch (info) {
        .@"struct" => |s| {
            try w.append('(');
            inline for (s.fields, 0..) |field, i| {
                if (i > 0) try w.append(',');
                try serializeField(field.type, @field(value, field.name), w);
            }
            try w.append(')');
        },
        else => return error.UnsupportedType,
    }
}

fn serializeValue(comptime T: type, value: T, w: *Writer, typed: bool) !void {
    const info = @typeInfo(T);
    switch (info) {
        .@"struct" => |s| {
            try w.append('{');
            inline for (s.fields, 0..) |field, i| {
                if (i > 0) try w.append(',');
                try w.appendSlice(field.name);
                if (typed) {
                    try w.append(':');
                    try writeTypeHint(field.type, w);
                }
            }
            try w.appendSlice("}:");

            try w.append('(');
            inline for (s.fields, 0..) |field, i| {
                if (i > 0) try w.append(',');
                try serializeField(field.type, @field(value, field.name), w);
            }
            try w.append(')');
        },
        else => return error.UnsupportedType,
    }
}

fn serializeField(comptime T: type, value: T, w: *Writer) !void {
    const info = @typeInfo(T);
    switch (info) {
        .bool => {
            if (value) {
                try w.appendSlice("true");
            } else {
                try w.appendSlice("false");
            }
        },
        .int => |i_info| {
            if (i_info.signedness == .signed) {
                try writeI64(w, @as(i64, @intCast(value)));
            } else {
                try writeU64(w, @as(u64, @intCast(value)));
            }
        },
        .float => {
            try writeF64(w, @as(f64, @floatCast(value)));
        },
        .optional => |opt| {
            if (value) |v| {
                try serializeField(opt.child, v, w);
            }
            // else: empty (null)
        },
        .pointer => |ptr| {
            if (ptr.size == .slice and ptr.child == u8) {
                try writeString(w, value);
            } else if (ptr.size == .slice) {
                try w.append('[');
                for (value, 0..) |item, i| {
                    if (i > 0) try w.append(',');
                    try serializeField(ptr.child, item, w);
                }
                try w.append(']');
            } else {
                return error.UnsupportedType;
            }
        },
        .@"struct" => |s| {
            try w.append('(');
            inline for (s.fields, 0..) |field, i| {
                if (i > 0) try w.append(',');
                try serializeField(field.type, @field(value, field.name), w);
            }
            try w.append(')');
        },
        else => return error.UnsupportedType,
    }
}

fn needsQuoting(s: []const u8) bool {
    if (s.len == 0) return true;
    if (s[0] == ' ' or s[s.len - 1] == ' ') return true;
    if (mem.eql(u8, s, "true") or mem.eql(u8, s, "false")) return true;
    if (simdHasSpecialChars(s)) return true;

    var start: usize = 0;
    if (s[0] == '-') start = 1;
    if (start < s.len) {
        var could_be_number = true;
        for (s[start..]) |c| {
            if (!std.ascii.isDigit(c) and c != '.') {
                could_be_number = false;
                break;
            }
        }
        if (could_be_number) return true;
    }
    return false;
}

fn writeString(w: *Writer, s: []const u8) !void {
    if (needsQuoting(s)) {
        try writeEscaped(w, s);
    } else {
        try w.appendSlice(s);
    }
}

fn writeEscaped(w: *Writer, s: []const u8) !void {
    try w.append('"');
    var start: usize = 0;
    while (start < s.len) {
        const next = simdFindEscape(s, start);
        if (next > start) {
            try w.appendSlice(s[start..next]);
        }
        if (next >= s.len) break;
        const b = s[next];
        switch (b) {
            '"' => try w.appendSlice("\\\""),
            '\\' => try w.appendSlice("\\\\"),
            '\n' => try w.appendSlice("\\n"),
            '\t' => try w.appendSlice("\\t"),
            '\r' => try w.appendSlice("\\r"),
            else => {
                try w.appendSlice("\\u00");
                const HEX = "0123456789abcdef";
                try w.append(HEX[b >> 4]);
                try w.append(HEX[b & 0xf]);
            },
        }
        start = next + 1;
    }
    try w.append('"');
}

fn writeI64(w: *Writer, v: i64) !void {
    if (v < 0) {
        try w.append('-');
        if (v == math.minInt(i64)) {
            try w.appendSlice("9223372036854775808");
            return;
        }
        try writeU64(w, @intCast(-v));
    } else {
        try writeU64(w, @intCast(v));
    }
}

fn writeU64(w: *Writer, v: u64) !void {
    if (v < 10) {
        try w.append(@as(u8, @intCast(v)) + '0');
        return;
    }
    if (v < 100) {
        const idx: usize = @intCast(v * 2);
        try w.append(DEC_DIGITS[idx]);
        try w.append(DEC_DIGITS[idx + 1]);
        return;
    }
    var tmp: [20]u8 = undefined;
    var i: usize = 20;
    var val = v;
    while (val >= 100) {
        const rem: usize = @intCast(val % 100);
        val /= 100;
        i -= 2;
        tmp[i] = DEC_DIGITS[rem * 2];
        tmp[i + 1] = DEC_DIGITS[rem * 2 + 1];
    }
    if (val >= 10) {
        const idx: usize = @intCast(val * 2);
        i -= 2;
        tmp[i] = DEC_DIGITS[idx];
        tmp[i + 1] = DEC_DIGITS[idx + 1];
    } else {
        i -= 1;
        tmp[i] = @as(u8, @intCast(val)) + '0';
    }
    try w.appendSlice(tmp[i..20]);
}

fn writeF64(w: *Writer, v: f64) !void {
    var tmp: [64]u8 = undefined;
    const s = std.fmt.bufPrint(&tmp, "{d}", .{v}) catch return error.InvalidNumber;
    try w.appendSlice(s);
    if (mem.indexOfScalar(u8, s, '.') == null and mem.indexOfScalar(u8, s, 'e') == null) {
        try w.appendSlice(".0");
    }
}

// ============================================================================
// Helper: is a Zig type a slice?
// ============================================================================

fn isSlice(comptime T: type) bool {
    const info = @typeInfo(T);
    return info == .pointer and info.pointer.size == .slice;
}

fn SliceChild(comptime T: type) type {
    return @typeInfo(T).pointer.child;
}

// ============================================================================
// Text Deserializer (decode)
// ============================================================================

pub fn decode(comptime T: type, input: []const u8, allocator: Allocator) !T {
    var parser = Parser{ .input = input, .pos = 0, .allocator = allocator };
    parser.skipWhitespaceAndComments();

    if (parser.pos < parser.input.len and parser.input[parser.pos] == '{') {
        try parser.skipSchema();
        parser.skipWhitespaceAndComments();
        if (parser.pos < parser.input.len and parser.input[parser.pos] == ':') {
            parser.pos += 1;
        }
        parser.skipWhitespaceAndComments();
    }

    return parser.parseStruct(T);
}

pub fn decodeVec(comptime T: type, input: []const u8, allocator: Allocator) ![]T {
    var parser = Parser{ .input = input, .pos = 0, .allocator = allocator };
    parser.skipWhitespaceAndComments();

    if (parser.pos < parser.input.len and parser.input[parser.pos] == '{') {
        try parser.skipSchema();
        parser.skipWhitespaceAndComments();
        if (parser.pos < parser.input.len and parser.input[parser.pos] == ':') {
            parser.pos += 1;
        }
        parser.skipWhitespaceAndComments();
    }

    var results: std.ArrayList(T) = .{};
    errdefer results.deinit(allocator);

    while (parser.pos < parser.input.len) {
        parser.skipWhitespaceAndComments();
        if (parser.pos >= parser.input.len) break;
        if (parser.input[parser.pos] != '(') break;

        const val = try parser.parseStruct(T);
        try results.append(allocator, val);

        parser.skipWhitespaceAndComments();
        if (parser.pos < parser.input.len and parser.input[parser.pos] == ',') {
            parser.pos += 1;
            parser.skipWhitespaceAndComments();
        }
    }

    return results.toOwnedSlice(allocator);
}

const Parser = struct {
    input: []const u8,
    pos: usize,
    allocator: Allocator,

    fn peekByte(self: *Parser) !u8 {
        if (self.pos < self.input.len) return self.input[self.pos];
        return error.Eof;
    }

    fn nextByte(self: *Parser) !u8 {
        if (self.pos < self.input.len) {
            const b = self.input[self.pos];
            self.pos += 1;
            return b;
        }
        return error.Eof;
    }

    fn skipWhitespaceAndComments(self: *Parser) void {
        while (true) {
            self.pos = simdSkipWhitespace(self.input, self.pos);
            if (self.pos + 1 < self.input.len and self.input[self.pos] == '/' and self.input[self.pos + 1] == '*') {
                self.pos += 2;
                while (self.pos + 1 < self.input.len) {
                    if (self.input[self.pos] == '*' and self.input[self.pos + 1] == '/') {
                        self.pos += 2;
                        break;
                    }
                    self.pos += 1;
                }
                continue;
            }
            break;
        }
    }

    fn skipSchema(self: *Parser) !void {
        if (self.input[self.pos] != '{') return error.ExpectedOpenBrace;
        var depth: usize = 0;
        while (self.pos < self.input.len) {
            const b = self.input[self.pos];
            self.pos += 1;
            if (b == '{') {
                depth += 1;
            } else if (b == '}') {
                depth -= 1;
                if (depth == 0) return;
            }
        }
        return error.Eof;
    }

    fn parseStruct(self: *Parser, comptime T: type) !T {
        const info = @typeInfo(T);
        switch (info) {
            .@"struct" => |s| {
                self.skipWhitespaceAndComments();
                if ((try self.peekByte()) != '(') return error.ExpectedOpenParen;
                self.pos += 1;
                var result: T = undefined;
                inline for (s.fields, 0..) |field, i| {
                    if (i > 0) {
                        self.skipWhitespaceAndComments();
                        if ((try self.peekByte()) == ',') {
                            self.pos += 1;
                        }
                    }
                    self.skipWhitespaceAndComments();
                    @field(result, field.name) = try self.parseField(field.type);
                }
                self.skipWhitespaceAndComments();
                if ((try self.peekByte()) != ')') return error.ExpectedCloseParen;
                self.pos += 1;
                return result;
            },
            else => return error.UnsupportedType,
        }
    }

    fn parseField(self: *Parser, comptime T: type) !T {
        const info = @typeInfo(T);
        switch (info) {
            .bool => return self.parseBool(),
            .int => return self.parseInt(T),
            .float => return self.parseFloat(T),
            .pointer => |ptr| {
                if (ptr.size == .slice and ptr.child == u8) {
                    return self.parseString();
                } else if (ptr.size == .slice) {
                    return self.parseArray(ptr.child);
                } else {
                    return error.UnsupportedType;
                }
            },
            .optional => |opt| {
                self.skipWhitespaceAndComments();
                const b = try self.peekByte();
                if (b == ',' or b == ')') return null;
                return try self.parseField(opt.child);
            },
            .@"struct" => return self.parseStruct(T),
            else => return error.UnsupportedType,
        }
    }

    fn parseBool(self: *Parser) !bool {
        self.skipWhitespaceAndComments();
        if (self.pos + 4 <= self.input.len and mem.eql(u8, self.input[self.pos .. self.pos + 4], "true")) {
            self.pos += 4;
            return true;
        }
        if (self.pos + 5 <= self.input.len and mem.eql(u8, self.input[self.pos .. self.pos + 5], "false")) {
            self.pos += 5;
            return false;
        }
        return error.InvalidBool;
    }

    fn parseInt(self: *Parser, comptime T: type) !T {
        self.skipWhitespaceAndComments();
        var negative = false;
        if (self.pos < self.input.len and self.input[self.pos] == '-') {
            negative = true;
            self.pos += 1;
        }
        var val: i128 = 0;
        var found = false;
        while (self.pos < self.input.len and std.ascii.isDigit(self.input[self.pos])) {
            val = val * 10 + @as(i128, self.input[self.pos] - '0');
            self.pos += 1;
            found = true;
        }
        if (self.pos < self.input.len and self.input[self.pos] == '.') {
            self.pos += 1;
            while (self.pos < self.input.len and std.ascii.isDigit(self.input[self.pos])) {
                self.pos += 1;
            }
        }
        if (!found) return error.InvalidNumber;
        if (negative) val = -val;
        return @intCast(val);
    }

    fn parseFloat(self: *Parser, comptime T: type) !T {
        self.skipWhitespaceAndComments();
        const start = self.pos;
        if (self.pos < self.input.len and (self.input[self.pos] == '-' or self.input[self.pos] == '+')) {
            self.pos += 1;
        }
        while (self.pos < self.input.len and std.ascii.isDigit(self.input[self.pos])) {
            self.pos += 1;
        }
        if (self.pos < self.input.len and self.input[self.pos] == '.') {
            self.pos += 1;
            while (self.pos < self.input.len and std.ascii.isDigit(self.input[self.pos])) {
                self.pos += 1;
            }
        }
        if (self.pos < self.input.len and (self.input[self.pos] == 'e' or self.input[self.pos] == 'E')) {
            self.pos += 1;
            if (self.pos < self.input.len and (self.input[self.pos] == '-' or self.input[self.pos] == '+')) {
                self.pos += 1;
            }
            while (self.pos < self.input.len and std.ascii.isDigit(self.input[self.pos])) {
                self.pos += 1;
            }
        }
        if (self.pos == start) return error.InvalidNumber;
        const slice = self.input[start..self.pos];
        return std.fmt.parseFloat(T, slice) catch return error.InvalidNumber;
    }

    fn parseString(self: *Parser) ![]const u8 {
        self.skipWhitespaceAndComments();
        if (self.pos >= self.input.len) return error.Eof;

        if (self.input[self.pos] == '"') {
            return self.parseQuotedString();
        } else {
            return self.parsePlainString();
        }
    }

    fn parseQuotedString(self: *Parser) ![]const u8 {
        self.pos += 1; // skip opening "
        var result: std.ArrayList(u8) = .{};
        errdefer result.deinit(self.allocator);

        while (self.pos < self.input.len) {
            const next = simdFindQuoteOrBackslash(self.input, self.pos);
            if (next > self.pos) {
                try result.appendSlice(self.allocator, self.input[self.pos..next]);
            }
            if (next >= self.input.len) return error.UnclosedString;
            self.pos = next;
            const b = self.input[self.pos];
            if (b == '"') {
                self.pos += 1;
                return result.toOwnedSlice(self.allocator);
            }
            // Backslash escape
            self.pos += 1;
            if (self.pos >= self.input.len) return error.InvalidEscape;
            const esc = self.input[self.pos];
            self.pos += 1;
            switch (esc) {
                '"' => try result.append(self.allocator, '"'),
                '\\' => try result.append(self.allocator, '\\'),
                'n' => try result.append(self.allocator, '\n'),
                't' => try result.append(self.allocator, '\t'),
                'r' => try result.append(self.allocator, '\r'),
                'u' => {
                    if (self.pos + 4 > self.input.len) return error.InvalidEscape;
                    const hex = self.input[self.pos .. self.pos + 4];
                    const cp = std.fmt.parseInt(u21, hex, 16) catch return error.InvalidEscape;
                    var utf8_buf: [4]u8 = undefined;
                    const len = std.unicode.utf8Encode(cp, &utf8_buf) catch return error.InvalidEscape;
                    try result.appendSlice(self.allocator, utf8_buf[0..len]);
                    self.pos += 4;
                },
                else => return error.InvalidEscape,
            }
        }
        return error.UnclosedString;
    }

    fn parsePlainString(self: *Parser) ![]const u8 {
        const start = self.pos;
        const end_pos = simdFindPlainDelimiter(self.input, self.pos);
        self.pos = end_pos;
        var end = end_pos;
        while (end > start and (self.input[end - 1] == ' ' or self.input[end - 1] == '\t')) {
            end -= 1;
        }
        if (end == start) return "";
        const slice = self.input[start..end];
        return self.allocator.dupe(u8, slice);
    }

    fn parseArray(self: *Parser, comptime Child: type) ![]Child {
        self.skipWhitespaceAndComments();
        if ((try self.peekByte()) != '[') return error.InvalidFormat;
        self.pos += 1;

        var result: std.ArrayList(Child) = .{};
        errdefer result.deinit(self.allocator);

        self.skipWhitespaceAndComments();
        if (self.pos < self.input.len and self.input[self.pos] == ']') {
            self.pos += 1;
            return result.toOwnedSlice(self.allocator);
        }

        while (true) {
            self.skipWhitespaceAndComments();
            const item = try self.parseField(Child);
            try result.append(self.allocator, item);
            self.skipWhitespaceAndComments();
            if (self.pos >= self.input.len) return error.Eof;
            if (self.input[self.pos] == ']') {
                self.pos += 1;
                break;
            }
            if (self.input[self.pos] == ',') {
                self.pos += 1;
            }
        }

        return result.toOwnedSlice(self.allocator);
    }
};

// ============================================================================
// Binary Serializer (encodeBinary)
// ============================================================================

pub fn encodeBinary(comptime T: type, value: T, allocator: Allocator) ![]u8 {
    var w = Writer.init(allocator);
    errdefer w.deinit();
    try binSerialize(T, value, &w);
    return w.toOwnedSlice();
}

pub fn encodeBinaryVec(comptime T: type, values: []const T, allocator: Allocator) ![]u8 {
    var w = Writer.init(allocator);
    errdefer w.deinit();
    try binWriteU32(&w, @intCast(values.len));
    for (values) |v| {
        try binSerialize(T, v, &w);
    }
    return w.toOwnedSlice();
}

fn binSerialize(comptime T: type, value: T, w: *Writer) !void {
    const info = @typeInfo(T);
    switch (info) {
        .bool => try w.append(@intFromBool(value)),
        .int => |i_info| {
            switch (i_info.bits) {
                8 => try w.append(@bitCast(value)),
                16 => {
                    const T16 = if (i_info.signedness == .signed) i16 else u16;
                    const v: T16 = @intCast(value);
                    try w.appendSlice(&mem.toBytes(mem.nativeToLittle(T16, v)));
                },
                32 => {
                    const T32 = if (i_info.signedness == .signed) i32 else u32;
                    const v: T32 = @intCast(value);
                    try w.appendSlice(&mem.toBytes(mem.nativeToLittle(T32, v)));
                },
                64 => {
                    const T64 = if (i_info.signedness == .signed) i64 else u64;
                    const v: T64 = @intCast(value);
                    try w.appendSlice(&mem.toBytes(mem.nativeToLittle(T64, v)));
                },
                else => return error.UnsupportedType,
            }
        },
        .float => |f_info| {
            switch (f_info.bits) {
                32 => {
                    const bits: u32 = @bitCast(@as(f32, @floatCast(value)));
                    try w.appendSlice(&mem.toBytes(mem.nativeToLittle(u32, bits)));
                },
                64 => {
                    const bits: u64 = @bitCast(@as(f64, value));
                    try w.appendSlice(&mem.toBytes(mem.nativeToLittle(u64, bits)));
                },
                else => return error.UnsupportedType,
            }
        },
        .pointer => |ptr| {
            if (ptr.size == .slice and ptr.child == u8) {
                try binWriteU32(w, @intCast(value.len));
                try simdBulkExtend(w, value);
            } else if (ptr.size == .slice) {
                try binWriteU32(w, @intCast(value.len));
                if (comptime isNumericType(ptr.child)) {
                    const byte_len = value.len * @sizeOf(ptr.child);
                    const bytes: [*]const u8 = @ptrCast(value.ptr);
                    try simdBulkExtend(w, bytes[0..byte_len]);
                } else {
                    for (value) |item| {
                        try binSerialize(ptr.child, item, w);
                    }
                }
            } else {
                return error.UnsupportedType;
            }
        },
        .optional => |opt| {
            if (value) |v| {
                try w.append(1);
                try binSerialize(opt.child, v, w);
            } else {
                try w.append(0);
            }
        },
        .@"struct" => |s| {
            inline for (s.fields) |field| {
                try binSerialize(field.type, @field(value, field.name), w);
            }
        },
        else => return error.UnsupportedType,
    }
}

fn isNumericType(comptime T: type) bool {
    const info = @typeInfo(T);
    return info == .int or info == .float;
}

fn binWriteU32(w: *Writer, v: u32) !void {
    try w.appendSlice(&mem.toBytes(mem.nativeToLittle(u32, v)));
}

// ============================================================================
// Binary Deserializer (decodeBinary) — zero-copy
// ============================================================================

pub fn decodeBinary(comptime T: type, data: []const u8, allocator: Allocator) !T {
    var reader = BinReader{ .data = data, .pos = 0, .allocator = allocator };
    return reader.readValue(T);
}

pub fn decodeBinaryVec(comptime T: type, data: []const u8, allocator: Allocator) ![]T {
    var reader = BinReader{ .data = data, .pos = 0, .allocator = allocator };
    const count = try reader.readU32();
    var result = try std.ArrayList(T).initCapacity(allocator, count);
    errdefer result.deinit(allocator);
    for (0..count) |_| {
        try result.append(allocator, try reader.readValue(T));
    }
    return result.toOwnedSlice(allocator);
}

const BinReader = struct {
    data: []const u8,
    pos: usize,
    allocator: Allocator,

    inline fn ensure(self: *BinReader, n: usize) !void {
        if (self.pos + n > self.data.len) return error.Eof;
    }

    inline fn readU8(self: *BinReader) !u8 {
        try self.ensure(1);
        const v = self.data[self.pos];
        self.pos += 1;
        return v;
    }

    inline fn readU16(self: *BinReader) !u16 {
        try self.ensure(2);
        const v = mem.readInt(u16, self.data[self.pos..][0..2], .little);
        self.pos += 2;
        return v;
    }

    inline fn readU32(self: *BinReader) !u32 {
        try self.ensure(4);
        const v = mem.readInt(u32, self.data[self.pos..][0..4], .little);
        self.pos += 4;
        return v;
    }

    inline fn readU64(self: *BinReader) !u64 {
        try self.ensure(8);
        const v = mem.readInt(u64, self.data[self.pos..][0..8], .little);
        self.pos += 8;
        return v;
    }

    inline fn readI8(self: *BinReader) !i8 {
        return @bitCast(try self.readU8());
    }

    inline fn readI16(self: *BinReader) !i16 {
        try self.ensure(2);
        const v = mem.readInt(i16, self.data[self.pos..][0..2], .little);
        self.pos += 2;
        return v;
    }

    inline fn readI32(self: *BinReader) !i32 {
        try self.ensure(4);
        const v = mem.readInt(i32, self.data[self.pos..][0..4], .little);
        self.pos += 4;
        return v;
    }

    inline fn readI64(self: *BinReader) !i64 {
        try self.ensure(8);
        const v = mem.readInt(i64, self.data[self.pos..][0..8], .little);
        self.pos += 8;
        return v;
    }

    inline fn readF32(self: *BinReader) !f32 {
        const bits = try self.readU32();
        return @bitCast(bits);
    }

    inline fn readF64(self: *BinReader) !f64 {
        const bits = try self.readU64();
        return @bitCast(bits);
    }

    /// Zero-copy string read: returns slice into input data
    inline fn readStrZerocopy(self: *BinReader) ![]const u8 {
        const len = try self.readU32();
        try self.ensure(len);
        const slice = self.data[self.pos .. self.pos + len];
        self.pos += len;
        return slice;
    }

    fn readValue(self: *BinReader, comptime T: type) !T {
        const info = @typeInfo(T);
        switch (info) {
            .bool => return (try self.readU8()) != 0,
            .int => |i_info| {
                switch (i_info.bits) {
                    8 => {
                        if (i_info.signedness == .signed) return try self.readI8();
                        return try self.readU8();
                    },
                    16 => {
                        if (i_info.signedness == .signed) return try self.readI16();
                        return try self.readU16();
                    },
                    32 => {
                        if (i_info.signedness == .signed) return try self.readI32();
                        return @bitCast(try self.readU32());
                    },
                    64 => {
                        if (i_info.signedness == .signed) return try self.readI64();
                        return try self.readU64();
                    },
                    else => return error.UnsupportedType,
                }
            },
            .float => |f_info| {
                switch (f_info.bits) {
                    32 => return try self.readF32(),
                    64 => return try self.readF64(),
                    else => return error.UnsupportedType,
                }
            },
            .pointer => |ptr| {
                if (ptr.size == .slice and ptr.child == u8) {
                    return try self.readStrZerocopy();
                } else if (ptr.size == .slice) {
                    const count = try self.readU32();
                    if (comptime isNumericType(ptr.child)) {
                        const byte_len = count * @sizeOf(ptr.child);
                        try self.ensure(byte_len);
                        const res = try self.allocator.alloc(ptr.child, count);
                        const dst_bytes: [*]u8 = @ptrCast(res.ptr);
                        @memcpy(dst_bytes[0..byte_len], self.data[self.pos .. self.pos + byte_len]);
                        self.pos += byte_len;
                        return res;
                    }
                    const res = try self.allocator.alloc(ptr.child, count);
                    for (0..count) |i| {
                        res[i] = try self.readValue(ptr.child);
                    }
                    return res;
                } else {
                    return error.UnsupportedType;
                }
            },
            .optional => |opt| {
                const tag = try self.readU8();
                if (tag == 0) return null;
                return try self.readValue(opt.child);
            },
            .@"struct" => |s| {
                var result: T = undefined;
                inline for (s.fields) |field| {
                    @field(result, field.name) = try self.readValue(field.type);
                }
                return result;
            },
            else => return error.UnsupportedType,
        }
    }
};

// ============================================================================
// JSON helpers (minimal encode/decode for benchmarking)
// ============================================================================

pub fn jsonEncode(comptime T: type, value: T, allocator: Allocator) ![]const u8 {
    var w = Writer.init(allocator);
    errdefer w.deinit();
    try jsonSerialize(T, value, &w);
    return w.toOwnedSlice();
}

pub fn jsonEncodeVec(comptime T: type, values: []const T, allocator: Allocator) ![]const u8 {
    var w = Writer.init(allocator);
    errdefer w.deinit();
    try w.append('[');
    for (values, 0..) |v, i| {
        if (i > 0) try w.append(',');
        try jsonSerialize(T, v, &w);
    }
    try w.append(']');
    return w.toOwnedSlice();
}

fn jsonSerialize(comptime T: type, value: T, w: *Writer) !void {
    const info = @typeInfo(T);
    switch (info) {
        .bool => {
            if (value) {
                try w.appendSlice("true");
            } else {
                try w.appendSlice("false");
            }
        },
        .int => |i_info| {
            if (i_info.signedness == .signed) {
                try writeI64(w, @as(i64, @intCast(value)));
            } else {
                try writeU64(w, @as(u64, @intCast(value)));
            }
        },
        .float => {
            try writeF64(w, @as(f64, @floatCast(value)));
        },
        .pointer => |ptr| {
            if (ptr.size == .slice and ptr.child == u8) {
                try jsonWriteString(w, value);
            } else if (ptr.size == .slice) {
                try w.append('[');
                for (value, 0..) |item, i| {
                    if (i > 0) try w.append(',');
                    try jsonSerialize(ptr.child, item, w);
                }
                try w.append(']');
            } else {
                return error.UnsupportedType;
            }
        },
        .optional => |opt| {
            if (value) |v| {
                try jsonSerialize(opt.child, v, w);
            } else {
                try w.appendSlice("null");
            }
        },
        .@"struct" => |s| {
            try w.append('{');
            inline for (s.fields, 0..) |field, i| {
                if (i > 0) try w.append(',');
                try w.append('"');
                try w.appendSlice(field.name);
                try w.appendSlice("\":");
                try jsonSerialize(field.type, @field(value, field.name), w);
            }
            try w.append('}');
        },
        else => return error.UnsupportedType,
    }
}

fn jsonWriteString(w: *Writer, s: []const u8) !void {
    try w.append('"');
    var i: usize = 0;
    while (i < s.len) : (i += 1) {
        switch (s[i]) {
            '"' => try w.appendSlice("\\\""),
            '\\' => try w.appendSlice("\\\\"),
            '\n' => try w.appendSlice("\\n"),
            '\t' => try w.appendSlice("\\t"),
            '\r' => try w.appendSlice("\\r"),
            else => |c| {
                if (c < 0x20) {
                    try w.appendSlice("\\u00");
                    const HEX = "0123456789abcdef";
                    try w.append(HEX[c >> 4]);
                    try w.append(HEX[c & 0xf]);
                } else {
                    try w.append(c);
                }
            },
        }
    }
    try w.append('"');
}

/// Minimal JSON decoder
pub fn jsonDecode(comptime T: type, input: []const u8, allocator: Allocator) !T {
    var p = JsonParser{ .input = input, .pos = 0, .allocator = allocator };
    return p.parseValue(T);
}

pub fn jsonDecodeVec(comptime T: type, input: []const u8, allocator: Allocator) ![]T {
    var p = JsonParser{ .input = input, .pos = 0, .allocator = allocator };
    return p.parseJsonArray(T);
}

const JsonParser = struct {
    input: []const u8,
    pos: usize,
    allocator: Allocator,

    fn skipWs(self: *JsonParser) void {
        while (self.pos < self.input.len and (self.input[self.pos] == ' ' or self.input[self.pos] == '\t' or self.input[self.pos] == '\n' or self.input[self.pos] == '\r')) {
            self.pos += 1;
        }
    }

    fn parseValue(self: *JsonParser, comptime T: type) !T {
        const info = @typeInfo(T);
        self.skipWs();
        switch (info) {
            .bool => {
                if (self.pos + 4 <= self.input.len and mem.eql(u8, self.input[self.pos .. self.pos + 4], "true")) {
                    self.pos += 4;
                    return true;
                }
                if (self.pos + 5 <= self.input.len and mem.eql(u8, self.input[self.pos .. self.pos + 5], "false")) {
                    self.pos += 5;
                    return false;
                }
                return error.InvalidBool;
            },
            .int => return self.parseJsonInt(T),
            .float => return self.parseJsonFloat(T),
            .pointer => |ptr| {
                if (ptr.size == .slice and ptr.child == u8) {
                    return self.parseJsonString();
                } else if (ptr.size == .slice) {
                    return self.parseJsonArray(ptr.child);
                } else {
                    return error.UnsupportedType;
                }
            },
            .optional => |opt| {
                if (self.pos + 4 <= self.input.len and mem.eql(u8, self.input[self.pos .. self.pos + 4], "null")) {
                    self.pos += 4;
                    return null;
                }
                return try self.parseValue(opt.child);
            },
            .@"struct" => return self.parseJsonObject(T),
            else => return error.UnsupportedType,
        }
    }

    fn parseJsonInt(self: *JsonParser, comptime T: type) !T {
        self.skipWs();
        var negative = false;
        if (self.pos < self.input.len and self.input[self.pos] == '-') {
            negative = true;
            self.pos += 1;
        }
        var val: i128 = 0;
        while (self.pos < self.input.len and std.ascii.isDigit(self.input[self.pos])) {
            val = val * 10 + @as(i128, self.input[self.pos] - '0');
            self.pos += 1;
        }
        if (self.pos < self.input.len and self.input[self.pos] == '.') {
            self.pos += 1;
            while (self.pos < self.input.len and std.ascii.isDigit(self.input[self.pos])) {
                self.pos += 1;
            }
        }
        if (negative) val = -val;
        return @intCast(val);
    }

    fn parseJsonFloat(self: *JsonParser, comptime T: type) !T {
        self.skipWs();
        const start = self.pos;
        if (self.pos < self.input.len and (self.input[self.pos] == '-' or self.input[self.pos] == '+')) {
            self.pos += 1;
        }
        while (self.pos < self.input.len and std.ascii.isDigit(self.input[self.pos])) {
            self.pos += 1;
        }
        if (self.pos < self.input.len and self.input[self.pos] == '.') {
            self.pos += 1;
            while (self.pos < self.input.len and std.ascii.isDigit(self.input[self.pos])) {
                self.pos += 1;
            }
        }
        if (self.pos < self.input.len and (self.input[self.pos] == 'e' or self.input[self.pos] == 'E')) {
            self.pos += 1;
            if (self.pos < self.input.len and (self.input[self.pos] == '-' or self.input[self.pos] == '+')) {
                self.pos += 1;
            }
            while (self.pos < self.input.len and std.ascii.isDigit(self.input[self.pos])) {
                self.pos += 1;
            }
        }
        return std.fmt.parseFloat(T, self.input[start..self.pos]) catch return error.InvalidNumber;
    }

    fn parseJsonString(self: *JsonParser) ![]const u8 {
        self.skipWs();
        if (self.pos >= self.input.len or self.input[self.pos] != '"') return error.InvalidFormat;
        self.pos += 1;
        var result: std.ArrayList(u8) = .{};
        errdefer result.deinit(self.allocator);
        while (self.pos < self.input.len) {
            const b = self.input[self.pos];
            if (b == '"') {
                self.pos += 1;
                return result.toOwnedSlice(self.allocator);
            }
            if (b == '\\') {
                self.pos += 1;
                if (self.pos >= self.input.len) return error.InvalidEscape;
                const esc = self.input[self.pos];
                self.pos += 1;
                switch (esc) {
                    '"' => try result.append(self.allocator, '"'),
                    '\\' => try result.append(self.allocator, '\\'),
                    'n' => try result.append(self.allocator, '\n'),
                    't' => try result.append(self.allocator, '\t'),
                    'r' => try result.append(self.allocator, '\r'),
                    'u' => {
                        if (self.pos + 4 > self.input.len) return error.InvalidEscape;
                        const hex = self.input[self.pos .. self.pos + 4];
                        const cp = std.fmt.parseInt(u21, hex, 16) catch return error.InvalidEscape;
                        var utf8_buf: [4]u8 = undefined;
                        const len = std.unicode.utf8Encode(cp, &utf8_buf) catch return error.InvalidEscape;
                        try result.appendSlice(self.allocator, utf8_buf[0..len]);
                        self.pos += 4;
                    },
                    else => try result.append(self.allocator, esc),
                }
            } else {
                try result.append(self.allocator, b);
                self.pos += 1;
            }
        }
        return error.UnclosedString;
    }

    fn parseJsonArray(self: *JsonParser, comptime Child: type) ![]Child {
        self.skipWs();
        if (self.pos >= self.input.len or self.input[self.pos] != '[') return error.InvalidFormat;
        self.pos += 1;
        var result: std.ArrayList(Child) = .{};
        errdefer result.deinit(self.allocator);
        self.skipWs();
        if (self.pos < self.input.len and self.input[self.pos] == ']') {
            self.pos += 1;
            return result.toOwnedSlice(self.allocator);
        }
        while (true) {
            self.skipWs();
            const item = try self.parseValue(Child);
            try result.append(self.allocator, item);
            self.skipWs();
            if (self.pos >= self.input.len) return error.Eof;
            if (self.input[self.pos] == ']') {
                self.pos += 1;
                break;
            }
            if (self.input[self.pos] == ',') {
                self.pos += 1;
            }
        }
        return result.toOwnedSlice(self.allocator);
    }

    fn parseJsonObject(self: *JsonParser, comptime T: type) !T {
        const s = @typeInfo(T).@"struct";
        self.skipWs();
        if (self.pos >= self.input.len or self.input[self.pos] != '{') return error.InvalidFormat;
        self.pos += 1;
        var result: T = undefined;
        inline for (s.fields) |field| {
            if (@typeInfo(field.type) == .optional) {
                @field(result, field.name) = null;
            }
        }
        self.skipWs();
        if (self.pos < self.input.len and self.input[self.pos] == '}') {
            self.pos += 1;
            return result;
        }
        while (true) {
            self.skipWs();
            const key = try self.parseJsonString();
            defer self.allocator.free(key);
            self.skipWs();
            if (self.pos < self.input.len and self.input[self.pos] == ':') {
                self.pos += 1;
            }
            self.skipWs();
            var matched = false;
            inline for (s.fields) |field| {
                if (mem.eql(u8, key, field.name)) {
                    @field(result, field.name) = try self.parseValue(field.type);
                    matched = true;
                }
            }
            if (!matched) {
                try self.skipJsonValue();
            }
            self.skipWs();
            if (self.pos >= self.input.len) break;
            if (self.input[self.pos] == '}') {
                self.pos += 1;
                break;
            }
            if (self.input[self.pos] == ',') {
                self.pos += 1;
            }
        }
        return result;
    }

    fn skipJsonValue(self: *JsonParser) !void {
        self.skipWs();
        if (self.pos >= self.input.len) return;
        const b = self.input[self.pos];
        if (b == '"') {
            const s = try self.parseJsonString();
            self.allocator.free(s);
            return;
        }
        if (b == '{') {
            self.pos += 1;
            var depth: usize = 1;
            while (self.pos < self.input.len and depth > 0) {
                if (self.input[self.pos] == '{') depth += 1;
                if (self.input[self.pos] == '}') depth -= 1;
                if (self.input[self.pos] == '"') {
                    const s = try self.parseJsonString();
                    self.allocator.free(s);
                    continue;
                }
                self.pos += 1;
            }
            return;
        }
        if (b == '[') {
            self.pos += 1;
            var depth: usize = 1;
            while (self.pos < self.input.len and depth > 0) {
                if (self.input[self.pos] == '[') depth += 1;
                if (self.input[self.pos] == ']') depth -= 1;
                if (self.input[self.pos] == '"') {
                    const s = try self.parseJsonString();
                    self.allocator.free(s);
                    continue;
                }
                self.pos += 1;
            }
            return;
        }
        while (self.pos < self.input.len and self.input[self.pos] != ',' and self.input[self.pos] != '}' and self.input[self.pos] != ']') {
            self.pos += 1;
        }
    }
};

// ============================================================================
// Tests
// ============================================================================

test "encode and decode basic struct" {
    const allocator = std.testing.allocator;
    const User = struct {
        id: i64,
        name: []const u8,
        active: bool,
    };
    const user = User{ .id = 1, .name = "Alice", .active = true };
    const encoded = try encode(User, user, allocator);
    defer allocator.free(encoded);
    const decoded = try decode(User, encoded, allocator);
    defer allocator.free(decoded.name);
    try std.testing.expectEqual(@as(i64, 1), decoded.id);
    try std.testing.expectEqualStrings("Alice", decoded.name);
    try std.testing.expect(decoded.active);
}

test "binary roundtrip" {
    const allocator = std.testing.allocator;
    const User = struct {
        id: i64,
        name: []const u8,
        score: f64,
        active: bool,
    };
    const user = User{ .id = 42, .name = "Alice", .score = 9.5, .active = true };
    const bin = try encodeBinary(User, user, allocator);
    defer allocator.free(bin);
    const decoded = try decodeBinary(User, bin, allocator);
    try std.testing.expectEqual(@as(i64, 42), decoded.id);
    try std.testing.expectEqualStrings("Alice", decoded.name);
    try std.testing.expectEqual(@as(f64, 9.5), decoded.score);
    try std.testing.expect(decoded.active);
}

test "binary vec roundtrip" {
    const allocator = std.testing.allocator;
    const User = struct {
        id: i64,
        name: []const u8,
        active: bool,
    };
    const users = [_]User{
        .{ .id = 1, .name = "Alice", .active = true },
        .{ .id = 2, .name = "Bob", .active = false },
    };
    const bin = try encodeBinaryVec(User, &users, allocator);
    defer allocator.free(bin);
    const decoded = try decodeBinaryVec(User, bin, allocator);
    defer allocator.free(decoded);
    try std.testing.expectEqual(@as(usize, 2), decoded.len);
    try std.testing.expectEqualStrings("Alice", decoded[0].name);
    try std.testing.expectEqualStrings("Bob", decoded[1].name);
}

test "SIMD has special chars" {
    try std.testing.expect(!simdHasSpecialChars("hello world"));
    try std.testing.expect(simdHasSpecialChars("hello,world"));
    try std.testing.expect(simdHasSpecialChars("hello(world"));
    try std.testing.expect(simdHasSpecialChars("hello\"world"));
    try std.testing.expect(!simdHasSpecialChars("abcdefghijklmnop"));
    try std.testing.expect(simdHasSpecialChars("abcdefghijklmnop,"));
}

test "SIMD skip whitespace" {
    try std.testing.expectEqual(@as(usize, 3), simdSkipWhitespace("   hello", 0));
    try std.testing.expectEqual(@as(usize, 0), simdSkipWhitespace("hello", 0));
    try std.testing.expectEqual(@as(usize, 18), simdSkipWhitespace("                  hello", 0));
}
