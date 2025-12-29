package io.github.athxx.ason;

import java.util.*;

/**
 * Represents an ASON value.
 */
public class Value {
    
    public enum Type {
        NULL, BOOL, INTEGER, FLOAT, STRING, ARRAY, OBJECT
    }
    
    private final Type type;
    private final Object data;
    private final List<String> keys; // for preserving object key order
    
    private Value(Type type, Object data) {
        this(type, data, null);
    }
    
    private Value(Type type, Object data, List<String> keys) {
        this.type = type;
        this.data = data;
        this.keys = keys;
    }
    
    // Factory methods
    public static Value ofNull() {
        return new Value(Type.NULL, null);
    }
    
    public static Value ofBool(boolean v) {
        return new Value(Type.BOOL, v);
    }
    
    public static Value ofInteger(long v) {
        return new Value(Type.INTEGER, v);
    }
    
    public static Value ofFloat(double v) {
        return new Value(Type.FLOAT, v);
    }
    
    public static Value ofString(String v) {
        return new Value(Type.STRING, v);
    }
    
    public static Value ofArray() {
        return new Value(Type.ARRAY, new ArrayList<Value>());
    }
    
    public static Value ofArray(Value... items) {
        List<Value> list = new ArrayList<>(Arrays.asList(items));
        return new Value(Type.ARRAY, list);
    }
    
    public static Value ofObject() {
        return new Value(Type.OBJECT, new LinkedHashMap<String, Value>(), new ArrayList<>());
    }
    
    // Type checks
    public Type getType() { return type; }
    public boolean isNull() { return type == Type.NULL; }
    public boolean isBool() { return type == Type.BOOL; }
    public boolean isInteger() { return type == Type.INTEGER; }
    public boolean isFloat() { return type == Type.FLOAT; }
    public boolean isNumber() { return isInteger() || isFloat(); }
    public boolean isString() { return type == Type.STRING; }
    public boolean isArray() { return type == Type.ARRAY; }
    public boolean isObject() { return type == Type.OBJECT; }
    
    // Value getters
    public boolean asBool() {
        return type == Type.BOOL ? (Boolean) data : false;
    }
    
    public long asInteger() {
        if (type == Type.INTEGER) return (Long) data;
        if (type == Type.FLOAT) return ((Double) data).longValue();
        return 0;
    }
    
    public double asFloat() {
        if (type == Type.FLOAT) return (Double) data;
        if (type == Type.INTEGER) return ((Long) data).doubleValue();
        return 0;
    }
    
    public String asString() {
        return type == Type.STRING ? (String) data : "";
    }
    
    // Array operations
    @SuppressWarnings("unchecked")
    public int size() {
        if (type == Type.ARRAY) return ((List<Value>) data).size();
        if (type == Type.OBJECT) return ((Map<String, Value>) data).size();
        return 0;
    }
    
    @SuppressWarnings("unchecked")
    public Value get(int index) {
        if (type != Type.ARRAY) return null;
        List<Value> list = (List<Value>) data;
        if (index < 0 || index >= list.size()) return null;
        return list.get(index);
    }
    
    @SuppressWarnings("unchecked")
    public void push(Value item) {
        if (type != Type.ARRAY) return;
        ((List<Value>) data).add(item);
    }
    
    @SuppressWarnings("unchecked")
    public List<Value> items() {
        if (type != Type.ARRAY) return Collections.emptyList();
        return Collections.unmodifiableList((List<Value>) data);
    }
    
    // Object operations
    @SuppressWarnings("unchecked")
    public Value get(String key) {
        if (type != Type.OBJECT) return null;
        return ((Map<String, Value>) data).get(key);
    }
    
    @SuppressWarnings("unchecked")
    public void set(String key, Value value) {
        if (type != Type.OBJECT) return;
        Map<String, Value> map = (Map<String, Value>) data;
        if (!map.containsKey(key)) {
            keys.add(key);
        }
        map.put(key, value);
    }
    
    public List<String> keys() {
        if (type != Type.OBJECT) return Collections.emptyList();
        return Collections.unmodifiableList(keys);
    }
    
    @Override
    public String toString() {
        return "Value{type=" + type + ", data=" + data + "}";
    }
}

