package io.ason;

import java.lang.reflect.*;
import java.nio.ByteOrder;
import java.nio.charset.StandardCharsets;
import java.util.*;

/**
 * ASON Binary codec — high-performance binary encoding/decoding.
 * <p>
 * Wire format (all integers little-endian):
 * <pre>
 * bool      → 1 byte  (0x00=false, 0x01=true)
 * byte      → 1 byte
 * short     → 2 bytes LE
 * int       → 4 bytes LE
 * long      → 8 bytes LE
 * float     → 4 bytes LE (IEEE 754 bit-cast)
 * double    → 8 bytes LE (IEEE 754 bit-cast)
 * char      → 4 bytes LE (Unicode scalar as int)
 * String    → u32 LE length + UTF-8 bytes
 * Optional  → u8 tag (0=None, 1=Some) + [payload if Some]
 * List      → u32 LE count + [element × count]
 * Map       → u32 LE count + [(key, value) × count]
 * struct    → fields in declaration order (no length prefix)
 * </pre>
 */
final class AsonBinary {

    private AsonBinary() {}

    // ========================================================================
    // Encode
    // ========================================================================

    static byte[] encode(Object value) {
        if (value instanceof List<?> list) {
            return encodeList(list);
        }
        ByteBuffer buf = new ByteBuffer(256);
        writeObject(buf, value, value.getClass(), value.getClass());
        return buf.toBytes();
    }

    private static byte[] encodeList(List<?> list) {
        ByteBuffer buf = new ByteBuffer(256);
        buf.appendLEU32(list.size());
        for (Object item : list) {
            writeObject(buf, item, item.getClass(), item.getClass());
        }
        return buf.toBytes();
    }

    @SuppressWarnings("unchecked")
    private static void writeObject(ByteBuffer buf, Object value, Class<?> type, Type genericType) {
        if (value == null) return;

        if (value instanceof Optional<?> opt) {
            if (opt.isPresent()) {
                buf.append((byte) 1);
                Object inner = opt.get();
                writeObject(buf, inner, inner.getClass(), inner.getClass());
            } else {
                buf.append((byte) 0);
            }
            return;
        }

        if (type == boolean.class || type == Boolean.class) {
            buf.append((byte) ((Boolean) value ? 1 : 0));
        } else if (type == byte.class || type == Byte.class) {
            buf.append((Byte) value);
        } else if (type == short.class || type == Short.class) {
            buf.appendLEU16((Short) value);
        } else if (type == int.class || type == Integer.class) {
            buf.appendLEU32((Integer) value);
        } else if (type == long.class || type == Long.class) {
            buf.appendLEU64((Long) value);
        } else if (type == float.class || type == Float.class) {
            buf.appendLEU32(Float.floatToRawIntBits((Float) value));
        } else if (type == double.class || type == Double.class) {
            buf.appendLEU64(Double.doubleToRawLongBits((Double) value));
        } else if (type == char.class || type == Character.class) {
            buf.appendLEU32((Character) value);
        } else if (type == String.class) {
            byte[] bytes = ((String) value).getBytes(StandardCharsets.UTF_8);
            buf.appendLEU32(bytes.length);
            buf.appendBytes(bytes, 0, bytes.length);
        } else if (List.class.isAssignableFrom(type)) {
            List<?> list = (List<?>) value;
            buf.appendLEU32(list.size());
            Type elemType = Object.class;
            if (genericType instanceof ParameterizedType pt) {
                elemType = pt.getActualTypeArguments()[0];
            }
            Class<?> elemClass = (elemType instanceof Class<?> c) ? c : Object.class;
            for (Object item : list) {
                if (item != null) {
                    writeObject(buf, item, elemClass, elemType);
                }
            }
        } else if (Map.class.isAssignableFrom(type)) {
            Map<?, ?> map = (Map<?, ?>) value;
            buf.appendLEU32(map.size());
            Type keyType = Object.class, valType = Object.class;
            if (genericType instanceof ParameterizedType pt) {
                Type[] args = pt.getActualTypeArguments();
                keyType = args[0];
                valType = args[1];
            }
            Class<?> keyClass = (keyType instanceof Class<?> c) ? c : Object.class;
            Class<?> valClass = (valType instanceof Class<?> c) ? c : Object.class;
            for (Map.Entry<?, ?> entry : map.entrySet()) {
                writeObject(buf, entry.getKey(), keyClass, keyType);
                writeObject(buf, entry.getValue(), valClass, valType);
            }
        } else {
            // Struct: write fields in order
            Field[] fields = Ason.getFields(type);
            for (Field f : fields) {
                try {
                    f.setAccessible(true);
                    Object fv = f.get(value);
                    writeObject(buf, fv, f.getType(), f.getGenericType());
                } catch (IllegalAccessException e) {
                    throw new AsonException("Failed to read field: " + f.getName(), e);
                }
            }
        }
    }

    // ========================================================================
    // Decode
    // ========================================================================

    static <T> T decode(byte[] data, Class<T> clazz) {
        int[] pos = {0};
        return readObject(data, pos, clazz, clazz);
    }

    static <T> List<T> decodeList(byte[] data, Class<T> clazz) {
        int[] pos = {0};
        int count = readU32(data, pos);
        List<T> result = new ArrayList<>(count);
        for (int i = 0; i < count; i++) {
            result.add(readObject(data, pos, clazz, clazz));
        }
        return result;
    }

    @SuppressWarnings("unchecked")
    private static <T> T readObject(byte[] data, int[] pos, Class<T> type, Type genericType) {
        if (type == boolean.class || type == Boolean.class) {
            return (T) Boolean.valueOf(data[pos[0]++] != 0);
        }
        if (type == byte.class || type == Byte.class) {
            return (T) Byte.valueOf(data[pos[0]++]);
        }
        if (type == short.class || type == Short.class) {
            return (T) Short.valueOf(readI16(data, pos));
        }
        if (type == int.class || type == Integer.class) {
            return (T) Integer.valueOf(readI32(data, pos));
        }
        if (type == long.class || type == Long.class) {
            return (T) Long.valueOf(readI64(data, pos));
        }
        if (type == float.class || type == Float.class) {
            return (T) Float.valueOf(Float.intBitsToFloat(readU32(data, pos)));
        }
        if (type == double.class || type == Double.class) {
            return (T) Double.valueOf(Double.longBitsToDouble(readU64(data, pos)));
        }
        if (type == char.class || type == Character.class) {
            return (T) Character.valueOf((char) readU32(data, pos));
        }
        if (type == String.class) {
            int len = readU32(data, pos);
            String s = new String(data, pos[0], len, StandardCharsets.UTF_8);
            pos[0] += len;
            return (T) s;
        }
        if (type == Optional.class) {
            byte tag = data[pos[0]++];
            if (tag == 0) return (T) Optional.empty();
            Type innerType = Object.class;
            if (genericType instanceof ParameterizedType pt) {
                innerType = pt.getActualTypeArguments()[0];
            }
            Class<?> innerClass = (innerType instanceof Class<?> c) ? c : Object.class;
            return (T) Optional.of(readObject(data, pos, innerClass, innerType));
        }
        if (List.class.isAssignableFrom(type)) {
            int count = readU32(data, pos);
            Type elemType = Object.class;
            if (genericType instanceof ParameterizedType pt) {
                elemType = pt.getActualTypeArguments()[0];
            }
            Class<?> elemClass = (elemType instanceof Class<?> c) ? c : Object.class;
            List<Object> list = new ArrayList<>(count);
            for (int i = 0; i < count; i++) {
                list.add(readObject(data, pos, elemClass, elemType));
            }
            return (T) list;
        }
        if (Map.class.isAssignableFrom(type)) {
            int count = readU32(data, pos);
            Type keyType = Object.class, valType = Object.class;
            if (genericType instanceof ParameterizedType pt) {
                Type[] args = pt.getActualTypeArguments();
                keyType = args[0];
                valType = args[1];
            }
            Class<?> keyClass = (keyType instanceof Class<?> c) ? c : Object.class;
            Class<?> valClass = (valType instanceof Class<?> c) ? c : Object.class;
            Map<Object, Object> map = new LinkedHashMap<>(count);
            for (int i = 0; i < count; i++) {
                Object k = readObject(data, pos, keyClass, keyType);
                Object v = readObject(data, pos, valClass, valType);
                map.put(k, v);
            }
            return (T) map;
        }

        // Struct: read fields in order
        try {
            T obj = type.getDeclaredConstructor().newInstance();
            Field[] fields = Ason.getFields(type);
            for (Field f : fields) {
                f.setAccessible(true);
                Object val = readObject(data, pos, f.getType(), f.getGenericType());
                f.set(obj, val);
            }
            return obj;
        } catch (ReflectiveOperationException e) {
            throw new AsonException("Failed to create instance of " + type.getName(), e);
        }
    }

    // ========================================================================
    // Primitive readers (little-endian)
    // ========================================================================

    private static int readU32(byte[] data, int[] pos) {
        int p = pos[0];
        int v = (data[p] & 0xFF) | ((data[p + 1] & 0xFF) << 8) |
            ((data[p + 2] & 0xFF) << 16) | ((data[p + 3] & 0xFF) << 24);
        pos[0] = p + 4;
        return v;
    }

    private static long readU64(byte[] data, int[] pos) {
        int p = pos[0];
        long v = (data[p] & 0xFFL) | ((data[p + 1] & 0xFFL) << 8) |
            ((data[p + 2] & 0xFFL) << 16) | ((data[p + 3] & 0xFFL) << 24) |
            ((data[p + 4] & 0xFFL) << 32) | ((data[p + 5] & 0xFFL) << 40) |
            ((data[p + 6] & 0xFFL) << 48) | ((data[p + 7] & 0xFFL) << 56);
        pos[0] = p + 8;
        return v;
    }

    private static short readI16(byte[] data, int[] pos) {
        int p = pos[0];
        short v = (short) ((data[p] & 0xFF) | ((data[p + 1] & 0xFF) << 8));
        pos[0] = p + 2;
        return v;
    }

    private static int readI32(byte[] data, int[] pos) {
        return readU32(data, pos);
    }

    private static long readI64(byte[] data, int[] pos) {
        return readU64(data, pos);
    }
}
