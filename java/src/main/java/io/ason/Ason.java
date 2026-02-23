package io.ason;

import java.lang.reflect.*;
import java.nio.charset.StandardCharsets;
import java.util.*;

/**
 * ASON (Array-Schema Object Notation) — Java implementation.
 * <p>
 * High-performance, reflection-based serialization with SIMD acceleration.
 * <p>
 * Public API:
 * <ul>
 *   <li>{@link #encode(Object)} — ASON text (untyped schema)</li>
 *   <li>{@link #encodeTyped(Object)} — ASON text (with type annotations)</li>
 *   <li>{@link #decode(String, Class)} — single struct from ASON text</li>
 *   <li>{@link #decodeList(String, Class)} — list of structs from ASON text</li>
 *   <li>{@link #encodeBinary(Object)} — ASON binary</li>
 *   <li>{@link #decodeBinary(byte[], Class)} — single struct from ASON binary</li>
 *   <li>{@link #decodeBinaryList(byte[], Class)} — list of structs from ASON binary</li>
 * </ul>
 */
public final class Ason {

    private Ason() {}

    // ========================================================================
    // Text Encode
    // ========================================================================

    public static String encode(Object value) {
        if (value instanceof List<?> list) {
            return encodeList(list, false);
        }
        return encodeSingle(value, false);
    }

    public static String encodeTyped(Object value) {
        if (value instanceof List<?> list) {
            return encodeList(list, true);
        }
        return encodeSingle(value, true);
    }

    private static String encodeSingle(Object value, boolean typed) {
        ByteBuffer buf = new ByteBuffer(256);
        Class<?> clazz = value.getClass();
        Field[] fields = getFields(clazz);
        buf.append('{');
        writeSchemaFields(buf, fields, typed, value);
        buf.appendStr("}:");
        writeTuple(buf, value, fields);
        return buf.toStringUtf8();
    }

    private static String encodeList(List<?> list, boolean typed) {
        ByteBuffer buf = new ByteBuffer(256);
        if (list.isEmpty()) {
            buf.appendStr("[{}]:");
            return buf.toStringUtf8();
        }
        Object first = list.getFirst();
        Class<?> clazz = first.getClass();
        Field[] fields = getFields(clazz);
        buf.appendStr("[{");
        writeSchemaFields(buf, fields, typed, first);
        buf.appendStr("}]:");
        for (int i = 0; i < list.size(); i++) {
            if (i > 0) buf.append(',');
            writeTuple(buf, list.get(i), fields);
        }
        return buf.toStringUtf8();
    }

    private static void writeSchemaFields(ByteBuffer buf, Field[] fields, boolean typed, Object sampleValue) {
        for (int i = 0; i < fields.length; i++) {
            if (i > 0) buf.append(',');
            buf.appendStr(fields[i].getName());
            if (typed) {
                String hint = typeHint(fields[i], sampleValue);
                if (hint != null) {
                    buf.append(':');
                    buf.appendStr(hint);
                }
            }
        }
    }

    private static String typeHint(Field f, Object sample) {
        Class<?> type = f.getType();
        if (type == boolean.class || type == Boolean.class) return "bool";
        if (type == int.class || type == Integer.class || type == long.class || type == Long.class
            || type == short.class || type == Short.class || type == byte.class || type == Byte.class) return "int";
        if (type == float.class || type == Float.class || type == double.class || type == Double.class) return "float";
        if (type == String.class || type == char.class || type == Character.class) return "str";
        if (type == Optional.class) {
            // Check if it has a value for type hint
            try {
                f.setAccessible(true);
                Object val = f.get(sample);
                if (val instanceof Optional<?> opt && opt.isPresent()) {
                    Object inner = opt.get();
                    return typeHintForValue(inner);
                }
            } catch (Exception e) { /* ignore */ }
            return null;
        }
        if (List.class.isAssignableFrom(type)) {
            Type genType = f.getGenericType();
            if (genType instanceof ParameterizedType pt) {
                Type arg = pt.getActualTypeArguments()[0];
                String inner = typeHintForType(arg);
                if (inner != null) return "[" + inner + "]";
            }
            return null;
        }
        // Nested struct — no simple type hint
        return null;
    }

    private static String typeHintForValue(Object val) {
        if (val instanceof Boolean) return "bool";
        if (val instanceof Integer || val instanceof Long || val instanceof Short || val instanceof Byte) return "int";
        if (val instanceof Float || val instanceof Double) return "float";
        if (val instanceof String || val instanceof Character) return "str";
        return null;
    }

    private static String typeHintForType(Type type) {
        if (type instanceof Class<?> c) {
            if (c == String.class) return "str";
            if (c == Integer.class || c == int.class || c == Long.class || c == long.class
                || c == Short.class || c == short.class || c == Byte.class || c == byte.class) return "int";
            if (c == Float.class || c == float.class || c == Double.class || c == double.class) return "float";
            if (c == Boolean.class || c == boolean.class) return "bool";
        }
        if (type instanceof ParameterizedType pt) {
            Type raw = pt.getRawType();
            if (raw == List.class) {
                String inner = typeHintForType(pt.getActualTypeArguments()[0]);
                if (inner != null) return "[" + inner + "]";
            }
        }
        return null;
    }

    private static void writeTuple(ByteBuffer buf, Object value, Field[] fields) {
        buf.append('(');
        for (int i = 0; i < fields.length; i++) {
            if (i > 0) buf.append(',');
            try {
                fields[i].setAccessible(true);
                Object fv = fields[i].get(value);
                writeFieldValue(buf, fv, fields[i].getType(), fields[i].getGenericType());
            } catch (Exception e) {
                throw new AsonException("Failed to read field: " + fields[i].getName(), e);
            }
        }
        buf.append(')');
    }

    @SuppressWarnings("unchecked")
    private static void writeFieldValue(ByteBuffer buf, Object value, Class<?> type, Type genericType) {
        if (value == null) {
            // empty = null for Optional
            return;
        }
        if (value instanceof Optional<?> opt) {
            if (opt.isPresent()) {
                Object inner = opt.get();
                writeFieldValue(buf, inner, inner.getClass(), inner.getClass());
            }
            return;
        }
        if (type == boolean.class || type == Boolean.class) {
            buf.appendStr((Boolean) value ? "true" : "false");
        } else if (type == int.class || type == Integer.class) {
            writeInt(buf, (Integer) value);
        } else if (type == long.class || type == Long.class) {
            writeLong(buf, (Long) value);
        } else if (type == short.class || type == Short.class) {
            writeInt(buf, (Short) value);
        } else if (type == byte.class || type == Byte.class) {
            writeInt(buf, (Byte) value);
        } else if (type == float.class || type == Float.class) {
            writeFloat(buf, (Float) value);
        } else if (type == double.class || type == Double.class) {
            writeDouble(buf, (Double) value);
        } else if (type == char.class || type == Character.class) {
            writeString(buf, String.valueOf(value));
        } else if (type == String.class) {
            writeString(buf, (String) value);
        } else if (List.class.isAssignableFrom(type)) {
            List<?> list = (List<?>) value;
            Type elemType = Object.class;
            if (genericType instanceof ParameterizedType pt) {
                elemType = pt.getActualTypeArguments()[0];
            }
            buf.append('[');
            Class<?> elemClass;
            if (elemType instanceof Class<?> c) {
                elemClass = c;
            } else if (elemType instanceof ParameterizedType pt2) {
                elemClass = (Class<?>) pt2.getRawType();
            } else {
                elemClass = Object.class;
            }
            for (int i = 0; i < list.size(); i++) {
                if (i > 0) buf.append(',');
                Object item = list.get(i);
                if (item != null) {
                    writeFieldValue(buf, item, elemClass, elemType);
                }
            }
            buf.append(']');
        } else if (Map.class.isAssignableFrom(type)) {
            Map<?, ?> map = (Map<?, ?>) value;
            Type keyType = String.class, valType = Object.class;
            if (genericType instanceof ParameterizedType pt) {
                Type[] args = pt.getActualTypeArguments();
                keyType = args[0];
                valType = args[1];
            }
            Class<?> keyClass = (keyType instanceof Class<?> c) ? c : String.class;
            Class<?> valClass = (valType instanceof Class<?> c) ? c : Object.class;
            buf.append('[');
            boolean first = true;
            for (var entry : map.entrySet()) {
                if (!first) buf.append(',');
                first = false;
                buf.append('(');
                writeFieldValue(buf, entry.getKey(), keyClass, keyType);
                buf.append(',');
                writeFieldValue(buf, entry.getValue(), valClass, valType);
                buf.append(')');
            }
            buf.append(']');
        } else {
            // Nested struct
            Field[] fields = getFields(type);
            writeTuple(buf, value, fields);
        }
    }

    // ========================================================================
    // Fast integer/float formatting
    // ========================================================================

    private static final byte[] DEC_DIGITS = "00010203040506070809101112131415161718192021222324252627282930313233343536373839404142434445464748495051525354555657585960616263646566676869707172737475767778798081828384858687888990919293949596979899".getBytes(StandardCharsets.US_ASCII);

    static void writeInt(ByteBuffer buf, int v) {
        writeLong(buf, v);
    }

    static void writeLong(ByteBuffer buf, long v) {
        if (v < 0) {
            if (v == Long.MIN_VALUE) {
                buf.appendStr("-9223372036854775808");
                return;
            }
            buf.append('-');
            v = -v;
        }
        writeULong(buf, v);
    }

    static void writeULong(ByteBuffer buf, long v) {
        if (v < 10) {
            buf.append((byte) ('0' + v));
            return;
        }
        if (v < 100) {
            int idx = (int) (v * 2);
            buf.append(DEC_DIGITS[idx]);
            buf.append(DEC_DIGITS[idx + 1]);
            return;
        }
        // Stack-based: write digits right-to-left
        byte[] tmp = new byte[20];
        int pos = 20;
        while (v >= 100) {
            int rem = (int) (v % 100);
            v /= 100;
            pos -= 2;
            tmp[pos] = DEC_DIGITS[rem * 2];
            tmp[pos + 1] = DEC_DIGITS[rem * 2 + 1];
        }
        if (v >= 10) {
            int idx = (int) (v * 2);
            pos -= 2;
            tmp[pos] = DEC_DIGITS[idx];
            tmp[pos + 1] = DEC_DIGITS[idx + 1];
        } else {
            pos--;
            tmp[pos] = (byte) ('0' + v);
        }
        buf.appendBytes(tmp, pos, 20 - pos);
    }

    static void writeDouble(ByteBuffer buf, double v) {
        if (Double.isFinite(v) && v == Math.floor(v) && Math.abs(v) < (double) Long.MAX_VALUE) {
            writeLong(buf, (long) v);
            buf.appendStr(".0");
            return;
        }
        if (Double.isFinite(v)) {
            double v10 = v * 10.0;
            if (v10 == Math.floor(v10) && Math.abs(v10) < 1e18) {
                long vi = (long) v10;
                if (vi < 0) {
                    buf.append('-');
                    vi = -vi;
                }
                writeULong(buf, vi / 10);
                buf.append('.');
                buf.append((byte) ('0' + (vi % 10)));
                return;
            }
        }
        buf.appendStr(Double.toString(v));
    }

    static void writeFloat(ByteBuffer buf, float v) {
        writeDouble(buf, v);
    }

    // ========================================================================
    // String quoting
    // ========================================================================

    static void writeString(ByteBuffer buf, String s) {
        if (needsQuoting(s)) {
            writeEscaped(buf, s);
        } else {
            buf.appendStr(s);
        }
    }

    static boolean needsQuoting(String s) {
        if (s.isEmpty()) return true;
        byte[] bytes = s.getBytes(StandardCharsets.UTF_8);
        if (bytes[0] == ' ' || bytes[bytes.length - 1] == ' ') return true;
        if (s.equals("true") || s.equals("false")) return true;
        if (SimdUtils.hasSpecialChars(bytes, 0, bytes.length)) return true;
        // Check if it looks like a number
        int start = bytes[0] == '-' ? 1 : 0;
        if (start < bytes.length) {
            boolean couldBeNumber = true;
            for (int i = start; i < bytes.length; i++) {
                byte b = bytes[i];
                if ((b < '0' || b > '9') && b != '.') {
                    couldBeNumber = false;
                    break;
                }
            }
            if (couldBeNumber) return true;
        }
        return false;
    }

    static void writeEscaped(ByteBuffer buf, String s) {
        byte[] bytes = s.getBytes(StandardCharsets.UTF_8);
        buf.append('"');
        int start = 0;
        while (start < bytes.length) {
            int next = SimdUtils.findEscape(bytes, start, bytes.length - start);
            if (next > 0) {
                buf.appendBytes(bytes, start, next);
            }
            int idx = start + next;
            if (idx >= bytes.length) break;
            byte b = bytes[idx];
            switch (b) {
                case '"' -> buf.appendStr("\\\"");
                case '\\' -> buf.appendStr("\\\\");
                case '\n' -> buf.appendStr("\\n");
                case '\t' -> buf.appendStr("\\t");
                case '\r' -> buf.appendStr("\\r");
                default -> {
                    buf.appendStr("\\u00");
                    buf.append((byte) "0123456789abcdef".charAt((b >> 4) & 0xf));
                    buf.append((byte) "0123456789abcdef".charAt(b & 0xf));
                }
            }
            start = idx + 1;
        }
        buf.append('"');
    }

    // ========================================================================
    // Text Decode
    // ========================================================================

    public static <T> T decode(String input, Class<T> clazz) {
        return new AsonDecoder(input.getBytes(StandardCharsets.UTF_8)).decodeSingle(clazz);
    }

    public static <T> List<T> decodeList(String input, Class<T> clazz) {
        return new AsonDecoder(input.getBytes(StandardCharsets.UTF_8)).decodeList(clazz);
    }

    // ========================================================================
    // Binary Encode/Decode
    // ========================================================================

    public static byte[] encodeBinary(Object value) {
        return AsonBinary.encode(value);
    }

    public static <T> T decodeBinary(byte[] data, Class<T> clazz) {
        return AsonBinary.decode(data, clazz);
    }

    public static <T> List<T> decodeBinaryList(byte[] data, Class<T> clazz) {
        return AsonBinary.decodeList(data, clazz);
    }

    // ========================================================================
    // Reflection cache
    // ========================================================================

    private static final Map<Class<?>, Field[]> FIELD_CACHE = new WeakHashMap<>();

    static Field[] getFields(Class<?> clazz) {
        return FIELD_CACHE.computeIfAbsent(clazz, c -> {
            Field[] all = c.getDeclaredFields();
            List<Field> result = new ArrayList<>();
            for (Field f : all) {
                if (Modifier.isStatic(f.getModifiers()) || Modifier.isTransient(f.getModifiers())) continue;
                if (f.isSynthetic()) continue;
                f.setAccessible(true);
                result.add(f);
            }
            return result.toArray(new Field[0]);
        });
    }
}
