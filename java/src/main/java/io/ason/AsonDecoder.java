package io.ason;

import java.lang.reflect.*;
import java.nio.charset.StandardCharsets;
import java.util.*;

/**
 * ASON text decoder — reflection-based, zero-copy where possible.
 * Parses both single struct ({schema}:(data)) and struct arrays ([{schema}]:(d1),(d2),...).
 */
final class AsonDecoder {
    private final byte[] input;
    private int pos;

    AsonDecoder(byte[] input) {
        this.input = input;
        this.pos = 0;
    }

    // ========================================================================
    // Public entry points
    // ========================================================================

    <T> T decodeSingle(Class<T> clazz) {
        skipWhitespaceAndComments();
        // Detect format: [{schema}]: or {schema}:
        if (pos < input.length && input[pos] == '[') {
            // [{schema}]:(d1),(d2),... => take first element only
            List<T> list = decodeListInternal(clazz);
            if (list.isEmpty()) throw new AsonException("Empty array, cannot decode single");
            return list.getFirst();
        }

        Field[] fields = Ason.getFields(clazz);
        // Check if starts with {schema}:
        if (pos < input.length && input[pos] == '{') {
            // Parse and skip schema
            skipSchema();
            skipWhitespaceAndComments();
            expect(':');
            skipWhitespaceAndComments();
        } else if (pos < input.length && input[pos] == '(') {
            // Schema-less tuple: positional fields
        }
        return parseTuple(clazz, fields);
    }

    <T> List<T> decodeList(Class<T> clazz) {
        skipWhitespaceAndComments();
        return decodeListInternal(clazz);
    }

    private <T> List<T> decodeListInternal(Class<T> clazz) {
        Field[] fields = Ason.getFields(clazz);
        // [{schema}]:(d1),(d2),...
        if (pos < input.length && input[pos] == '[') {
            pos++; // skip '['
            skipSchema();
            skipWhitespaceAndComments();
            expect(']');
            skipWhitespaceAndComments();
            expect(':');
            skipWhitespaceAndComments();

            List<T> result = new ArrayList<>();
            while (pos < input.length) {
                skipWhitespaceAndComments();
                if (pos >= input.length || input[pos] != '(') break;
                result.add(parseTuple(clazz, fields));
                skipWhitespaceAndComments();
                if (pos < input.length && input[pos] == ',') {
                    pos++;
                }
            }
            return result;
        }
        throw new AsonException("Expected '[' for list format");
    }

    // ========================================================================
    // Schema parsing (skip only — field mapping done by reflection)
    // ========================================================================

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
    // Tuple parsing
    // ========================================================================

    @SuppressWarnings("unchecked")
    private <T> T parseTuple(Class<T> clazz, Field[] fields) {
        expect('(');
        try {
            T obj = clazz.getDeclaredConstructor().newInstance();
            for (int i = 0; i < fields.length; i++) {
                if (i > 0) {
                    skipWhitespaceAndComments();
                    if (pos < input.length && input[pos] == ',') {
                        pos++;
                    } else if (pos < input.length && input[pos] == ')') {
                        break;
                    }
                }
                skipWhitespaceAndComments();
                Object value = parseFieldValue(fields[i].getType(), fields[i].getGenericType());
                fields[i].setAccessible(true);
                fields[i].set(obj, value);
            }
            skipWhitespaceAndComments();
            if (pos < input.length && input[pos] == ')') pos++;
            return obj;
        } catch (ReflectiveOperationException e) {
            throw new AsonException("Failed to create instance of " + clazz.getName(), e);
        }
    }

    // ========================================================================
    // Value parsing
    // ========================================================================

    @SuppressWarnings({"unchecked", "rawtypes"})
    private Object parseFieldValue(Class<?> type, Type genericType) {
        skipWhitespaceAndComments();
        if (pos >= input.length) return null;

        // Optional<T>
        if (type == Optional.class) {
            if (atValueEnd()) {
                return Optional.empty();
            }
            Type innerType = Object.class;
            if (genericType instanceof ParameterizedType pt) {
                innerType = pt.getActualTypeArguments()[0];
            }
            Class<?> innerClass = (innerType instanceof Class<?> c) ? c : Object.class;
            Object inner = parseFieldValue(innerClass, innerType);
            return Optional.ofNullable(inner);
        }

        // Null/empty for value end
        if (atValueEnd()) {
            return defaultValue(type);
        }

        byte b = input[pos];

        // boolean
        if (type == boolean.class || type == Boolean.class) {
            return parseBool();
        }

        // Integer types
        if (type == int.class || type == Integer.class) {
            return (int) parseLong();
        }
        if (type == long.class || type == Long.class) {
            return parseLong();
        }
        if (type == short.class || type == Short.class) {
            return (short) parseLong();
        }
        if (type == byte.class || type == Byte.class) {
            return (byte) parseLong();
        }

        // Float types
        if (type == float.class || type == Float.class) {
            return (float) parseDouble();
        }
        if (type == double.class || type == Double.class) {
            return parseDouble();
        }

        // String
        if (type == String.class) {
            return parseStringValue();
        }

        // char
        if (type == char.class || type == Character.class) {
            String s = parseStringValue();
            return s.isEmpty() ? '\0' : s.charAt(0);
        }

        // List<T>
        if (List.class.isAssignableFrom(type)) {
            return parseList(genericType);
        }

        // Map<K,V>
        if (Map.class.isAssignableFrom(type)) {
            return parseMap(genericType);
        }

        // Nested struct
        if (b == '(') {
            Field[] fields = Ason.getFields(type);
            return parseTuple(type, fields);
        }

        // Fallback: string
        return parseStringValue();
    }

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

    private double parseDouble() {
        int start = pos;
        if (pos < input.length && input[pos] == '-') pos++;
        while (pos < input.length && input[pos] >= '0' && input[pos] <= '9') pos++;
        if (pos < input.length && input[pos] == '.') {
            pos++;
            while (pos < input.length && input[pos] >= '0' && input[pos] <= '9') pos++;
        }
        // Scientific notation
        if (pos < input.length && (input[pos] == 'e' || input[pos] == 'E')) {
            pos++;
            if (pos < input.length && (input[pos] == '+' || input[pos] == '-')) pos++;
            while (pos < input.length && input[pos] >= '0' && input[pos] <= '9') pos++;
        }
        if (pos == start) throw new AsonException("Expected number at pos " + pos);
        return Double.parseDouble(new String(input, start, pos - start, StandardCharsets.US_ASCII));
    }

    private String parseStringValue() {
        skipWhitespaceAndComments();
        if (pos >= input.length || atValueEnd()) return "";
        if (input[pos] == '"') {
            return parseQuotedString();
        }
        return parsePlainString();
    }

    private String parseQuotedString() {
        pos++; // skip '"'
        int start = pos;

        // SIMD fast scan for closing quote or backslash
        int hit = SimdUtils.findQuoteOrBackslash(input, pos, input.length - pos);
        int hitPos = pos + hit;
        if (hitPos < input.length && input[hitPos] == '"') {
            // No escapes — fast path
            String s = new String(input, start, hitPos - start, StandardCharsets.UTF_8);
            pos = hitPos + 1;
            return s;
        }

        // Slow path: has escapes
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
                // After escape, bulk scan for next special
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
        while (pos < input.length) {
            byte b = input[pos];
            if (b == ',' || b == ')' || b == ']') break;
            if (b == '\\') {
                pos += 2;
            } else {
                pos++;
            }
        }
        String raw = new String(input, start, pos - start, StandardCharsets.UTF_8).trim();
        if (raw.contains("\\")) {
            return unescapePlain(raw);
        }
        return raw;
    }

    @SuppressWarnings({"unchecked", "rawtypes"})
    private List<?> parseList(Type genericType) {
        expect('[');
        Type elemType = Object.class;
        if (genericType instanceof ParameterizedType pt) {
            elemType = pt.getActualTypeArguments()[0];
        }
        Class<?> elemClass;
        if (elemType instanceof Class<?> c) {
            elemClass = c;
        } else if (elemType instanceof ParameterizedType pt) {
            elemClass = (Class<?>) pt.getRawType();
        } else {
            elemClass = Object.class;
        }
        List<Object> result = new ArrayList<>();
        boolean first = true;
        while (pos < input.length) {
            skipWhitespaceAndComments();
            if (pos < input.length && input[pos] == ']') {
                pos++;
                return result;
            }
            if (!first) {
                if (pos < input.length && input[pos] == ',') {
                    pos++;
                    skipWhitespaceAndComments();
                    if (pos < input.length && input[pos] == ']') {
                        pos++;
                        return result;
                    }
                }
            }
            first = false;
            skipWhitespaceAndComments();

            // Check if it might be a nested struct list [{schema}]
            if (input[pos] == '(' && !isPrimitive(elemClass)
                && !List.class.isAssignableFrom(elemClass)
                && !Map.class.isAssignableFrom(elemClass)) {
                Field[] fields = Ason.getFields(elemClass);
                result.add(parseTuple(elemClass, fields));
            } else {
                result.add(parseFieldValue(elemClass, elemType));
            }
        }
        return result;
    }

    @SuppressWarnings("unchecked")
    private Map<?, ?> parseMap(Type genericType) {
        expect('[');
        Type keyType = String.class;
        Type valType = Object.class;
        if (genericType instanceof ParameterizedType pt) {
            Type[] args = pt.getActualTypeArguments();
            keyType = args[0];
            valType = args[1];
        }
        Class<?> keyClass = (keyType instanceof Class<?> c) ? c : String.class;
        Class<?> valClass = (valType instanceof Class<?> c) ? c : Object.class;
        Map<Object, Object> result = new LinkedHashMap<>();
        boolean first = true;
        while (pos < input.length) {
            skipWhitespaceAndComments();
            if (pos < input.length && input[pos] == ']') {
                pos++;
                return result;
            }
            if (!first) {
                if (pos < input.length && input[pos] == ',') {
                    pos++;
                    skipWhitespaceAndComments();
                    if (pos < input.length && input[pos] == ']') {
                        pos++;
                        return result;
                    }
                }
            }
            first = false;
            // Expect (key,value)
            expect('(');
            skipWhitespaceAndComments();
            Object key = parseFieldValue(keyClass, keyType);
            skipWhitespaceAndComments();
            if (pos < input.length && input[pos] == ',') pos++;
            skipWhitespaceAndComments();
            Object val = parseFieldValue(valClass, valType);
            skipWhitespaceAndComments();
            if (pos < input.length && input[pos] == ')') pos++;
            result.put(key, val);
        }
        return result;
    }

    // ========================================================================
    // Utility methods
    // ========================================================================

    private void skipWhitespaceAndComments() {
        while (true) {
            // Skip whitespace
            while (pos < input.length) {
                byte b = input[pos];
                if (b != ' ' && b != '\t' && b != '\n' && b != '\r') break;
                pos++;
            }
            // Skip /* ... */ comments
            if (pos + 1 < input.length && input[pos] == '/' && input[pos + 1] == '*') {
                pos += 2;
                while (pos + 1 < input.length) {
                    if (input[pos] == '*' && input[pos + 1] == '/') {
                        pos += 2;
                        break;
                    }
                    pos++;
                }
            } else {
                break;
            }
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
                    default -> {
                        sb.append('\\');
                        sb.append(next);
                    }
                }
            } else {
                sb.append(c);
            }
        }
        return sb.toString();
    }
}
