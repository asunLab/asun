package io.ason;

import io.ason.ClassMeta.FieldMeta;
import java.lang.reflect.*;
import java.nio.charset.StandardCharsets;
import java.util.*;

/**
 * ASON (Array-Schema Object Notation) — Java implementation.
 * High-performance encoder with ClassMeta pre-computed metadata,
 * MethodHandle invokeExact, and single-pass fused string writing.
 */
public final class Ason {

    private Ason() {}

    private static final byte[] TRUE_BYTES  = {'t','r','u','e'};
    private static final byte[] FALSE_BYTES = {'f','a','l','s','e'};
    private static final byte[] EMPTY_SCHEMA = {'[','{','}',']',':'};
    private static final byte[] HEX = "0123456789abcdef".getBytes(StandardCharsets.US_ASCII);

    // Character classification table for writeString fast path (0-127)
    // 0 = normal (pass-through), 1 = needs quoting (special ASON delimiters)
    private static final byte[] CHAR_CLASS = new byte[128];
    static {
        // Control chars
        for (int i = 0; i < 0x20; i++) CHAR_CLASS[i] = 1;
        // ASON delimiters
        CHAR_CLASS[','] = 1; CHAR_CLASS['('] = 1; CHAR_CLASS[')'] = 1;
        CHAR_CLASS['['] = 1; CHAR_CLASS[']'] = 1;
        CHAR_CLASS['"'] = 1; CHAR_CLASS['\\'] = 1;
    }

    // ========================================================================
    // Text Encode
    // ========================================================================

    public static String encode(Object value) {
        if (value instanceof List<?> list) return encodeList(list, false);
        return encodeSingle(value, false);
    }

    public static String encodeTyped(Object value) {
        if (value instanceof List<?> list) return encodeList(list, true);
        return encodeSingle(value, true);
    }

    private static String encodeSingle(Object value, boolean typed) {
        ByteBuffer buf = new ByteBuffer();
        ClassMeta meta = ClassMeta.of(value.getClass());
        if (typed) {
            buf.append('{');
            writeSchemaFieldsTyped(buf, meta.fields, value);
            buf.append('}');
        } else {
            buf.append('{');
            writeSchemaFields(buf, meta.fields);
            buf.append('}');
        }
        buf.append(':');
        writeTuple(buf, value, meta);
        return buf.toStringUtf8AndClose();
    }

    private static String encodeList(List<?> list, boolean typed) {
        ByteBuffer buf = new ByteBuffer();
        if (list.isEmpty()) {
            buf.appendBytes(EMPTY_SCHEMA, 0, 5);
            return buf.toStringUtf8AndClose();
        }
        Object first = list.getFirst();
        ClassMeta meta = ClassMeta.of(first.getClass());
        if (typed) {
            buf.append('['); buf.append('{');
            writeSchemaFieldsTyped(buf, meta.fields, first);
            buf.append('}'); buf.append(']');
        } else {
            buf.append('['); buf.append('{');
            writeSchemaFields(buf, meta.fields);
            buf.append('}'); buf.append(']');
        }
        buf.append(':');
        for (int i = 0; i < list.size(); i++) {
            if (i > 0) buf.append(',');
            writeTuple(buf, list.get(i), meta);
        }
        return buf.toStringUtf8AndClose();
    }

    /**
     * Write recursive schema fields (untyped): name, or name:{nested}, or name:[{nested}]
     */
    private static void writeSchemaFields(ByteBuffer buf, FieldMeta[] fields) {
        for (int i = 0; i < fields.length; i++) {
            if (i > 0) buf.append(',');
            FieldMeta fm = fields[i];
            buf.appendBytes(fm.nameBytes, 0, fm.nameBytes.length);
            if (fm.typeTag == FieldMeta.T_STRUCT && fm.nestedMeta != null) {
                buf.append(':');
                buf.append('{');
                writeSchemaFields(buf, fm.nestedMeta.fields);
                buf.append('}');
            } else if (fm.typeTag == FieldMeta.T_LIST && fm.listElemMeta != null) {
                buf.append(':');
                buf.append('['); buf.append('{');
                writeSchemaFields(buf, fm.listElemMeta.fields);
                buf.append('}'); buf.append(']');
            }
        }
    }

    /**
     * Write recursive schema fields (typed): name:type, or name:{nested}, or name:[{nested}]
     */
    @SuppressWarnings("unchecked")
    private static void writeSchemaFieldsTyped(ByteBuffer buf, FieldMeta[] fields, Object sample) {
        for (int i = 0; i < fields.length; i++) {
            if (i > 0) buf.append(',');
            FieldMeta fm = fields[i];
            buf.appendBytes(fm.nameBytes, 0, fm.nameBytes.length);
            if (fm.typeTag == FieldMeta.T_STRUCT && fm.nestedMeta != null) {
                buf.append(':');
                buf.append('{');
                Object nestedObj = sample != null ? fm.get(sample) : null;
                if (nestedObj != null) {
                    writeSchemaFieldsTyped(buf, fm.nestedMeta.fields, nestedObj);
                } else {
                    writeSchemaFields(buf, fm.nestedMeta.fields);
                }
                buf.append('}');
            } else if (fm.typeTag == FieldMeta.T_LIST && fm.listElemMeta != null) {
                buf.append(':');
                buf.append('['); buf.append('{');
                List<?> listSample = sample != null ? (List<?>) fm.get(sample) : null;
                if (listSample != null && !listSample.isEmpty()) {
                    writeSchemaFieldsTyped(buf, fm.listElemMeta.fields, listSample.getFirst());
                } else {
                    writeSchemaFields(buf, fm.listElemMeta.fields);
                }
                buf.append('}'); buf.append(']');
            } else {
                String hint = typeHintField(fm, sample);
                if (hint != null) { buf.append(':'); buf.appendAscii(hint); }
            }
        }
    }

    private static String typeHintField(FieldMeta fm, Object sample) {
        return switch (fm.typeTag) {
            case FieldMeta.T_BOOLEAN -> "bool";
            case FieldMeta.T_INT, FieldMeta.T_LONG, FieldMeta.T_SHORT, FieldMeta.T_BYTE -> "int";
            case FieldMeta.T_FLOAT, FieldMeta.T_DOUBLE -> "float";
            case FieldMeta.T_STRING, FieldMeta.T_CHAR -> "str";
            case FieldMeta.T_OPTIONAL -> {
                try {
                    Object val = fm.get(sample);
                    if (val instanceof Optional<?> opt && opt.isPresent()) yield typeHintForValue(opt.get());
                } catch (Exception e) { /* ignore */ }
                yield null;
            }
            case FieldMeta.T_LIST -> {
                if (fm.elemType != null) {
                    String inner = typeHintForType(fm.elemType);
                    if (inner != null) yield "[" + inner + "]";
                }
                yield null;
            }
            default -> null;
        };
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

    // ========================================================================
    // Tuple writing — type-tag dispatched, no boxing for primitives
    // ========================================================================

    @SuppressWarnings("unchecked")
    private static void writeTuple(ByteBuffer buf, Object obj, ClassMeta meta) {
        buf.append('(');
        FieldMeta[] fields = meta.fields;
        for (int i = 0; i < fields.length; i++) {
            if (i > 0) buf.append(',');
            FieldMeta fm = fields[i];
            switch (fm.typeTag) {
                case FieldMeta.T_BOOLEAN -> {
                    if (fm.isPrimitive) {
                        boolean v = fm.getBoolean(obj);
                        buf.appendBytes(v ? TRUE_BYTES : FALSE_BYTES, 0, v ? 4 : 5);
                    } else {
                        Object v = fm.get(obj);
                        if (v != null) { boolean bv = (Boolean) v; buf.appendBytes(bv ? TRUE_BYTES : FALSE_BYTES, 0, bv ? 4 : 5); }
                    }
                }
                case FieldMeta.T_INT -> {
                    if (fm.isPrimitive) writeInt(buf, fm.getInt(obj));
                    else { Object v = fm.get(obj); if (v != null) writeInt(buf, (Integer) v); }
                }
                case FieldMeta.T_LONG -> {
                    if (fm.isPrimitive) writeLong(buf, fm.getLong(obj));
                    else { Object v = fm.get(obj); if (v != null) writeLong(buf, (Long) v); }
                }
                case FieldMeta.T_SHORT -> {
                    if (fm.isPrimitive) writeInt(buf, fm.getShort(obj));
                    else { Object v = fm.get(obj); if (v != null) writeInt(buf, (Short) v); }
                }
                case FieldMeta.T_BYTE -> {
                    if (fm.isPrimitive) writeInt(buf, fm.getByte(obj));
                    else { Object v = fm.get(obj); if (v != null) writeInt(buf, (Byte) v); }
                }
                case FieldMeta.T_FLOAT -> {
                    if (fm.isPrimitive) writeFloat(buf, fm.getFloat(obj));
                    else { Object v = fm.get(obj); if (v != null) writeFloat(buf, (Float) v); }
                }
                case FieldMeta.T_DOUBLE -> {
                    if (fm.isPrimitive) writeDouble(buf, fm.getDouble(obj));
                    else { Object v = fm.get(obj); if (v != null) writeDouble(buf, (Double) v); }
                }
                case FieldMeta.T_CHAR -> {
                    if (fm.isPrimitive) writeString(buf, String.valueOf(fm.getChar(obj)));
                    else { Object v = fm.get(obj); if (v != null) writeString(buf, String.valueOf(v)); }
                }
                case FieldMeta.T_STRING -> {
                    String sv = (String) fm.get(obj);
                    if (sv != null) writeString(buf, sv);
                }
                case FieldMeta.T_OPTIONAL -> {
                    Optional<?> opt = (Optional<?>) fm.get(obj);
                    if (opt != null && opt.isPresent()) {
                        Object inner = opt.get();
                        writeValue(buf, inner, inner.getClass(), inner.getClass());
                    }
                }
                case FieldMeta.T_LIST -> {
                    List<?> list = (List<?>) fm.get(obj);
                    if (list != null) writeListValue(buf, list,
                        fm.elemClass != null ? fm.elemClass : Object.class,
                        fm.elemType != null ? fm.elemType : Object.class);
                }
                case FieldMeta.T_MAP -> {
                    Map<?, ?> map = (Map<?, ?>) fm.get(obj);
                    if (map != null) writeMapValue(buf, map,
                        fm.elemClass != null ? fm.elemClass : String.class,
                        fm.elemType != null ? fm.elemType : String.class,
                        fm.valClass != null ? fm.valClass : Object.class,
                        fm.valType != null ? fm.valType : Object.class);
                }
                default -> {
                    Object nested = fm.get(obj);
                    if (nested != null) writeTuple(buf, nested, ClassMeta.of(fm.type));
                }
            }
        }
        buf.append(')');
    }

    @SuppressWarnings("unchecked")
    private static void writeValue(ByteBuffer buf, Object value, Class<?> type, Type genericType) {
        if (value == null) return;
        if (type == Object.class) type = value.getClass();
        if (value instanceof Optional<?> opt) {
            if (opt.isPresent()) { Object inner = opt.get(); writeValue(buf, inner, inner.getClass(), inner.getClass()); }
            return;
        }
        if (type == boolean.class || type == Boolean.class) {
            boolean v = (Boolean) value; buf.appendBytes(v ? TRUE_BYTES : FALSE_BYTES, 0, v ? 4 : 5);
        } else if (type == int.class || type == Integer.class) { writeInt(buf, (Integer) value);
        } else if (type == long.class || type == Long.class) { writeLong(buf, (Long) value);
        } else if (type == short.class || type == Short.class) { writeInt(buf, (Short) value);
        } else if (type == byte.class || type == Byte.class) { writeInt(buf, (Byte) value);
        } else if (type == float.class || type == Float.class) { writeFloat(buf, (Float) value);
        } else if (type == double.class || type == Double.class) { writeDouble(buf, (Double) value);
        } else if (type == char.class || type == Character.class) { writeString(buf, String.valueOf(value));
        } else if (type == String.class) { writeString(buf, (String) value);
        } else if (List.class.isAssignableFrom(type)) {
            List<?> list = (List<?>) value;
            Type elemType = Object.class; Class<?> elemClass = Object.class;
            if (genericType instanceof ParameterizedType pt) { elemType = pt.getActualTypeArguments()[0]; elemClass = FieldMeta.resolveClass(elemType); }
            writeListValue(buf, list, elemClass, elemType);
        } else if (Map.class.isAssignableFrom(type)) {
            Map<?, ?> map = (Map<?, ?>) value;
            Type keyType = String.class, valType = Object.class; Class<?> keyClass = String.class, valClass = Object.class;
            if (genericType instanceof ParameterizedType pt) {
                Type[] args = pt.getActualTypeArguments();
                keyType = args[0]; keyClass = FieldMeta.resolveClass(keyType);
                valType = args[1]; valClass = FieldMeta.resolveClass(valType);
            }
            writeMapValue(buf, map, keyClass, keyType, valClass, valType);
        } else { writeTuple(buf, value, ClassMeta.of(type)); }
    }

    private static void writeListValue(ByteBuffer buf, List<?> list, Class<?> elemClass, Type elemType) {
        buf.append('[');
        for (int i = 0; i < list.size(); i++) {
            if (i > 0) buf.append(',');
            Object item = list.get(i);
            if (item != null) writeValue(buf, item, elemClass, elemType);
        }
        buf.append(']');
    }

    private static void writeMapValue(ByteBuffer buf, Map<?, ?> map,
            Class<?> keyClass, Type keyType, Class<?> valClass, Type valType) {
        buf.append('[');
        boolean first = true;
        for (var entry : map.entrySet()) {
            if (!first) buf.append(',');
            first = false;
            buf.append('(');
            writeValue(buf, entry.getKey(), keyClass, keyType);
            buf.append(',');
            writeValue(buf, entry.getValue(), valClass, valType);
            buf.append(')');
        }
        buf.append(']');
    }

    // ========================================================================
    // Fast integer/float formatting
    // ========================================================================

    private static final byte[] DEC_DIGITS = "00010203040506070809101112131415161718192021222324252627282930313233343536373839404142434445464748495051525354555657585960616263646566676869707172737475767778798081828384858687888990919293949596979899".getBytes(StandardCharsets.US_ASCII);

    static void writeInt(ByteBuffer buf, int v) { writeLong(buf, v); }

    static void writeLong(ByteBuffer buf, long v) {
        if (v < 0) {
            if (v == Long.MIN_VALUE) { buf.appendAscii("-9223372036854775808"); return; }
            buf.append('-');
            v = -v;
        }
        writeULong(buf, v);
    }

    static void writeULong(ByteBuffer buf, long v) {
        if (v < 10) { buf.append((byte) ('0' + v)); return; }
        if (v < 100) {
            int idx = (int) (v * 2);
            buf.ensureCapacity(2);
            buf.data[buf.len++] = DEC_DIGITS[idx];
            buf.data[buf.len++] = DEC_DIGITS[idx + 1];
            return;
        }
        int digits = digitCount(v);
        buf.ensureCapacity(digits);
        int pos = buf.len + digits;
        buf.len = pos;
        while (v >= 100) {
            int rem = (int) (v % 100);
            v /= 100;
            pos -= 2;
            buf.data[pos] = DEC_DIGITS[rem * 2];
            buf.data[pos + 1] = DEC_DIGITS[rem * 2 + 1];
        }
        if (v >= 10) {
            int idx = (int) (v * 2);
            pos -= 2;
            buf.data[pos] = DEC_DIGITS[idx];
            buf.data[pos + 1] = DEC_DIGITS[idx + 1];
        } else {
            buf.data[pos - 1] = (byte) ('0' + v);
        }
    }

    private static int digitCount(long v) {
        if (v < 10L) return 1;       if (v < 100L) return 2;
        if (v < 1000L) return 3;     if (v < 10000L) return 4;
        if (v < 100000L) return 5;   if (v < 1000000L) return 6;
        if (v < 10000000L) return 7; if (v < 100000000L) return 8;
        if (v < 1000000000L) return 9; if (v < 10000000000L) return 10;
        if (v < 100000000000L) return 11; if (v < 1000000000000L) return 12;
        if (v < 10000000000000L) return 13; if (v < 100000000000000L) return 14;
        if (v < 1000000000000000L) return 15; if (v < 10000000000000000L) return 16;
        if (v < 100000000000000000L) return 17; if (v < 1000000000000000000L) return 18;
        return 19;
    }

    static void writeDouble(ByteBuffer buf, double v) {
        if (Double.isFinite(v) && v == Math.floor(v) && Math.abs(v) < (double) Long.MAX_VALUE) {
            writeLong(buf, (long) v);
            buf.ensureCapacity(2); buf.data[buf.len++] = '.'; buf.data[buf.len++] = '0';
            return;
        }
        if (Double.isFinite(v)) {
            double v10 = v * 10.0;
            if (v10 == Math.floor(v10) && Math.abs(v10) < 1e18) {
                long vi = (long) v10;
                if (vi < 0) { buf.append('-'); vi = -vi; }
                writeULong(buf, vi / 10);
                buf.ensureCapacity(2); buf.data[buf.len++] = '.'; buf.data[buf.len++] = (byte) ('0' + (vi % 10));
                return;
            }
        }
        buf.appendAscii(Double.toString(v));
    }

    static void writeFloat(ByteBuffer buf, float v) { writeDouble(buf, v); }

    // ========================================================================
    // String writing — single-pass fused scan+write, no double getBytes
    // ========================================================================

    static void writeString(ByteBuffer buf, String s) {
        int slen = s.length();
        if (slen == 0) {
            buf.ensureCapacity(2); buf.data[buf.len++] = '"'; buf.data[buf.len++] = '"';
            return;
        }

        char first = s.charAt(0);
        char last = s.charAt(slen - 1);

        // Leading/trailing space
        if (first == ' ' || last == ' ') { writeQuoted(buf, s, slen); return; }

        // "true" / "false"
        if (slen == 4 && first == 't' && s.charAt(1) == 'r' && s.charAt(2) == 'u' && s.charAt(3) == 'e') {
            writeQuoted(buf, s, slen); return;
        }
        if (slen == 5 && first == 'f' && s.charAt(1) == 'a' && s.charAt(2) == 'l' && s.charAt(3) == 's' && s.charAt(4) == 'e') {
            writeQuoted(buf, s, slen); return;
        }

        // Combined scan + write in a single pass
        buf.ensureCapacity(slen);
        int startLen = buf.len;
        boolean couldBeNumber = true;
        int numStart = first == '-' ? 1 : 0;

        for (int i = 0; i < slen; i++) {
            char c = s.charAt(i);
            if (c >= 128) {
                // Non-ASCII: rollback, handle with getBytes
                buf.len = startLen;
                writeStringNonAscii(buf, s, slen);
                return;
            }
            if (CHAR_CLASS[c] != 0) {
                // Special char: rollback, write quoted
                buf.len = startLen;
                writeQuoted(buf, s, slen);
                return;
            }
            if (couldBeNumber && i >= numStart) {
                if ((c < '0' || c > '9') && c != '.') couldBeNumber = false;
            }
            buf.data[buf.len++] = (byte) c;
        }

        // Number check: rollback + quote
        if (couldBeNumber && numStart < slen) {
            // Already wrote slen bytes; insert quotes around them
            buf.ensureCapacity(2);
            System.arraycopy(buf.data, startLen, buf.data, startLen + 1, slen);
            buf.data[startLen] = '"';
            buf.len = startLen + 1 + slen;
            buf.data[buf.len++] = '"';
        }
        // else: success, bytes already written
    }

    private static void writeStringNonAscii(ByteBuffer buf, String s, int slen) {
        buf.hasNonAscii = true; // reached here because scan found non-ASCII
        boolean needsQuote = false;
        boolean couldBeNumber = true;
        int numStart = s.charAt(0) == '-' ? 1 : 0;
        for (int i = 0; i < slen; i++) {
            char c = s.charAt(i);
            if (c < 128 && CHAR_CLASS[c] != 0) {
                needsQuote = true; break;
            }
            if (couldBeNumber && i >= numStart && (c < '0' || c > '9') && c != '.') couldBeNumber = false;
        }
        if (!needsQuote && couldBeNumber && numStart < slen) needsQuote = true;
        if (needsQuote) { writeQuoted(buf, s, slen); }
        else { buf.appendStr(s); }
    }

    private static void writeQuoted(ByteBuffer buf, String s, int slen) {
        buf.append('"');
        boolean needsEscape = false;
        boolean hasNonAscii = false;
        for (int i = 0; i < slen; i++) {
            char c = s.charAt(i);
            if (c == '"' || c == '\\' || c < 0x20) { needsEscape = true; break; }
            if (c >= 128) hasNonAscii = true;
        }
        if (!needsEscape) {
            if (!hasNonAscii) buf.appendCharsAsBytes(s, slen);
            else buf.appendStr(s);
        } else if (!hasNonAscii) {
            writeEscapedAscii(buf, s, slen);
        } else {
            writeEscapedBytes(buf, s);
        }
        buf.append('"');
    }

    private static void writeEscapedAscii(ByteBuffer buf, String s, int slen) {
        buf.ensureCapacity(slen);
        for (int i = 0; i < slen; i++) {
            char c = s.charAt(i);
            switch (c) {
                case '"' -> { buf.ensureCapacity(2); buf.data[buf.len++] = '\\'; buf.data[buf.len++] = '"'; }
                case '\\' -> { buf.ensureCapacity(2); buf.data[buf.len++] = '\\'; buf.data[buf.len++] = '\\'; }
                case '\n' -> { buf.ensureCapacity(2); buf.data[buf.len++] = '\\'; buf.data[buf.len++] = 'n'; }
                case '\t' -> { buf.ensureCapacity(2); buf.data[buf.len++] = '\\'; buf.data[buf.len++] = 't'; }
                case '\r' -> { buf.ensureCapacity(2); buf.data[buf.len++] = '\\'; buf.data[buf.len++] = 'r'; }
                default -> {
                    if (c < 0x20) {
                        buf.ensureCapacity(6);
                        buf.data[buf.len++] = '\\'; buf.data[buf.len++] = 'u';
                        buf.data[buf.len++] = '0'; buf.data[buf.len++] = '0';
                        buf.data[buf.len++] = HEX[(c >> 4) & 0xf]; buf.data[buf.len++] = HEX[c & 0xf];
                    } else {
                        if (buf.len >= buf.data.length) buf.ensureCapacity(1);
                        buf.data[buf.len++] = (byte) c;
                    }
                }
            }
        }
    }

    private static void writeEscapedBytes(ByteBuffer buf, String s) {
        buf.hasNonAscii = true; // we know the string has non-ASCII chars
        byte[] bytes = s.getBytes(StandardCharsets.UTF_8);
        int start = 0;
        while (start < bytes.length) {
            int next = SimdUtils.findEscape(bytes, start, bytes.length - start);
            if (next > 0) buf.appendBytes(bytes, start, next);
            int idx = start + next;
            if (idx >= bytes.length) break;
            byte b = bytes[idx];
            switch (b) {
                case '"' -> { buf.ensureCapacity(2); buf.data[buf.len++] = '\\'; buf.data[buf.len++] = '"'; }
                case '\\' -> { buf.ensureCapacity(2); buf.data[buf.len++] = '\\'; buf.data[buf.len++] = '\\'; }
                case '\n' -> { buf.ensureCapacity(2); buf.data[buf.len++] = '\\'; buf.data[buf.len++] = 'n'; }
                case '\t' -> { buf.ensureCapacity(2); buf.data[buf.len++] = '\\'; buf.data[buf.len++] = 't'; }
                case '\r' -> { buf.ensureCapacity(2); buf.data[buf.len++] = '\\'; buf.data[buf.len++] = 'r'; }
                default -> {
                    buf.ensureCapacity(6);
                    buf.data[buf.len++] = '\\'; buf.data[buf.len++] = 'u';
                    buf.data[buf.len++] = '0'; buf.data[buf.len++] = '0';
                    buf.data[buf.len++] = HEX[(b >> 4) & 0xf]; buf.data[buf.len++] = HEX[b & 0xf];
                }
            }
            start = idx + 1;
        }
    }

    // ========================================================================
    // Text Decode
    // ========================================================================

    public static <T> T decode(String input, Class<T> clazz) {
        return new AsonDecoder(input.getBytes(StandardCharsets.UTF_8)).decodeSingle(clazz);
    }

    public static <T> T decode(byte[] input, Class<T> clazz) {
        return new AsonDecoder(input).decodeSingle(clazz);
    }

    public static <T> List<T> decodeList(String input, Class<T> clazz) {
        return new AsonDecoder(input.getBytes(StandardCharsets.UTF_8)).decodeList(clazz);
    }

    public static <T> List<T> decodeList(byte[] input, Class<T> clazz) {
        return new AsonDecoder(input).decodeList(clazz);
    }

    // ========================================================================
    // Binary Encode/Decode
    // ========================================================================

    public static byte[] encodeBinary(Object value) { return AsonBinary.encode(value); }
    public static <T> T decodeBinary(byte[] data, Class<T> clazz) { return AsonBinary.decode(data, clazz); }
    public static <T> List<T> decodeBinaryList(byte[] data, Class<T> clazz) { return AsonBinary.decodeList(data, clazz); }
}
