package io.github.athxx.ason

/**
 * Represents an ASON value.
 */
sealed class Value {
    
    object Null : Value() {
        override fun toString() = "null"
    }
    
    data class Bool(val value: Boolean) : Value() {
        override fun toString() = if (value) "true" else "false"
    }
    
    data class Integer(val value: Long) : Value() {
        override fun toString() = value.toString()
    }
    
    data class Float(val value: Double) : Value() {
        override fun toString() = value.toString()
    }
    
    data class Str(val value: String) : Value() {
        override fun toString() = "\"$value\""
    }
    
    data class Array(val items: MutableList<Value> = mutableListOf()) : Value() {
        operator fun get(index: Int): Value? = items.getOrNull(index)
        fun add(item: Value) = items.add(item)
        val size: Int get() = items.size
    }
    
    data class Object(
        private val map: MutableMap<String, Value> = mutableMapOf(),
        val keys: MutableList<String> = mutableListOf()
    ) : Value() {
        operator fun get(key: String): Value? = map[key]
        
        operator fun set(key: String, value: Value) {
            if (!map.containsKey(key)) {
                keys.add(key)
            }
            map[key] = value
        }
        
        val size: Int get() = map.size
        
        fun entries(): List<Pair<String, Value>> = keys.map { it to map[it]!! }
    }
    
    // Type checks
    val isNull: Boolean get() = this is Null
    val isBool: Boolean get() = this is Bool
    val isInteger: Boolean get() = this is Integer
    val isFloat: Boolean get() = this is Float
    val isNumber: Boolean get() = isInteger || isFloat
    val isString: Boolean get() = this is Str
    val isArray: Boolean get() = this is Array
    val isObject: Boolean get() = this is Object
    
    // Value getters
    fun asBool(): Boolean = (this as? Bool)?.value ?: false
    fun asInteger(): Long = when (this) {
        is Integer -> value
        is Float -> value.toLong()
        else -> 0
    }
    fun asFloat(): Double = when (this) {
        is Float -> value
        is Integer -> value.toDouble()
        else -> 0.0
    }
    fun asString(): String = (this as? Str)?.value ?: ""
    fun asArray(): Array? = this as? Array
    fun asObject(): Object? = this as? Object
    
    companion object {
        fun ofNull() = Null
        fun ofBool(v: Boolean) = Bool(v)
        fun ofInteger(v: Long) = Integer(v)
        fun ofFloat(v: Double) = Float(v)
        fun ofString(v: String) = Str(v)
        fun ofArray(vararg items: Value) = Array(items.toMutableList())
        fun ofObject() = Object()
    }
}

