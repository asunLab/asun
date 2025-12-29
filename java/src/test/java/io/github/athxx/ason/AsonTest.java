package io.github.athxx.ason;

import org.junit.jupiter.api.Test;
import static org.junit.jupiter.api.Assertions.*;

/**
 * ASON Unit Tests
 */
class AsonTest {

    @Test
    void testParseSimpleObject() throws ParseException {
        Value result = Ason.parse("{name,age}:(Alice,30)");
        assertEquals("Alice", result.get("name").asString());
        assertEquals(30, result.get("age").asInteger());
    }

    @Test
    void testParseMultipleObjects() throws ParseException {
        Value result = Ason.parse("{name,age}:(Alice,30),(Bob,25)");
        assertEquals(2, result.size());
        assertEquals("Alice", result.get(0).get("name").asString());
        assertEquals("Bob", result.get(1).get("name").asString());
    }

    @Test
    void testParseNestedObject() throws ParseException {
        Value result = Ason.parse("{name,addr{city,zip}}:(Alice,(NYC,10001))");
        assertEquals("Alice", result.get("name").asString());
        assertEquals("NYC", result.get("addr").get("city").asString());
        assertEquals(10001, result.get("addr").get("zip").asInteger());
    }

    @Test
    void testParseArrayField() throws ParseException {
        Value result = Ason.parse("{name,scores[]}:(Alice,[90,85,95])");
        assertEquals("Alice", result.get("name").asString());
        assertEquals(3, result.get("scores").size());
        assertEquals(90, result.get("scores").get(0).asInteger());
    }

    @Test
    void testParseObjectArray() throws ParseException {
        Value result = Ason.parse("{users[{id,name}]}:([(1,Alice),(2,Bob)])");
        assertEquals(2, result.get("users").size());
        assertEquals(1, result.get("users").get(0).get("id").asInteger());
        assertEquals("Bob", result.get("users").get(1).get("name").asString());
    }

    @Test
    void testUnicode() throws ParseException {
        Value result = Ason.parse("{name,city}:(小明,北京)");
        assertEquals("小明", result.get("name").asString());
        assertEquals("北京", result.get("city").asString());
    }

    @Test
    void testBooleanValues() throws ParseException {
        Value result = Ason.parse("{active,verified}:(true,false)");
        assertTrue(result.get("active").asBool());
        assertFalse(result.get("verified").asBool());
    }

    @Test
    void testNullValue() throws ParseException {
        Value result = Ason.parse("{name,age}:(Alice,null)");
        assertEquals("Alice", result.get("name").asString());
        assertTrue(result.get("age").isNull());
    }

    @Test
    void testFloatValue() throws ParseException {
        Value result = Ason.parse("{name,score}:(Alice,3.14)");
        assertEquals("Alice", result.get("name").asString());
        assertEquals(3.14, result.get("score").asFloat(), 0.001);
    }

    @Test
    void testSerialize() {
        Value obj = Value.ofObject();
        obj.set("name", Value.ofString("Alice"));
        obj.set("age", Value.ofInteger(30));
        assertEquals("(Alice,30)", Ason.serialize(obj));
    }

    @Test
    void testSerializeWithSchema() {
        Value obj = Value.ofObject();
        obj.set("name", Value.ofString("Alice"));
        obj.set("age", Value.ofInteger(30));
        assertEquals("{name,age}:(Alice,30)", Ason.serializeWithSchema(obj));
    }

    @Test
    void testSerializeArray() {
        Value arr = Value.ofArray();
        
        Value obj1 = Value.ofObject();
        obj1.set("name", Value.ofString("Alice"));
        obj1.set("age", Value.ofInteger(30));
        arr.push(obj1);
        
        Value obj2 = Value.ofObject();
        obj2.set("name", Value.ofString("Bob"));
        obj2.set("age", Value.ofInteger(25));
        arr.push(obj2);
        
        assertEquals("{name,age}:(Alice,30),(Bob,25)", Ason.serializeWithSchema(arr));
    }

    @Test
    void testRoundTrip() throws ParseException {
        String original = "{name,age}:(Alice,30)";
        Value parsed = Ason.parse(original);
        String serialized = Ason.serializeWithSchema(parsed);
        Value reparsed = Ason.parse(serialized);
        
        assertEquals(parsed.get("name").asString(), reparsed.get("name").asString());
        assertEquals(parsed.get("age").asInteger(), reparsed.get("age").asInteger());
    }
}

