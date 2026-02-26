package io.ason;

import java.lang.invoke.MethodHandle;
import java.lang.invoke.MethodHandles;
import java.lang.invoke.MethodType;
import java.lang.reflect.*;
import java.nio.charset.StandardCharsets;
import java.util.*;
import java.util.concurrent.ConcurrentHashMap;

/**
 * Pre-computed per-class metadata for extreme serialization performance.
 * Uses MethodHandles with invokeExact() via pre-adapted types for
 * near-native field access speed - no type adaptation overhead per call.
 */
final class ClassMeta {

    private static final ConcurrentHashMap<Class<?>, ClassMeta> CACHE = new ConcurrentHashMap<>();

    final Class<?> clazz;
    final FieldMeta[] fields;
    final MethodHandle constructor;     // adapted: () -> Object
    final byte[] schemaBytes;           // pre-encoded "{field1,field2}"
    final byte[] schemaBytesVec;        // pre-encoded "[{field1,field2}]"
    private volatile boolean nestedResolved;  // lazy init flag for nested metas

    private ClassMeta(Class<?> clazz) {
        this.clazz = clazz;

        Field[] rawFields = clazz.getDeclaredFields();
        List<FieldMeta> fms = new ArrayList<>();
        MethodHandles.Lookup lookup = MethodHandles.lookup();
        for (Field f : rawFields) {
            int mods = f.getModifiers();
            if (Modifier.isStatic(mods) || Modifier.isTransient(mods)) continue;
            if (f.isSynthetic()) continue;
            f.setAccessible(true);
            fms.add(new FieldMeta(f, lookup));
        }
        this.fields = fms.toArray(new FieldMeta[0]);

        // NOTE: nestedMeta/listElemMeta are NOT initialized here to avoid
        // recursive computeIfAbsent during ClassMeta construction.
        // They are lazily initialized on first access via ensureNestedMeta().

        // Constructor — adapted to () -> Object for invokeExact
        MethodHandle ctor;
        try {
            Constructor<?> c = clazz.getDeclaredConstructor();
            c.setAccessible(true);
            ctor = lookup.unreflectConstructor(c)
                    .asType(MethodType.methodType(Object.class));
        } catch (Exception e) {
            ctor = null;
        }
        this.constructor = ctor;

        // Pre-encode schema
        StringBuilder sb = new StringBuilder(fields.length * 8);
        sb.append('{');
        for (int i = 0; i < fields.length; i++) {
            if (i > 0) sb.append(',');
            sb.append(fields[i].name);
        }
        sb.append('}');
        this.schemaBytes = sb.toString().getBytes(StandardCharsets.UTF_8);

        sb.setLength(0);
        sb.append("[{");
        for (int i = 0; i < fields.length; i++) {
            if (i > 0) sb.append(',');
            sb.append(fields[i].name);
        }
        sb.append("}]");
        this.schemaBytesVec = sb.toString().getBytes(StandardCharsets.UTF_8);
    }

    Object newInstance() {
        try {
            return constructor.invokeExact();
        } catch (Throwable e) {
            throw new AsonException("Failed to create " + clazz.getName(), e);
        }
    }

    static ClassMeta of(Class<?> clazz) {
        ClassMeta meta = CACHE.get(clazz);
        if (meta != null) {
            if (!meta.nestedResolved) meta.resolveNestedMetas();
            return meta;
        }
        meta = CACHE.computeIfAbsent(clazz, ClassMeta::new);
        if (!meta.nestedResolved) meta.resolveNestedMetas();
        return meta;
    }

    // Lazily resolve nested ClassMeta references (avoids recursive computeIfAbsent)
    private void resolveNestedMetas() {
        for (FieldMeta fm : this.fields) {
            if (fm.typeTag == FieldMeta.T_STRUCT && fm.nestedMeta == null) {
                fm.nestedMeta = ClassMeta.of(fm.type);
            } else if (fm.typeTag == FieldMeta.T_LIST && fm.elemClass != null
                    && fm.listElemMeta == null
                    && FieldMeta.computeTypeTag(fm.elemClass) == FieldMeta.T_STRUCT) {
                fm.listElemMeta = ClassMeta.of(fm.elemClass);
            }
        }
        this.nestedResolved = true;
    }

    /**
     * Per-field metadata with pre-adapted MethodHandles for invokeExact().
     * Primitive fields: getter adapted to (Object)->primitiveType, invokeExact zero-overhead.
     * Reference fields: getter adapted to (Object)->Object, invokeExact zero-overhead.
     */
    static final class FieldMeta {
        static final int T_BOOLEAN   = 0;
        static final int T_INT       = 1;
        static final int T_LONG      = 2;
        static final int T_SHORT     = 3;
        static final int T_BYTE      = 4;
        static final int T_FLOAT     = 5;
        static final int T_DOUBLE    = 6;
        static final int T_CHAR      = 7;
        static final int T_STRING    = 8;
        static final int T_OPTIONAL  = 9;
        static final int T_LIST      = 10;
        static final int T_MAP       = 11;
        static final int T_STRUCT    = 12;

        final String name;
        final byte[] nameBytes;
        final int typeTag;
        final boolean isPrimitive;
        final Class<?> type;
        final Type genericType;
        final Field field;              // original Field reference for legacy compatibility
        final MethodHandle getter;      // adapted for invokeExact: (Object)->type or (Object)->Object
        final MethodHandle setter;      // adapted for invokeExact: (Object,type)->void or (Object,Object)->void
        final MethodHandle objGetter;   // always (Object)->Object for generic access (boxes primitives)
        final MethodHandle objSetter;   // always (Object,Object)->void for generic access

        // Pre-resolved generic info
        final Class<?> elemClass;
        final Type elemType;
        final Class<?> valClass;
        final Type valType;

        // Pre-cached ClassMeta for nested struct/list-element types (avoids ConcurrentHashMap lookup)
        ClassMeta nestedMeta; // lazily set for T_STRUCT fields
        ClassMeta listElemMeta; // lazily set for T_LIST with struct elements

        FieldMeta(Field f, MethodHandles.Lookup lookup) {
            this.field = f;
            this.name = f.getName();
            this.nameBytes = name.getBytes(StandardCharsets.UTF_8);
            this.type = f.getType();
            this.genericType = f.getGenericType();
            this.typeTag = computeTypeTag(type);
            this.isPrimitive = f.getType().isPrimitive();

            // Pre-adapted MethodHandles for invokeExact
            MethodHandle g = null, s = null, og = null, os = null;
            try {
                MethodHandle rawG = lookup.unreflectGetter(f);
                MethodHandle rawS = lookup.unreflectSetter(f);
                // Object-typed handles (always work for generic access)
                og = rawG.asType(MethodType.methodType(Object.class, Object.class));
                os = rawS.asType(MethodType.methodType(void.class, Object.class, Object.class));
                // Type-specific handles (for invokeExact in primitive paths)
                if (isPrimitive) {
                    g = rawG.asType(MethodType.methodType(type, Object.class));
                    s = rawS.asType(MethodType.methodType(void.class, Object.class, type));
                } else {
                    g = og;
                    s = os;
                }
            } catch (IllegalAccessException e) {
                // leave null
            }
            this.getter = g;
            this.setter = s;
            this.objGetter = og;
            this.objSetter = os;

            // Pre-resolve generic element types
            if (genericType instanceof ParameterizedType pt) {
                Type[] args = pt.getActualTypeArguments();
                if (typeTag == T_LIST || typeTag == T_OPTIONAL) {
                    this.elemType = args[0];
                    this.elemClass = resolveClass(args[0]);
                    this.valClass = null;
                    this.valType = null;
                } else if (typeTag == T_MAP) {
                    this.elemType = args[0];
                    this.elemClass = resolveClass(args[0]);
                    this.valType = args[1];
                    this.valClass = resolveClass(args[1]);
                } else {
                    this.elemType = null; this.elemClass = null;
                    this.valType = null; this.valClass = null;
                }
            } else {
                this.elemType = null; this.elemClass = null;
                this.valType = null; this.valClass = null;
            }
        }

        // Generic Object access (uses pre-adapted objGetter/objSetter with invokeExact)
        Object get(Object obj) {
            try {
                return (Object) objGetter.invokeExact(obj);
            } catch (Throwable e) {
                throw new AsonException("Failed to get field: " + name, e);
            }
        }

        void set(Object obj, Object value) {
            try {
                objSetter.invokeExact(obj, value);
            } catch (Throwable e) {
                throw new AsonException("Failed to set field " + name + " = " + value, e);
            }
        }

        // Primitive getters — invokeExact with pre-adapted types, zero overhead
        boolean getBoolean(Object obj) {
            try { return (boolean) getter.invokeExact(obj); } catch (Throwable e) { throw new AsonException("get " + name, e); }
        }
        int getInt(Object obj) {
            try { return (int) getter.invokeExact(obj); } catch (Throwable e) { throw new AsonException("get " + name, e); }
        }
        long getLong(Object obj) {
            try { return (long) getter.invokeExact(obj); } catch (Throwable e) { throw new AsonException("get " + name, e); }
        }
        short getShort(Object obj) {
            try { return (short) getter.invokeExact(obj); } catch (Throwable e) { throw new AsonException("get " + name, e); }
        }
        byte getByte(Object obj) {
            try { return (byte) getter.invokeExact(obj); } catch (Throwable e) { throw new AsonException("get " + name, e); }
        }
        float getFloat(Object obj) {
            try { return (float) getter.invokeExact(obj); } catch (Throwable e) { throw new AsonException("get " + name, e); }
        }
        double getDouble(Object obj) {
            try { return (double) getter.invokeExact(obj); } catch (Throwable e) { throw new AsonException("get " + name, e); }
        }
        char getChar(Object obj) {
            try { return (char) getter.invokeExact(obj); } catch (Throwable e) { throw new AsonException("get " + name, e); }
        }

        // Primitive setters — invokeExact with pre-adapted types, zero overhead
        void setBoolean(Object obj, boolean v) {
            try { setter.invokeExact(obj, v); } catch (Throwable e) { throw new AsonException("set " + name, e); }
        }
        void setInt(Object obj, int v) {
            try { setter.invokeExact(obj, v); } catch (Throwable e) { throw new AsonException("set " + name, e); }
        }
        void setLong(Object obj, long v) {
            try { setter.invokeExact(obj, v); } catch (Throwable e) { throw new AsonException("set " + name, e); }
        }
        void setShort(Object obj, short v) {
            try { setter.invokeExact(obj, v); } catch (Throwable e) { throw new AsonException("set " + name, e); }
        }
        void setByte(Object obj, byte v) {
            try { setter.invokeExact(obj, v); } catch (Throwable e) { throw new AsonException("set " + name, e); }
        }
        void setFloat(Object obj, float v) {
            try { setter.invokeExact(obj, v); } catch (Throwable e) { throw new AsonException("set " + name, e); }
        }
        void setDouble(Object obj, double v) {
            try { setter.invokeExact(obj, v); } catch (Throwable e) { throw new AsonException("set " + name, e); }
        }
        void setChar(Object obj, char v) {
            try { setter.invokeExact(obj, v); } catch (Throwable e) { throw new AsonException("set " + name, e); }
        }

        static int computeTypeTag(Class<?> type) {
            if (type == boolean.class || type == Boolean.class) return T_BOOLEAN;
            if (type == int.class || type == Integer.class) return T_INT;
            if (type == long.class || type == Long.class) return T_LONG;
            if (type == short.class || type == Short.class) return T_SHORT;
            if (type == byte.class || type == Byte.class) return T_BYTE;
            if (type == float.class || type == Float.class) return T_FLOAT;
            if (type == double.class || type == Double.class) return T_DOUBLE;
            if (type == char.class || type == Character.class) return T_CHAR;
            if (type == String.class) return T_STRING;
            if (type == Optional.class) return T_OPTIONAL;
            if (List.class.isAssignableFrom(type)) return T_LIST;
            if (Map.class.isAssignableFrom(type)) return T_MAP;
            return T_STRUCT;
        }

        static Class<?> resolveClass(Type type) {
            if (type instanceof Class<?> c) return c;
            if (type instanceof ParameterizedType pt) return (Class<?>) pt.getRawType();
            return Object.class;
        }
    }
}
