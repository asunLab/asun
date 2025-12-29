package io.github.athxx.ason

/**
 * Main entry point for ASON library.
 */
object Ason {
    
    /**
     * Parse an ASON string.
     */
    fun parse(input: String): Value = Parser.parse(input)
    
    /**
     * Serialize a Value to ASON format (data only).
     */
    fun serialize(value: Value): String = Serializer.serialize(value)
    
    /**
     * Serialize a Value to ASON format with schema.
     */
    fun serializeWithSchema(value: Value): String = Serializer.serializeWithSchema(value)
}

