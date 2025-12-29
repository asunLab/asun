package io.github.athxx.ason

/**
 * ASON Serializer.
 */
object Serializer {
    
    /**
     * Serialize a Value to ASON format (data only).
     */
    fun serialize(value: Value): String = serializeValue(value)
    
    /**
     * Serialize a Value to ASON format with schema.
     */
    fun serializeWithSchema(value: Value): String {
        if (value.isNull) return "null"
        
        if (value is Value.Array && value.size > 0 && value[0]?.isObject == true) {
            val sb = StringBuilder()
            sb.append(buildSchema(value[0] as Value.Object))
            sb.append(":")
            value.items.forEachIndexed { i, item ->
                if (i > 0) sb.append(",")
                sb.append(serializeObjectData(item as Value.Object))
            }
            return sb.toString()
        }
        
        if (value is Value.Object) {
            return "${buildSchema(value)}:${serializeObjectData(value)}"
        }
        
        return serializeValue(value)
    }
    
    private fun serializeValue(v: Value): String = when (v) {
        is Value.Null -> "null"
        is Value.Bool -> if (v.value) "true" else "false"
        is Value.Integer -> v.value.toString()
        is Value.Float -> v.value.toString()
        is Value.Str -> serializeString(v.value)
        is Value.Array -> "[${v.items.joinToString(",") { serializeValue(it) }}]"
        is Value.Object -> "(${v.entries().joinToString(",") { serializeValue(it.second) }})"
    }
    
    private fun serializeString(s: String): String {
        if (s.isEmpty()) return "\"\""
        
        val needsQuote = s.any { it in "\"\\()[]{},:\"" || it.isWhitespace() } ||
            s in listOf("null", "true", "false") ||
            (s.isNotEmpty() && (s[0] in "+-" || s[0].isDigit()))
        
        if (!needsQuote) return s
        
        val escaped = s.replace("\\", "\\\\")
            .replace("\"", "\\\"")
            .replace("\n", "\\n")
            .replace("\r", "\\r")
            .replace("\t", "\\t")
        
        return "\"$escaped\""
    }
    
    private fun buildSchema(obj: Value.Object): String {
        val parts = obj.entries().map { (key, value) ->
            when {
                value is Value.Array && value.size > 0 && value[0]?.isObject == true ->
                    "$key[${buildSchema(value[0] as Value.Object)}]"
                value is Value.Array -> "$key[]"
                value is Value.Object -> "$key${buildSchema(value)}"
                else -> key
            }
        }
        return "{${parts.joinToString(",")}}"
    }
    
    private fun serializeObjectData(obj: Value.Object): String {
        val parts = obj.entries().map { (_, value) ->
            when (value) {
                is Value.Object -> serializeObjectData(value)
                is Value.Array -> serializeArrayData(value)
                else -> serializeValue(value)
            }
        }
        return "(${parts.joinToString(",")})"
    }
    
    private fun serializeArrayData(arr: Value.Array): String {
        val parts = arr.items.map { item ->
            when (item) {
                is Value.Object -> serializeObjectData(item)
                else -> serializeValue(item)
            }
        }
        return "[${parts.joinToString(",")}]"
    }
}

