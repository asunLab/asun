package io.ason;

import java.nio.charset.StandardCharsets;
import java.util.Arrays;

/**
 * Resizable byte buffer with ThreadLocal reuse for zero-allocation ASON encoding.
 * Inspired by fastjson's SerializeWriter ThreadLocal buffer pattern.
 */
final class ByteBuffer {
    private static final int INITIAL_SIZE = 4096;
    private static final int MAX_REUSE_SIZE = 1024 * 1024; // 1MB

    private static final ThreadLocal<byte[]> BUF_LOCAL =
            ThreadLocal.withInitial(() -> new byte[INITIAL_SIZE]);

    byte[] data;
    int len;
    boolean hasNonAscii; // track if any non-ASCII byte was written

    ByteBuffer() {
        data = BUF_LOCAL.get();
        len = 0;
        hasNonAscii = false;
    }

    ByteBuffer(int hint) {
        byte[] cached = BUF_LOCAL.get();
        data = (hint <= cached.length) ? cached : new byte[hint];
        len = 0;
        hasNonAscii = false;
    }

    /** Return buffer to ThreadLocal pool for reuse. */
    void close() {
        if (data.length <= MAX_REUSE_SIZE) {
            BUF_LOCAL.set(data);
        }
    }

    void ensureCapacity(int additional) {
        int required = len + additional;
        if (required > data.length) {
            int newCap = Math.max(data.length << 1, required); // 2x growth
            data = Arrays.copyOf(data, newCap);
        }
    }

    void append(byte b) {
        if (len >= data.length) ensureCapacity(1);
        data[len++] = b;
    }

    void append(char c) {
        if (len >= data.length) ensureCapacity(1);
        data[len++] = (byte) c;
    }

    void appendBytes(byte[] src, int off, int length) {
        ensureCapacity(length);
        System.arraycopy(src, off, data, len, length);
        len += length;
    }

    void appendStr(String s) {
        byte[] bytes = s.getBytes(StandardCharsets.UTF_8);
        if (!hasNonAscii) {
            for (byte b : bytes) { if (b < 0) { hasNonAscii = true; break; } }
        }
        appendBytes(bytes, 0, bytes.length);
    }

    /** Write known-ASCII string by char iteration — avoids byte[] allocation. */
    void appendAscii(String s) {
        int slen = s.length();
        ensureCapacity(slen);
        for (int i = 0; i < slen; i++) {
            data[len++] = (byte) s.charAt(i);
        }
    }

    /** Write string chars directly to buffer — assumes all chars < 128 (ASCII). */
    void appendCharsAsBytes(String s, int slen) {
        ensureCapacity(slen);
        for (int i = 0; i < slen; i++) {
            data[len++] = (byte) s.charAt(i);
        }
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
        return new String(data, 0, len,
                hasNonAscii ? StandardCharsets.UTF_8 : StandardCharsets.ISO_8859_1);
    }

    /** Return string and recycle buffer to ThreadLocal pool. */
    String toStringUtf8AndClose() {
        String s = new String(data, 0, len,
                hasNonAscii ? StandardCharsets.UTF_8 : StandardCharsets.ISO_8859_1);
        close();
        return s;
    }

    /** Return bytes and recycle buffer to ThreadLocal pool. */
    byte[] toBytesAndClose() {
        byte[] result = Arrays.copyOf(data, len);
        close();
        return result;
    }
}
