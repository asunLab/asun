package io.github.athxx.ason

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertTrue
import kotlin.test.assertFalse
import kotlin.test.assertNotNull

/**
 * ASON Unit Tests
 */
class AsonTest {

    @Test
    fun testParseSimpleObject() {
        val result = Ason.parse("{name,age}:(Alice,30)")
        val obj = result.asObject()
        assertNotNull(obj)
        assertEquals("Alice", obj["name"]?.asString())
        assertEquals(30, obj["age"]?.asInteger())
    }

    @Test
    fun testParseMultipleObjects() {
        val result = Ason.parse("{name,age}:(Alice,30),(Bob,25)")
        val arr = result.asArray()
        assertNotNull(arr)
        assertEquals(2, arr.items.size)
        assertEquals("Alice", arr.items[0].asObject()?.get("name")?.asString())
        assertEquals("Bob", arr.items[1].asObject()?.get("name")?.asString())
    }

    @Test
    fun testParseNestedObject() {
        val result = Ason.parse("{name,addr{city,zip}}:(Alice,(NYC,10001))")
        val obj = result.asObject()
        assertNotNull(obj)
        assertEquals("Alice", obj["name"]?.asString())
        assertEquals("NYC", obj["addr"]?.asObject()?.get("city")?.asString())
        assertEquals(10001, obj["addr"]?.asObject()?.get("zip")?.asInteger())
    }

    @Test
    fun testParseArrayField() {
        val result = Ason.parse("{name,scores[]}:(Alice,[90,85,95])")
        val obj = result.asObject()
        assertNotNull(obj)
        assertEquals("Alice", obj["name"]?.asString())
        val scores = obj["scores"]?.asArray()
        assertNotNull(scores)
        assertEquals(3, scores.items.size)
        assertEquals(90, scores.items[0].asInteger())
    }

    @Test
    fun testParseObjectArray() {
        val result = Ason.parse("{users[{id,name}]}:([(1,Alice),(2,Bob)])")
        val obj = result.asObject()
        assertNotNull(obj)
        val users = obj["users"]?.asArray()
        assertNotNull(users)
        assertEquals(2, users.items.size)
        assertEquals(1, users.items[0].asObject()?.get("id")?.asInteger())
        assertEquals("Bob", users.items[1].asObject()?.get("name")?.asString())
    }

    @Test
    fun testUnicode() {
        val result = Ason.parse("{name,city}:(小明,北京)")
        val obj = result.asObject()
        assertNotNull(obj)
        assertEquals("小明", obj["name"]?.asString())
        assertEquals("北京", obj["city"]?.asString())
    }

    @Test
    fun testBooleanValues() {
        val result = Ason.parse("{active,verified}:(true,false)")
        val obj = result.asObject()
        assertNotNull(obj)
        assertTrue(obj["active"]?.asBool() ?: false)
        assertFalse(obj["verified"]?.asBool() ?: true)
    }

    @Test
    fun testNullValue() {
        val result = Ason.parse("{name,age}:(Alice,null)")
        val obj = result.asObject()
        assertNotNull(obj)
        assertEquals("Alice", obj["name"]?.asString())
        assertTrue(obj["age"]?.isNull() ?: false)
    }

    @Test
    fun testFloatValue() {
        val result = Ason.parse("{name,score}:(Alice,3.14)")
        val obj = result.asObject()
        assertNotNull(obj)
        assertEquals("Alice", obj["name"]?.asString())
        assertEquals(3.14, obj["score"]?.asFloat() ?: 0.0, 0.001)
    }

    @Test
    fun testSerialize() {
        val obj = Value.ofObject() as Value.Object
        obj["name"] = Value.ofString("Alice")
        obj["age"] = Value.ofInteger(30)
        assertEquals("(Alice,30)", Ason.serialize(obj))
    }

    @Test
    fun testSerializeWithSchema() {
        val obj = Value.ofObject() as Value.Object
        obj["name"] = Value.ofString("Alice")
        obj["age"] = Value.ofInteger(30)
        assertEquals("{name,age}:(Alice,30)", Ason.serializeWithSchema(obj))
    }

    @Test
    fun testRoundTrip() {
        val original = "{name,age}:(Alice,30)"
        val parsed = Ason.parse(original)
        val serialized = Ason.serializeWithSchema(parsed)
        val reparsed = Ason.parse(serialized)
        
        assertEquals(
            parsed.asObject()?.get("name")?.asString(),
            reparsed.asObject()?.get("name")?.asString()
        )
        assertEquals(
            parsed.asObject()?.get("age")?.asInteger(),
            reparsed.asObject()?.get("age")?.asInteger()
        )
    }
}

