package io.github.athxx.ason;

/**
 * ASON Java Example
 */
public class Example {
    public static void main(String[] args) throws Exception {
        System.out.println("=== ASON Java Example ===\n");
        
        // Test 1: Simple object
        Value result1 = Ason.parse("{name,age}:(Alice,30)");
        System.out.println("Test 1 - Simple object:");
        System.out.println("  Input: {name,age}:(Alice,30)");
        System.out.println("  name = " + result1.get("name").asString());
        System.out.println("  age = " + result1.get("age").asInteger());
        System.out.println();
        
        // Test 2: Multiple objects
        Value result2 = Ason.parse("{name,age}:(Alice,30),(Bob,25)");
        System.out.println("Test 2 - Multiple objects:");
        System.out.println("  Input: {name,age}:(Alice,30),(Bob,25)");
        for (int i = 0; i < result2.size(); i++) {
            Value user = result2.get(i);
            System.out.println("  [" + i + "] name=" + user.get("name").asString() + 
                             ", age=" + user.get("age").asInteger());
        }
        System.out.println();
        
        // Test 3: Nested object
        Value result3 = Ason.parse("{name,addr{city,zip}}:(Alice,(NYC,10001))");
        System.out.println("Test 3 - Nested object:");
        System.out.println("  Input: {name,addr{city,zip}}:(Alice,(NYC,10001))");
        System.out.println("  name = " + result3.get("name").asString());
        System.out.println("  addr.city = " + result3.get("addr").get("city").asString());
        System.out.println("  addr.zip = " + result3.get("addr").get("zip").asInteger());
        System.out.println();
        
        // Test 4: Array field
        Value result4 = Ason.parse("{name,scores[]}:(Alice,[90,85,95])");
        System.out.println("Test 4 - Array field:");
        System.out.println("  Input: {name,scores[]}:(Alice,[90,85,95])");
        System.out.println("  name = " + result4.get("name").asString());
        System.out.print("  scores = [");
        Value scores = result4.get("scores");
        for (int i = 0; i < scores.size(); i++) {
            if (i > 0) System.out.print(", ");
            System.out.print(scores.get(i).asInteger());
        }
        System.out.println("]");
        System.out.println();
        
        // Test 5: Unicode
        Value result5 = Ason.parse("{name,city}:(小明,北京)");
        System.out.println("Test 5 - Unicode:");
        System.out.println("  Input: {name,city}:(小明,北京)");
        System.out.println("  name = " + result5.get("name").asString());
        System.out.println("  city = " + result5.get("city").asString());
        System.out.println();
        
        // Test 6: Serialize
        Value obj = Value.ofObject();
        obj.set("name", Value.ofString("Alice"));
        obj.set("age", Value.ofInteger(30));
        System.out.println("Test 6 - Serialize:");
        System.out.println("  serialize: " + Ason.serialize(obj));
        System.out.println("  serializeWithSchema: " + Ason.serializeWithSchema(obj));
        System.out.println();
        
        System.out.println("=== All tests passed! ===");
    }
}

