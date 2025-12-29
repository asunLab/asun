package io.github.athxx.ason;

/**
 * Main entry point for ASON library.
 */
public class Ason {
    
    /**
     * Parse an ASON string.
     */
    public static Value parse(String input) throws ParseException {
        return Parser.parse(input);
    }
    
    /**
     * Serialize a Value to ASON format (data only).
     */
    public static String serialize(Value value) {
        return Serializer.serialize(value);
    }
    
    /**
     * Serialize a Value to ASON format with schema.
     */
    public static String serializeWithSchema(Value value) {
        return Serializer.serializeWithSchema(value);
    }
}

