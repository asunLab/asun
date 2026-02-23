package io.ason;

import io.ason.ClassMeta.FieldMeta;
import java.lang.reflect.*;
import java.nio.charset.StandardCharsets;
import java.util.*;

/**
 * ASON text decoder — MethodHandle-based, zero-copy where possible.
 * Uses ClassMeta for fast field access via invokeExact.
 */
final class AsonDecoder {
    private final byte[] input;
    private int pos;

    // Direct double parsing: POW10[i] = 10^i
    private static final double[] POW10 = {
        1e0, 1e1, 1e2, 1e3, 1e4, 1e5, 1e6, 1e7, 1e8, 1e9,
        1e10, 1e11, 1e12, 1e13, 1e14, 1e15, 1e16, 1e17, 1e18
    };

    AsonDecoder(byte[] input) {
        this.input = input;
        this.pos = 0;
    }

    // ========================================================================
    // Public entry points
    // ========================================================================

    <T> T decodeSingle(Class<T> clazz) {
        skipWhitespaceAndComments();
        if (pos < input.length && input[pos] == '[') {
            List<T> list = decodeListInternal(clazz);
            if (list.isEmpty()) throw new AsonException("Empty array, cannot decode single");
            return list.getFirst();
        }

        ClassMeta meta = ClassMeta.of(clazz);
        if (pos < input.length && input[pos] == '{') {
            skipSchema();
            skipWhitespaceAndComments();
            expect(':');
            skipWhitespaceAndComments();
        }
        return parseTuple(clazz, meta);
    }

    <T> List<T> decodeList(Class<T> clazz) {
        skipWhitespaceAndComments();
        return decodeListInternal(clazz);
    }

    @SuppressWarnings("unchecked")
    private <T> List<T> decodeListInternal(Class<T> clazz) {
        ClassMeta meta = ClassMeta.of(clazz);
        if (pos < input.length && input[pos] == '[') {
            pos++;
            skipSchema();
            skipWs();
            expect(']');
            skipWs();
            expect(':');
            skipWs();

            // Pre-size ArrayList based on remaining input / estimated tuple size
            int estimatedSize = Math.max(16, (input.length - pos) / (meta.fields.length * 5 + 3));
            List<T> result = new ArrayList<>(estimatedSize);
            while (pos < input.length) {
                if (pos < input.length && input[pos] <= ' ') skipWs();
                if (pos >= input.length || input[pos] != '(') break;
                result.add(parseTuple(clazz, meta));
                if (pos < input.length && input[pos] == ',') pos++;
            }
            return result;
        }
        throw new AsonException("Expected '[' for list format");
    }

    private void skipSchema() {
        expect('{');
        int depth = 1;
        while (pos < input.length && depth > 0) {
            byte b = input[pos++];
            if (b == '{') depth++;
            else if (b == '}') depth--;
        }
    }

    // ========================================================================
    // Tuple parsing — type-tag dispatched with invokeExact setters
    // ========================================================================

    @SuppressWarnings("unchecked")
    private <T> T parseTuple(Class<T> clazz, ClassMeta meta) {
        expect('(');
        Object obj = meta.newInstance();
        FieldMeta[] fields = meta.fields;
        final byte[] inp = this.input; // local copy for JIT

        for (int i = 0; i < fields.length; i++) {
            if (i > 0) {
                // Fast: check comma without ws skip first (compact format has no ws)
                if (pos < inp.length && inp[pos] == ',') {
                    pos++;
                } else {
                    skipWs();
                    if (pos < inp.length && inp[pos] == ',') pos++;
                    else if (pos < inp.length && inp[pos] == ')') break;
                    else break;
                }
            }

            // Skip whitespace before value (usually no-op for compact format)
            if (pos < inp.length && inp[pos] <= ' ') skipWs();

            FieldMeta fm = fields[i];
            // Handle empty values
            if (pos >= inp.length) continue;
            byte b = inp[pos];
            if (b == ',' || b == ')' || b == ']') continue;

            switch (fm.typeTag) {
                case FieldMeta.T_BOOLEAN -> {
                    boolean v = parseBool();
                    if (fm.isPrimitive) fm.setBoolean(obj, v);
                    else fm.set(obj, v);
                }
                case FieldMeta.T_INT -> {
                    int v = (int) parseLong();
                    if (fm.isPrimitive) fm.setInt(obj, v);
                    else fm.set(obj, v);
                }
                case FieldMeta.T_LONG -> {
                    long v = parseLong();
                    if (fm.isPrimitive) fm.setLong(obj, v);
                    else fm.set(obj, v);
                }
                case FieldMeta.T_SHORT -> {
                    short v = (short) parseLong();
                    if (fm.isPrimitive) fm.setShort(obj, v);
                    else fm.set(obj, v);
                }
                case FieldMeta.T_BYTE -> {
                    byte v = (byte) parseLong();
                    if (fm.isPrimitive) fm.setByte(obj, v);
                    else fm.set(obj, v);
                }
                case FieldMeta.T_FLOAT -> {
                    float v = (float) parseDoubleDirect();
                    if (fm.isPrimitive) fm.setFloat(obj, v);
                    else fm.set(obj, v);
                }
                case FieldMeta.T_DOUBLE -> {
                    double v = parseDoubleDirect();
                    if (fm.isPrimitive) fm.setDouble(obj, v);
                    else fm.set(obj, v);
                }
                case FieldMeta.T_CHAR -> {
                    String s = parseStringValue();
                    char v = s.isEmpty() ? '\0' : s.charAt(0);
                    if (fm.isPrimitive) fm.setChar(obj, v);
                    else fm.set(obj, v);
                }
                case FieldMeta.T_STRING -> fm.set(obj, parseStringValue());
                case FieldMeta.T_OPTIONAL -> {
                    if (atValueEnd()) {
                        fm.set(obj, Optional.empty());
                    } else {
                        Type innerType = fm.elemType != null ? fm.elemType : Object.class;
                        Class<?> innerClass = fm.elemClass != null ? fm.elemClass : Object.class;
                        Object inner = parseFieldValue(innerClass, innerType);
                        fm.set(obj, Optional.ofNullable(inner));
                    }
                }
                case FieldMeta.T_LIST -> fm.set(obj, parseListField(fm));
                case FieldMeta.T_MAP -> fm.set(obj, parseMap(fm.genericType));
                default -> {
                    // T_STRUCT
                    if (inp[pos] == '(') {
                        ClassMeta nested = fm.nestedMeta != null ? fm.nestedMeta : ClassMeta.of(fm.type);
                        fm.set(obj, parseTuple(fm.type, nested));
                    } else {
                        fm.set(obj, parseFieldValue(fm.type, fm.genericType));
                    }
                }
            }
        }
        skipWs();
        if (pos < inp.length && inp[pos] == ')') pos++;
        return (T) obj;
    }

    // ========================================================================
    // Value parsing (for generic/nested contexts)
    // ========================================================================

    @SuppressWarnings({"unchecked", "rawtypes"})
    private Object parseFieldValue(Class<?> type, Type genericType) {
        skipWs();
        if (pos >= input.length) return null;

        if (type == Optional.class) {
            if (atValueEnd()) return Optional.empty();
            Type innerType = Object.class;
            if (genericType instanceof ParameterizedType pt) {
                innerType = pt.getActualTypeArguments()[0];
            }
            Class<?> innerClass = (innerType instanceof Class<?> c) ? c : Object.class;
            Object inner = parseFieldValue(innerClass, innerType);
            return Optional.ofNullable(inner);
        }

        if (atValueEnd()) return defaultValue(type);

        byte b = input[pos];

        if (type == boolean.class || type == Boolean.class) return parseBool();
        if (type == int.class || type == Integer.class) return (int) parseLong();
        if (type == long.class || type == Long.class) return parseLong();
        if (type == short.class || type == Short.class) return (short) parseLong();
        if (type == byte.class || type == Byte.class) return (byte) parseLong();
        if (type == float.class || type == Float.class) return (float) parseDoubleDirect();
        if (type == double.class || type == Double.class) return parseDoubleDirect();
        if (type == String.class) return parseStringValue();
        if (type == char.class || type == Character.class) {
            String s = parseStringValue();
            return s.isEmpty() ? '\0' : s.charAt(0);
        }
        if (List.class.isAssignableFrom(type)) return parseList(genericType);
        if (Map.class.isAssignableFrom(type)) return parseMap(genericType);

        if (b == '(') {
            ClassMeta nested = ClassMeta.of(type);
            return parseTuple(type, nested);
        }

        return parseStringValue();
    }

    // ========================================================================
    // Primitive parsers
    // ========================================================================

    private boolean parseBool() {
        if (pos + 4 <= input.length && input[pos] == 't' && input[pos + 1] == 'r'
            && input[pos + 2] == 'u' && input[pos + 3] == 'e') {
            pos += 4;
            return true;
        }
        if (pos + 5 <= input.length && input[pos] == 'f' && input[pos + 1] == 'a'
            && input[pos + 2] == 'l' && input[pos + 3] == 's' && input[pos + 4] == 'e') {
            pos += 5;
            return false;
        }
        throw new AsonException("Expected boolean at pos " + pos);
    }

    private long parseLong() {
        boolean negative = pos < input.length && input[pos] == '-';
        if (negative) pos++;
        long val = 0;
        int digits = 0;
        while (pos < input.length) {
            int d = input[pos] - '0';
            if (d < 0 || d > 9) break;
            val = val * 10 + d;
            pos++;
            digits++;
        }
        if (digits == 0) throw new AsonException("Expected integer at pos " + pos);
        return negative ? -val : val;
    }

    /**
     * Direct double parsing: avoids String allocation for simple decimals.
     * Falls back to Double.parseDouble for scientific notation.
     */
    private double parseDoubleDirect() {
        int start = pos;
        boolean negative = false;
        if (pos < input.length && input[pos] == '-') { negative = true; pos++; }

        long intPart = 0;
        int intDigits = 0;
        while (pos < input.length) {
            int d = input[pos] - '0';
            if (d < 0 || d > 9) break;
            intPart = intPart * 10 + d;
            pos++;
            intDigits++;
        }

        if (pos < input.length && input[pos] == '.') {
            pos++;
            long fracVal = 0;
            int fracDigits = 0;
            while (pos < input.length) {
                int d = input[pos] - '0';
                if (d < 0 || d > 9) break;
                fracVal = fracVal * 10 + d;
                pos++;
                fracDigits++;
            }
            // Check for scientific notation
            if (pos < input.length && (input[pos] == 'e' || input[pos] == 'E')) {
                return parseDoubleFallback(start);
            }
            if (fracDigits > 0 && fracDigits < POW10.length) {
                double v = intPart + fracVal / POW10[fracDigits];
                return negative ? -v : v;
            }
            // Fallback for very long fractions
            return parseDoubleFallback(start);
        }

        // Check for scientific notation
        if (pos < input.length && (input[pos] == 'e' || input[pos] == 'E')) {
            return parseDoubleFallback(start);
        }

        if (intDigits == 0) throw new AsonException("Expected number at pos " + pos);
        return negative ? -(double) intPart : (double) intPart;
    }

    private double parseDoubleFallback(int start) {
        // Already advanced pos past some digits; continue scanning
        if (pos < input.length && (input[pos] == 'e' || input[pos] == 'E')) {
            pos++;
            if (pos < input.length && (input[pos] == '+' || input[pos] == '-')) pos++;
            while (pos < input.length && input[pos] >= '0' && input[pos] <= '9') pos++;
        }
        return Double.parseDouble(new String(input, start, pos - start, StandardCharsets.US_ASCII));
    }

    // ========================================================================
    // String parsing
    // ========================================================================

    private String parseStringValue() {
        // Caller already skipped whitespace
        if (pos >= input.length) return "";
        byte b = input[pos];
        if (b == ',' || b == ')' || b == ']') return "";
        if (b == '"') return parseQuotedString();
        return parsePlainString();
    }

    private String parseQuotedString() {
        pos++; // skip '"'
        int start = pos;

        int hit = SimdUtils.findQuoteOrBackslash(input, pos, input.length - pos);
        int hitPos = pos + hit;
        if (hitPos < input.length && input[hitPos] == '"') {
            // Fast path: no escapes — check if ASCII-only for faster String construction
            boolean ascii = true;
            for (int i = start; i < hitPos; i++) {
                if (input[i] < 0) { ascii = false; break; }
            }
            String s = new String(input, start, hitPos - start,
                ascii ? StandardCharsets.ISO_8859_1 : StandardCharsets.UTF_8);
            pos = hitPos + 1;
            return s;
        }

        StringBuilder sb = new StringBuilder(hitPos - start + 16);
        if (hitPos > start) {
            sb.append(new String(input, start, hitPos - start, StandardCharsets.UTF_8));
        }
        pos = hitPos;

        while (pos < input.length) {
            byte b = input[pos];
            if (b == '"') {
                pos++;
                return sb.toString();
            }
            if (b == '\\') {
                pos++;
                if (pos >= input.length) throw new AsonException("Unclosed string");
                byte esc = input[pos++];
                switch (esc) {
                    case '"' -> sb.append('"');
                    case '\\' -> sb.append('\\');
                    case 'n' -> sb.append('\n');
                    case 't' -> sb.append('\t');
                    case 'r' -> sb.append('\r');
                    case ',' -> sb.append(',');
                    case '(' -> sb.append('(');
                    case ')' -> sb.append(')');
                    case '[' -> sb.append('[');
                    case ']' -> sb.append(']');
                    case 'u' -> {
                        if (pos + 4 > input.length) throw new AsonException("Invalid unicode escape");
                        String hex = new String(input, pos, 4, StandardCharsets.US_ASCII);
                        sb.append((char) Integer.parseInt(hex, 16));
                        pos += 4;
                    }
                    default -> throw new AsonException("Invalid escape: \\" + (char) esc);
                }
            } else {
                int nextHit = SimdUtils.findQuoteOrBackslash(input, pos, input.length - pos);
                int nextPos = pos + nextHit;
                if (nextPos > pos) {
                    sb.append(new String(input, pos, nextPos - pos, StandardCharsets.UTF_8));
                    pos = nextPos;
                } else {
                    sb.append((char) b);
                    pos++;
                }
            }
        }
        throw new AsonException("Unclosed string");
    }

    private String parsePlainString() {
        int start = pos;
        boolean hasEscape = false;
        boolean hasNonAscii = false;
        while (pos < input.length) {
            byte b = input[pos];
            if (b == ',' || b == ')' || b == ']') break;
            if (b == '\\') { hasEscape = true; pos += 2; }
            else {
                if (b < 0) hasNonAscii = true; // high bit set = non-ASCII UTF-8
                pos++;
            }
        }
        // Trim trailing whitespace (skip leading — already skipped by caller)
        int end = pos;
        while (end > start && (input[end - 1] == ' ' || input[end - 1] == '\t')) end--;
        // Skip leading whitespace
        int s = start;
        while (s < end && (input[s] == ' ' || input[s] == '\t')) s++;

        String raw = new String(input, s, end - s,
            hasNonAscii ? StandardCharsets.UTF_8 : StandardCharsets.ISO_8859_1);
        if (hasEscape) return unescapePlain(raw);
        return raw;
    }

    // ========================================================================
    // Collection parsing
    // ========================================================================

    @SuppressWarnings({"unchecked", "rawtypes"})
    private List<?> parseListField(FieldMeta fm) {
        expect('[');
        Class<?> elemClass = fm.elemClass != null ? fm.elemClass : Object.class;
        Type elemType = fm.elemType != null ? fm.elemType : Object.class;
        ClassMeta nestedMeta = fm.listElemMeta; // may be null for non-struct elements

        List<Object> result = new ArrayList<>();
        boolean first = true;
        while (pos < input.length) {
            skipWs();
            if (pos < input.length && input[pos] == ']') { pos++; return result; }
            if (!first) {
                if (pos < input.length && input[pos] == ',') {
                    pos++;
                    skipWs();
                    if (pos < input.length && input[pos] == ']') { pos++; return result; }
                }
            }
            first = false;
            skipWs();

            if (nestedMeta != null && input[pos] == '(') {
                result.add(parseTuple(elemClass, nestedMeta));
            } else {
                result.add(parseFieldValue(elemClass, elemType));
            }
        }
        return result;
    }

    @SuppressWarnings({"unchecked", "rawtypes"})
    private List<?> parseList(Type genericType) {
        expect('[');
        Type elemType = Object.class;
        if (genericType instanceof ParameterizedType pt) {
            elemType = pt.getActualTypeArguments()[0];
        }
        Class<?> elemClass;
        if (elemType instanceof Class<?> c) { elemClass = c; }
        else if (elemType instanceof ParameterizedType pt) { elemClass = (Class<?>) pt.getRawType(); }
        else { elemClass = Object.class; }

        List<Object> result = new ArrayList<>();
        boolean first = true;

        // Hoist struct checks and ClassMeta lookup out of the loop
        boolean isStruct = !isPrimitive(elemClass)
            && !List.class.isAssignableFrom(elemClass)
            && !Map.class.isAssignableFrom(elemClass);
        ClassMeta nestedMeta = isStruct ? ClassMeta.of(elemClass) : null;

        while (pos < input.length) {
            skipWs();
            if (pos < input.length && input[pos] == ']') { pos++; return result; }
            if (!first) {
                if (pos < input.length && input[pos] == ',') {
                    pos++;
                    skipWs();
                    if (pos < input.length && input[pos] == ']') { pos++; return result; }
                }
            }
            first = false;
            skipWs();

            if (nestedMeta != null && input[pos] == '(') {
                result.add(parseTuple(elemClass, nestedMeta));
            } else {
                result.add(parseFieldValue(elemClass, elemType));
            }
        }
        return result;
    }

    @SuppressWarnings("unchecked")
    private Map<?, ?> parseMap(Type genericType) {
        expect('[');
        Type keyType = String.class, valType = Object.class;
        if (genericType instanceof ParameterizedType pt) {
            Type[] args = pt.getActualTypeArguments();
            keyType = args[0]; valType = args[1];
        }
        Class<?> keyClass = (keyType instanceof Class<?> c) ? c : String.class;
        Class<?> valClass = (valType instanceof Class<?> c) ? c : Object.class;

        Map<Object, Object> result = new LinkedHashMap<>();
        boolean first = true;
        while (pos < input.length) {
            skipWs();
            if (pos < input.length && input[pos] == ']') { pos++; return result; }
            if (!first) {
                if (pos < input.length && input[pos] == ',') {
                    pos++;
                    skipWs();
                    if (pos < input.length && input[pos] == ']') { pos++; return result; }
                }
            }
            first = false;
            expect('(');
            skipWs();
            Object key = parseFieldValue(keyClass, keyType);
            skipWs();
            if (pos < input.length && input[pos] == ',') pos++;
            skipWs();
            Object val = parseFieldValue(valClass, valType);
            skipWs();
            if (pos < input.length && input[pos] == ')') pos++;
            result.put(key, val);
        }
        return result;
    }

    // ========================================================================
    // Utility
    // ========================================================================

    private void skipWhitespaceAndComments() {
        while (true) {
            while (pos < input.length) {
                byte b = input[pos];
                if (b > ' ') break; // fast path: most chars > 0x20
                if (b != ' ' && b != '\t' && b != '\n' && b != '\r') break;
                pos++;
            }
            if (pos + 1 < input.length && input[pos] == '/' && input[pos + 1] == '*') {
                pos += 2;
                while (pos + 1 < input.length) {
                    if (input[pos] == '*' && input[pos + 1] == '/') { pos += 2; break; }
                    pos++;
                }
            } else {
                break;
            }
        }
    }

    // Lean whitespace skip — no comment check (for hot paths)
    private void skipWs() {
        while (pos < input.length) {
            byte b = input[pos];
            if (b > ' ') return; // fast: most non-ws chars > 0x20
            if (b != ' ' && b != '\t' && b != '\n' && b != '\r') return;
            pos++;
        }
    }

    private void expect(char c) {
        skipWhitespaceAndComments();
        if (pos >= input.length || input[pos] != (byte) c) {
            throw new AsonException("Expected '" + c + "' at pos " + pos +
                (pos < input.length ? " got '" + (char) input[pos] + "'" : " got EOF"));
        }
        pos++;
    }

    private boolean atValueEnd() {
        if (pos >= input.length) return true;
        byte b = input[pos];
        return b == ',' || b == ')' || b == ']';
    }

    private static Object defaultValue(Class<?> type) {
        if (type == int.class) return 0;
        if (type == long.class) return 0L;
        if (type == float.class) return 0.0f;
        if (type == double.class) return 0.0;
        if (type == boolean.class) return false;
        if (type == short.class) return (short) 0;
        if (type == byte.class) return (byte) 0;
        if (type == char.class) return '\0';
        return null;
    }

    private static boolean isPrimitive(Class<?> c) {
        return c.isPrimitive() || c == String.class || c == Boolean.class ||
            c == Integer.class || c == Long.class || c == Short.class ||
            c == Byte.class || c == Float.class || c == Double.class ||
            c == Character.class;
    }

    private static String unescapePlain(String s) {
        StringBuilder sb = new StringBuilder(s.length());
        for (int i = 0; i < s.length(); i++) {
            char c = s.charAt(i);
            if (c == '\\' && i + 1 < s.length()) {
                char next = s.charAt(++i);
                switch (next) {
                    case ',' -> sb.append(',');
                    case '(' -> sb.append('(');
                    case ')' -> sb.append(')');
                    case '[' -> sb.append('[');
                    case ']' -> sb.append(']');
                    case '"' -> sb.append('"');
                    case '\\' -> sb.append('\\');
                    case 'n' -> sb.append('\n');
                    case 't' -> sb.append('\t');
                    case 'r' -> sb.append('\r');
                    case 'u' -> {
                        if (i + 4 < s.length()) {
                            String hex = s.substring(i + 1, i + 5);
                            sb.append((char) Integer.parseInt(hex, 16));
                            i += 4;
                        }
                    }
                    default -> { sb.append('\\'); sb.append(next); }
                }
            } else {
                sb.append(c);
            }
        }
        return sb.toString();
    }
}
