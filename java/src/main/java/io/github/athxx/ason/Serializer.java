package io.github.athxx.ason;

/**
 * ASON Serializer.
 */
public class Serializer {
    
    /**
     * Serialize a Value to ASON format (data only).
     */
    public static String serialize(Value value) {
        return serializeValue(value);
    }
    
    /**
     * Serialize a Value to ASON format with schema.
     */
    public static String serializeWithSchema(Value value) {
        if (value == null || value.isNull()) {
            return "null";
        }
        
        if (value.isArray() && value.size() > 0 && value.get(0).isObject()) {
            // Array of objects
            StringBuilder sb = new StringBuilder();
            sb.append(buildSchema(value.get(0)));
            sb.append(":");
            for (int i = 0; i < value.size(); i++) {
                if (i > 0) sb.append(",");
                sb.append(serializeObjectData(value.get(i)));
            }
            return sb.toString();
        }
        
        if (value.isObject()) {
            return buildSchema(value) + ":" + serializeObjectData(value);
        }
        
        return serializeValue(value);
    }
    
    private static String serializeValue(Value v) {
        if (v == null || v.isNull()) return "null";
        if (v.isBool()) return v.asBool() ? "true" : "false";
        if (v.isInteger()) return String.valueOf(v.asInteger());
        if (v.isFloat()) return String.valueOf(v.asFloat());
        if (v.isString()) return serializeString(v.asString());
        if (v.isArray()) {
            StringBuilder sb = new StringBuilder("[");
            for (int i = 0; i < v.size(); i++) {
                if (i > 0) sb.append(",");
                sb.append(serializeValue(v.get(i)));
            }
            sb.append("]");
            return sb.toString();
        }
        if (v.isObject()) {
            StringBuilder sb = new StringBuilder("(");
            boolean first = true;
            for (String key : v.keys()) {
                if (!first) sb.append(",");
                first = false;
                sb.append(serializeValue(v.get(key)));
            }
            sb.append(")");
            return sb.toString();
        }
        return "null";
    }
    
    private static String serializeString(String s) {
        if (s == null || s.isEmpty()) return "\"\"";
        
        boolean needsQuote = false;
        for (char ch : s.toCharArray()) {
            if ("\"\\()[]{},:".indexOf(ch) >= 0 || Character.isWhitespace(ch)) {
                needsQuote = true;
                break;
            }
        }
        
        if (!needsQuote) {
            if (s.equals("null") || s.equals("true") || s.equals("false")) {
                needsQuote = true;
            } else if (!s.isEmpty() && (s.charAt(0) == '+' || s.charAt(0) == '-' || Character.isDigit(s.charAt(0)))) {
                needsQuote = true;
            }
        }
        
        if (!needsQuote) return s;
        
        StringBuilder sb = new StringBuilder("\"");
        for (char ch : s.toCharArray()) {
            switch (ch) {
                case '"': sb.append("\\\""); break;
                case '\\': sb.append("\\\\"); break;
                case '\n': sb.append("\\n"); break;
                case '\r': sb.append("\\r"); break;
                case '\t': sb.append("\\t"); break;
                default: sb.append(ch);
            }
        }
        sb.append("\"");
        return sb.toString();
    }
    
    private static String buildSchema(Value obj) {
        StringBuilder sb = new StringBuilder("{");
        boolean first = true;
        for (String key : obj.keys()) {
            if (!first) sb.append(",");
            first = false;
            
            Value val = obj.get(key);
            if (val.isArray()) {
                if (val.size() > 0 && val.get(0).isObject()) {
                    sb.append(key).append("[").append(buildSchema(val.get(0))).append("]");
                } else {
                    sb.append(key).append("[]");
                }
            } else if (val.isObject()) {
                sb.append(key).append(buildSchema(val));
            } else {
                sb.append(key);
            }
        }
        sb.append("}");
        return sb.toString();
    }
    
    private static String serializeObjectData(Value obj) {
        StringBuilder sb = new StringBuilder("(");
        boolean first = true;
        for (String key : obj.keys()) {
            if (!first) sb.append(",");
            first = false;
            
            Value val = obj.get(key);
            if (val.isObject()) {
                sb.append(serializeObjectData(val));
            } else if (val.isArray()) {
                sb.append(serializeArrayData(val));
            } else {
                sb.append(serializeValue(val));
            }
        }
        sb.append(")");
        return sb.toString();
    }
    
    private static String serializeArrayData(Value arr) {
        StringBuilder sb = new StringBuilder("[");
        for (int i = 0; i < arr.size(); i++) {
            if (i > 0) sb.append(",");
            Value item = arr.get(i);
            if (item.isObject()) {
                sb.append(serializeObjectData(item));
            } else {
                sb.append(serializeValue(item));
            }
        }
        sb.append("]");
        return sb.toString();
    }
}

