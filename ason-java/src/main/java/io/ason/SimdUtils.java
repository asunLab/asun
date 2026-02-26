package io.ason;

import jdk.incubator.vector.ByteVector;
import jdk.incubator.vector.VectorMask;
import jdk.incubator.vector.VectorOperators;
import jdk.incubator.vector.VectorSpecies;

/**
 * SIMD-accelerated utilities for ASON text processing using Java Vector API.
 * Uses 256-bit (32-byte) or 128-bit (16-byte) SIMD lanes for:
 *   - Special character detection (comma, parens, brackets, quote, backslash, control chars)
 *   - Whitespace skipping
 *   - Escape character scanning
 *   - Bulk memory copy
 */
public final class SimdUtils {

    static final VectorSpecies<Byte> SPECIES;
    static final int LANES;

    static {
        VectorSpecies<Byte> preferred = ByteVector.SPECIES_PREFERRED;
        int bits = preferred.vectorBitSize();
        if (bits >= 256) {
            SPECIES = ByteVector.SPECIES_256;
        } else {
            SPECIES = ByteVector.SPECIES_128;
        }
        LANES = SPECIES.length();
    }

    private SimdUtils() {}

    // --- Special character detection ---

    private static final byte CTRL_MAX = 0x1f;
    private static final byte COMMA = ',';
    private static final byte LPAREN = '(';
    private static final byte RPAREN = ')';
    private static final byte LBRACKET = '[';
    private static final byte RBRACKET = ']';
    private static final byte QUOTE = '"';
    private static final byte BACKSLASH = '\\';

    /**
     * SIMD scan: does the byte array contain any ASON special character?
     * Special = control (<=0x1f), comma, parens, brackets, quote, backslash.
     */
    public static boolean hasSpecialChars(byte[] data, int off, int len) {
        int i = 0;
        int end = off + len;

        // SIMD path
        while (i + LANES <= len) {
            ByteVector chunk = ByteVector.fromArray(SPECIES, data, off + i);
            VectorMask<Byte> ctrl = chunk.compare(VectorOperators.LE, CTRL_MAX);
            VectorMask<Byte> comma = chunk.compare(VectorOperators.EQ, COMMA);
            VectorMask<Byte> lp = chunk.compare(VectorOperators.EQ, LPAREN);
            VectorMask<Byte> rp = chunk.compare(VectorOperators.EQ, RPAREN);
            VectorMask<Byte> lb = chunk.compare(VectorOperators.EQ, LBRACKET);
            VectorMask<Byte> rb = chunk.compare(VectorOperators.EQ, RBRACKET);
            VectorMask<Byte> qt = chunk.compare(VectorOperators.EQ, QUOTE);
            VectorMask<Byte> bs = chunk.compare(VectorOperators.EQ, BACKSLASH);
            VectorMask<Byte> any = ctrl.or(comma).or(lp).or(rp).or(lb).or(rb).or(qt).or(bs);
            if (any.anyTrue()) return true;
            i += LANES;
        }

        // Scalar tail
        for (int j = off + i; j < end; j++) {
            if (isSpecial(data[j])) return true;
        }
        return false;
    }

    private static boolean isSpecial(byte b) {
        return b <= CTRL_MAX || b == COMMA || b == LPAREN || b == RPAREN
                || b == LBRACKET || b == RBRACKET || b == QUOTE || b == BACKSLASH;
    }

    /**
     * SIMD scan: find first byte that needs escaping (quote, backslash, or control char).
     * Returns the index relative to 'off', or 'len' if none found.
     */
    public static int findEscape(byte[] data, int off, int len) {
        int i = 0;
        while (i + LANES <= len) {
            ByteVector chunk = ByteVector.fromArray(SPECIES, data, off + i);
            VectorMask<Byte> ctrl = chunk.compare(VectorOperators.LE, CTRL_MAX);
            VectorMask<Byte> qt = chunk.compare(VectorOperators.EQ, QUOTE);
            VectorMask<Byte> bs = chunk.compare(VectorOperators.EQ, BACKSLASH);
            VectorMask<Byte> any = ctrl.or(qt).or(bs);
            if (any.anyTrue()) {
                return i + any.firstTrue();
            }
            i += LANES;
        }
        for (int j = 0; j < len - i; j++) {
            byte b = data[off + i + j];
            if (b <= CTRL_MAX || b == QUOTE || b == BACKSLASH) return i + j;
        }
        return len;
    }

    /**
     * SIMD scan: find first delimiter for plain string parsing (, ) ] \).
     */
    public static int findPlainDelimiter(byte[] data, int off, int len) {
        int i = 0;
        while (i + LANES <= len) {
            ByteVector chunk = ByteVector.fromArray(SPECIES, data, off + i);
            VectorMask<Byte> comma = chunk.compare(VectorOperators.EQ, COMMA);
            VectorMask<Byte> rp = chunk.compare(VectorOperators.EQ, RPAREN);
            VectorMask<Byte> rb = chunk.compare(VectorOperators.EQ, RBRACKET);
            VectorMask<Byte> bs = chunk.compare(VectorOperators.EQ, BACKSLASH);
            VectorMask<Byte> any = comma.or(rp).or(rb).or(bs);
            if (any.anyTrue()) {
                return i + any.firstTrue();
            }
            i += LANES;
        }
        for (int j = 0; j < len - i; j++) {
            byte b = data[off + i + j];
            if (b == COMMA || b == RPAREN || b == RBRACKET || b == BACKSLASH) return i + j;
        }
        return len;
    }

    /**
     * SIMD scan: skip whitespace (space, tab, newline, cr). Returns new offset.
     */
    public static int skipWhitespace(byte[] data, int off, int len) {
        int i = off;
        while (i + LANES <= len) {
            ByteVector chunk = ByteVector.fromArray(SPECIES, data, i);
            VectorMask<Byte> sp = chunk.compare(VectorOperators.EQ, (byte) ' ');
            VectorMask<Byte> tab = chunk.compare(VectorOperators.EQ, (byte) '\t');
            VectorMask<Byte> nl = chunk.compare(VectorOperators.EQ, (byte) '\n');
            VectorMask<Byte> cr = chunk.compare(VectorOperators.EQ, (byte) '\r');
            VectorMask<Byte> ws = sp.or(tab).or(nl).or(cr);
            if (ws.allTrue()) {
                i += LANES;
                continue;
            }
            if (!ws.anyTrue()) {
                return i;
            }
            // Mixed: find first non-whitespace
            VectorMask<Byte> nonWs = ws.not();
            return i + nonWs.firstTrue();
        }
        // Scalar tail
        while (i < len) {
            byte b = data[i];
            if (b != ' ' && b != '\t' && b != '\n' && b != '\r') return i;
            i++;
        }
        return i;
    }

    /**
     * SIMD scan: find first quote or backslash for quoted string parsing.
     */
    public static int findQuoteOrBackslash(byte[] data, int off, int len) {
        int i = 0;
        while (i + LANES <= len) {
            ByteVector chunk = ByteVector.fromArray(SPECIES, data, off + i);
            VectorMask<Byte> qt = chunk.compare(VectorOperators.EQ, QUOTE);
            VectorMask<Byte> bs = chunk.compare(VectorOperators.EQ, BACKSLASH);
            VectorMask<Byte> any = qt.or(bs);
            if (any.anyTrue()) {
                return i + any.firstTrue();
            }
            i += LANES;
        }
        for (int j = 0; j < len - i; j++) {
            byte b = data[off + i + j];
            if (b == QUOTE || b == BACKSLASH) return i + j;
        }
        return len;
    }
}
