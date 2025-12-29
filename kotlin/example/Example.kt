package io.github.athxx.ason

/**
 * ASON Kotlin Example
 */
fun main() {
    println("=== ASON Kotlin Example ===\n")
    
    // Test 1: Simple object
    val result1 = Ason.parse("{name,age}:(Alice,30)")
    println("Test 1 - Simple object:")
    println("  Input: {name,age}:(Alice,30)")
    println("  name = ${result1.asObject()?.get("name")?.asString()}")
    println("  age = ${result1.asObject()?.get("age")?.asInteger()}")
    println()
    
    // Test 2: Multiple objects
    val result2 = Ason.parse("{name,age}:(Alice,30),(Bob,25)")
    println("Test 2 - Multiple objects:")
    println("  Input: {name,age}:(Alice,30),(Bob,25)")
    result2.asArray()?.items?.forEachIndexed { i, user ->
        val obj = user.asObject()
        println("  [$i] name=${obj?.get("name")?.asString()}, age=${obj?.get("age")?.asInteger()}")
    }
    println()
    
    // Test 3: Nested object
    val result3 = Ason.parse("{name,addr{city,zip}}:(Alice,(NYC,10001))")
    println("Test 3 - Nested object:")
    println("  Input: {name,addr{city,zip}}:(Alice,(NYC,10001))")
    val obj3 = result3.asObject()
    println("  name = ${obj3?.get("name")?.asString()}")
    println("  addr.city = ${obj3?.get("addr")?.asObject()?.get("city")?.asString()}")
    println("  addr.zip = ${obj3?.get("addr")?.asObject()?.get("zip")?.asInteger()}")
    println()
    
    // Test 4: Array field
    val result4 = Ason.parse("{name,scores[]}:(Alice,[90,85,95])")
    println("Test 4 - Array field:")
    println("  Input: {name,scores[]}:(Alice,[90,85,95])")
    val obj4 = result4.asObject()
    println("  name = ${obj4?.get("name")?.asString()}")
    val scores = obj4?.get("scores")?.asArray()?.items?.map { it.asInteger() }
    println("  scores = $scores")
    println()
    
    // Test 5: Unicode
    val result5 = Ason.parse("{name,city}:(小明,北京)")
    println("Test 5 - Unicode:")
    println("  Input: {name,city}:(小明,北京)")
    val obj5 = result5.asObject()
    println("  name = ${obj5?.get("name")?.asString()}")
    println("  city = ${obj5?.get("city")?.asString()}")
    println()
    
    // Test 6: Serialize
    val obj = Value.ofObject()
    (obj as Value.Object)["name"] = Value.ofString("Alice")
    obj["age"] = Value.ofInteger(30)
    println("Test 6 - Serialize:")
    println("  serialize: ${Ason.serialize(obj)}")
    println("  serializeWithSchema: ${Ason.serializeWithSchema(obj)}")
    println()
    
    println("=== All tests passed! ===")
}

