package io.ason;

import java.nio.charset.StandardCharsets;
import java.util.Arrays;

/**
 * Minimal resizable byte buffer for zero-copy ASON encoding.
 * Avoids StringBuilder and intermediate String allocations.
 */
final class ByteBuffer {
    byte[] data;
    int len;

    ByteBuffer(int initialCapacity) {
        data = new byte[initialCapacity];
        len = 0;
    }

    void ensureCapacity(int additional) {
        int required = len + additional;
        if (required > data.length) {
            int newCap = Math.max(data.length * 2, required);
            data = Arrays.copyOf(data, newCap);
        }
    }

    void append(byte b) {
        ensureCapacity(1);
        data[len++] = b;
    }

    void append(char c) {
        ensureCapacity(1);
        data[len++] = (byte) c;
    }

    void appendBytes(byte[] src, int off, int length) {
        ensureCapacity(length);
        System.arraycopy(src, off, data, len, length);
        len += length;
    }

    void appendStr(String s) {
        byte[] bytes = s.getBytes(StandardCharsets.UTF_8);
        appendBytes(bytes, 0, bytes.length);
    }

    void appendLEU16(int v) {
        ensureCapacity(2);
        data[len++] = (byte) (v & 0xFF);
        data[len++] = (byte) ((v >> 8) & 0xFF);
    }

    void appendLEU32(int v) {
        ensureCapacity(4);
        data[len++] = (byte) (v & 0xFF);
        data[len++] = (byte) ((v >> 8) & 0xFF);
        data[len++] = (byte) ((v >> 16) & 0xFF);
        data[len++] = (byte) ((v >> 24) & 0xFF);
    }

    void appendLEU64(long v) {
        ensureCapacity(8);
        data[len++] = (byte) (v & 0xFF);
        data[len++] = (byte) ((v >> 8) & 0xFF);
        data[len++] = (byte) ((v >> 16) & 0xFF);
        data[len++] = (byte) ((v >> 24) & 0xFF);
        data[len++] = (byte) ((v >> 32) & 0xFF);
        data[len++] = (byte) ((v >> 40) & 0xFF);
        data[len++] = (byte) ((v >> 48) & 0xFF);
        data[len++] = (byte) ((v >> 56) & 0xFF);
    }

    byte[] toBytes() {
        return Arrays.copyOf(data, len);
    }

    String toStringUtf8() {
        return new String(data, 0, len, StandardCharsets.UTF_8);
    }
}
