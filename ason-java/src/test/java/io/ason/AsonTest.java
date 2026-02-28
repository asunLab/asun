package io.ason;

import org.junit.jupiter.api.Test;
import java.util.*;
import static org.junit.jupiter.api.Assertions.*;

public class AsonTest {

    // ========================================================================
    // Test structs
    // ========================================================================

    public static class User {
        public long id;
        public String name;
        public boolean active;
        public User() {}
        public User(long id, String name, boolean active) {
            this.id = id; this.name = name; this.active = active;
        }
        @Override public boolean equals(Object o) {
            if (this == o) return true;
            if (!(o instanceof User u)) return false;
            return id == u.id && active == u.active && Objects.equals(name, u.name);
        }
        @Override public int hashCode() { return Objects.hash(id, name, active); }
        @Override public String toString() { return "User{id=" + id + ", name=" + name + ", active=" + active + "}"; }
    }

    public static class Item {
        public long id;
        public Optional<String> label;
        public Item() { label = Optional.empty(); }
        public Item(long id, Optional<String> label) { this.id = id; this.label = label; }
    }

    public static class Tagged {
        public String name;
        public List<String> tags;
        public Tagged() { tags = new ArrayList<>(); }
    }

    public static class Dept {
        public String title;
        public Dept() {}
        public Dept(String title) { this.title = title; }
        @Override public boolean equals(Object o) {
            if (this == o) return true;
            if (!(o instanceof Dept d)) return false;
            return Objects.equals(title, d.title);
        }
    }

    public static class Employee {
        public String name;
        public Dept dept;
        public Employee() {}
        public Employee(String name, Dept dept) { this.name = name; this.dept = dept; }
        @Override public boolean equals(Object o) {
            if (this == o) return true;
            if (!(o instanceof Employee e)) return false;
            return Objects.equals(name, e.name) && Objects.equals(dept, e.dept);
        }
    }

    public static class Score {
        public long id;
        public double value;
        public Score() {}
        public Score(long id, double value) { this.id = id; this.value = value; }
        @Override public boolean equals(Object o) {
            if (this == o) return true;
            if (!(o instanceof Score s)) return false;
            return id == s.id && Double.compare(s.value, value) == 0;
        }
    }

    public static class Row {
        public long id;
        public String name;
        public Row() {}
        public Row(long id, String name) { this.id = id; this.name = name; }
        @Override public boolean equals(Object o) {
            if (this == o) return true;
            if (!(o instanceof Row r)) return false;
            return id == r.id && Objects.equals(name, r.name);
        }
    }

    public static class Note {
        public String text;
        public Note() {}
        public Note(String text) { this.text = text; }
        @Override public boolean equals(Object o) {
            if (this == o) return true;
            if (!(o instanceof Note n)) return false;
            return Objects.equals(text, n.text);
        }
    }

    public static class AllJava {
        public boolean b;
        public int i32v;
        public long i64v;
        public float f32v;
        public double f64v;
        public String s;
        public Optional<Long> optSome;
        public Optional<Long> optNone;
        public List<Long> vecInt;
        public List<String> vecStr;
        public AllJava() { optSome = Optional.empty(); optNone = Optional.empty(); vecInt = new ArrayList<>(); vecStr = new ArrayList<>(); }
        @Override public boolean equals(Object o) {
            if (this == o) return true;
            if (!(o instanceof AllJava a)) return false;
            return b == a.b && i32v == a.i32v && i64v == a.i64v &&
                Float.compare(a.f32v, f32v) == 0 && Double.compare(a.f64v, f64v) == 0 &&
                Objects.equals(s, a.s) && Objects.equals(optSome, a.optSome) &&
                Objects.equals(optNone, a.optNone) && Objects.equals(vecInt, a.vecInt) &&
                Objects.equals(vecStr, a.vecStr);
        }
    }

    public static class WithMap {
        public String name;
        public Map<String, Long> attrs;
        public WithMap() { attrs = new LinkedHashMap<>(); }
    }

    // ========================================================================
    // Encode tests
    // ========================================================================

    @Test void testSerializeStruct() {
        User u = new User(1, "Alice", true);
        String s = Ason.encode(u);
        assertEquals("{id,name,active}:(1,Alice,true)", s);
    }

    @Test void testSerializeVec() {
        List<Row> rows = List.of(new Row(1, "Alice"), new Row(2, "Bob"));
        String s = Ason.encode(rows);
        assertEquals("[{id,name}]:(1,Alice),(2,Bob)", s);
    }

    @Test void testSerializeTyped() {
        User u = new User(1, "Alice", true);
        String s = Ason.encodeTyped(u);
        assertEquals("{id:int,name:str,active:bool}:(1,Alice,true)", s);
    }

    @Test void testSerializeTypedVec() {
        List<Row> rows = List.of(new Row(1, "Alice"), new Row(2, "Bob"));
        String s = Ason.encodeTyped(rows);
        assertEquals("[{id:int,name:str}]:(1,Alice),(2,Bob)", s);
    }

    // ========================================================================
    // Decode tests
    // ========================================================================

    @Test void testDeserializeStructWithSchema() {
        User u = Ason.decode("{id,name,active}:(1,Alice,true)", User.class);
        assertEquals(1, u.id);
        assertEquals("Alice", u.name);
        assertTrue(u.active);
    }

    @Test void testDeserializeStructWithTypedSchema() {
        User u = Ason.decode("{id:int,name:str,active:bool}:(1,Alice,true)", User.class);
        assertEquals(1, u.id);
        assertEquals("Alice", u.name);
        assertTrue(u.active);
    }

    @Test void testRoundtrip() {
        User original = new User(42, "Bob", false);
        String s = Ason.encode(original);
        User decoded = Ason.decode(s, User.class);
        assertEquals(original, decoded);
    }

    @Test void testVecDeserialize() {
        String input = "[{id:int,name:str,active:bool}]:(1,Alice,true),(2,Bob,false)";
        List<User> users = Ason.decodeList(input, User.class);
        assertEquals(2, users.size());
        assertEquals("Alice", users.get(0).name);
        assertEquals("Bob", users.get(1).name);
    }

    @Test void testMultiline() {
        String input = """
            [{id:int,name:str,active:bool}]:
              (1, Alice, true),
              (2, Bob, false)""";
        List<User> users = Ason.decodeList(input, User.class);
        assertEquals(2, users.size());
    }

    @Test void testQuotedString() {
        User u = Ason.decode("{id,name,active}:(1,\"Carol Smith\",true)", User.class);
        assertEquals("Carol Smith", u.name);
    }

    @Test void testOptionalField() {
        Item item = Ason.decode("{id,label}:(1,)", Item.class);
        assertEquals(1, item.id);
        assertEquals(Optional.empty(), item.label);
    }

    @Test void testOptionalFieldWithValue() {
        Item item = Ason.decode("{id,label}:(1,hello)", Item.class);
        assertEquals(1, item.id);
        assertEquals(Optional.of("hello"), item.label);
    }

    @Test void testArrayField() {
        Tagged t = Ason.decode("{name,tags}:(Alice,[rust,go])", Tagged.class);
        assertEquals("Alice", t.name);
        assertEquals(List.of("rust", "go"), t.tags);
    }

    @Test void testNestedStruct() {
        Employee e = Ason.decode("{name,dept:{title}}:(Alice,(Manager))", Employee.class);
        assertEquals("Alice", e.name);
        assertEquals("Manager", e.dept.title);
    }

    @Test void testNestedStructEncode() {
        Employee e = new Employee("Alice", new Dept("Manager"));
        String s = Ason.encode(e);
        assertEquals("{name,dept:{title}}:(Alice,(Manager))", s);
    }

    @Test void testNestedStructEncodeTyped() {
        Employee e = new Employee("Alice", new Dept("Manager"));
        String s = Ason.encodeTyped(e);
        assertEquals("{name:str,dept:{title:str}}:(Alice,(Manager))", s);
    }

    @Test void testNestedStructRoundtrip() {
        Employee original = new Employee("Bob", new Dept("Engineering"));
        String s = Ason.encode(original);
        Employee decoded = Ason.decode(s, Employee.class);
        assertEquals(original, decoded);
    }

    @Test void testFloatField() {
        Score s = Ason.decode("{id,value}:(1,95.5)", Score.class);
        assertEquals(1, s.id);
        assertEquals(95.5, s.value);
    }

    @Test void testEscapeRoundtrip() {
        Note note = new Note("hello, world (test)");
        String s = Ason.encode(note);
        Note note2 = Ason.decode(s, Note.class);
        assertEquals(note, note2);
    }

    @Test void testCommentStripping() {
        User u = Ason.decode("/* users */ {id,name,active}:(1,Alice,true)", User.class);
        assertEquals(1, u.id);
    }

    @Test void testMapField() {
        WithMap item = Ason.decode("{name,attrs}:(Alice,[(age,30),(score,95)])", WithMap.class);
        assertEquals("Alice", item.name);
        assertEquals(30L, item.attrs.get("age"));
        assertEquals(95L, item.attrs.get("score"));
    }

    // ========================================================================
    // Annotated vs Unannotated schema tests
    // ========================================================================

    @Test void testAnnotatedSimpleStruct() {
        String typed = "{id:int,name:str,active:bool}:(42,Bob,false)";
        String untyped = "{id,name,active}:(42,Bob,false)";
        User u1 = Ason.decode(typed, User.class);
        User u2 = Ason.decode(untyped, User.class);
        assertEquals(u1, u2);
        assertEquals(42, u1.id);
    }

    @Test void testAnnotatedVec() {
        String typed = "[{id:int,name:str,active:bool}]:(1,Alice,true),(2,Bob,false)";
        String untyped = "[{id,name,active}]:(1,Alice,true),(2,Bob,false)";
        List<User> v1 = Ason.decodeList(typed, User.class);
        List<User> v2 = Ason.decodeList(untyped, User.class);
        assertEquals(v1.size(), v2.size());
        assertEquals(v1.get(0), v2.get(0));
        assertEquals(v1.get(1), v2.get(1));
    }

    @Test void testAnnotatedMixedPartial() {
        String partial = "{id:int,name,active}:(1,Alice,true)";
        String full = "{id:int,name:str,active:bool}:(1,Alice,true)";
        String none = "{id,name,active}:(1,Alice,true)";
        User m1 = Ason.decode(partial, User.class);
        User m2 = Ason.decode(full, User.class);
        User m3 = Ason.decode(none, User.class);
        assertEquals(m1, m2);
        assertEquals(m2, m3);
    }

    // ========================================================================
    // Typed serialization tests
    // ========================================================================

    @Test void testEncodeTypedSimple() {
        User u = new User(1, "Alice", true);
        String s = Ason.encodeTyped(u);
        assertEquals("{id:int,name:str,active:bool}:(1,Alice,true)", s);
    }

    @Test void testEncodeTypedRoundtrip() {
        User u = new User(42, "Bob", false);
        String s = Ason.encodeTyped(u);
        assertTrue(s.startsWith("{id:int,name:str,active:bool}:"));
        User u2 = Ason.decode(s, User.class);
        assertEquals(u, u2);
    }

    @Test void testEncodeTypedFloats() {
        Score score = new Score(1, 95.5);
        String s = Ason.encodeTyped(score);
        assertEquals("{id:int,value:float}:(1,95.5)", s);
    }

    @Test void testSerializerOutputIsUnannotated() {
        User u = new User(1, "Alice", true);
        String s = Ason.encode(u);
        assertEquals("{id,name,active}:(1,Alice,true)", s);
        assertFalse(s.contains(":int"));
        assertFalse(s.contains(":str"));
        assertFalse(s.contains(":bool"));
    }

    // ========================================================================
    // Binary tests
    // ========================================================================

    @Test void testBinaryRoundtrip() {
        User u = new User(42, "Test User", true);
        byte[] bin = Ason.encodeBinary(u);
        User u2 = Ason.decodeBinary(bin, User.class);
        assertEquals(u, u2);
    }

    @Test void testBinaryVecRoundtrip() {
        List<User> users = List.of(
            new User(1, "Alice", true),
            new User(2, "Bob", false)
        );
        byte[] bin = Ason.encodeBinary(users);
        List<User> users2 = Ason.decodeBinaryList(bin, User.class);
        assertEquals(users.size(), users2.size());
        assertEquals(users.get(0), users2.get(0));
        assertEquals(users.get(1), users2.get(1));
    }

    @Test void testBinaryScore() {
        Score s = new Score(1, 95.5);
        byte[] bin = Ason.encodeBinary(s);
        Score s2 = Ason.decodeBinary(bin, Score.class);
        assertEquals(s, s2);
    }

    @Test void testBinaryNestedStruct() {
        Employee e = new Employee("Alice", new Dept("Engineering"));
        byte[] bin = Ason.encodeBinary(e);
        Employee e2 = Ason.decodeBinary(bin, Employee.class);
        assertEquals(e, e2);
    }

    // ========================================================================
    // Edge cases
    // ========================================================================

    @Test void testNegativeNumbers() {
        Score s = new Score(-42, -3.15);
        String encoded = Ason.encode(s);
        Score s2 = Ason.decode(encoded, Score.class);
        assertEquals(s, s2);
    }

    @Test void testBoolLikeString() {
        Note n = new Note("true");
        String s = Ason.encode(n);
        Note n2 = Ason.decode(s, Note.class);
        assertEquals(n, n2);
    }

    @Test void testNumberLikeString() {
        Note n = new Note("12345");
        String s = Ason.encode(n);
        Note n2 = Ason.decode(s, Note.class);
        assertEquals(n, n2);
    }

    @Test void testEmptyString() {
        Note n = new Note("");
        String s = Ason.encode(n);
        assertTrue(s.contains("\"\""));
        Note n2 = Ason.decode(s, Note.class);
        assertEquals(n, n2);
    }

    @Test void testSpecialCharsRoundtrip() {
        Note n = new Note("tabs\there, newlines\nhere, quotes\"and\\backslash");
        String s = Ason.encode(n);
        Note n2 = Ason.decode(s, Note.class);
        assertEquals(n, n2);
    }

    public static class Matrix3D {
        public List<List<List<Long>>> data;
        public Matrix3D() { data = new ArrayList<>(); }
        @Override public boolean equals(Object o) {
            if (this == o) return true;
            if (!(o instanceof Matrix3D m)) return false;
            return Objects.equals(data, m.data);
        }
    }

    @Test void testTripleNestedArrays() {
        Matrix3D m = new Matrix3D();
        m.data = List.of(
            List.of(List.of(1L, 2L), List.of(3L, 4L)),
            List.of(List.of(5L, 6L, 7L), List.of(8L))
        );
        String s = Ason.encode(m);
        Matrix3D m2 = Ason.decode(s, Matrix3D.class);
        assertEquals(m, m2);
    }

    // ========================================================================
    // Pretty encode tests
    // ========================================================================

    @Test void testPrettySimple() {
        User u = new User(1, "Alice", true);
        String pretty = Ason.encodePretty(u);
        assertEquals("{id, name, active}:(1, Alice, true)", pretty);
    }

    @Test void testPrettyTyped() {
        User u = new User(1, "Alice", true);
        String pretty = Ason.encodePrettyTyped(u);
        assertTrue(pretty.contains(":"));
        // Should have type annotations like int, str, bool
        assertTrue(pretty.contains("int") || pretty.contains("str") || pretty.contains("bool"));
    }

    @Test void testPrettyArray() {
        List<User> users = List.of(
            new User(1, "Alice", true),
            new User(2, "Bob", false)
        );
        String pretty = Ason.encodePretty(users);
        // Array format should have newlines for multiple tuples
        assertTrue(pretty.contains("[{id, name, active}]:"));
        assertTrue(pretty.contains("(1, Alice, true)"));
        assertTrue(pretty.contains("(2, Bob, false)"));
    }

    @Test void testPrettyRoundtrip() {
        User u = new User(42, "Test", true);
        String pretty = Ason.encodePretty(u);
        User decoded = Ason.decode(pretty, User.class);
        assertEquals(u, decoded);
    }

    @Test void testPrettyFormat() {
        String compact = "{id,name,active}:(1,Alice,true)";
        String pretty = Ason.prettyFormat(compact);
        assertEquals("{id, name, active}:(1, Alice, true)", pretty);
    }

    // ========================================================================
    // Format validation: {schema}: must be rejected for List; [{schema}]: required
    // ========================================================================

    static final String BAD_FMT  = "{id:int,name:str}:\n  (1,Alice),\n  (2,Bob),\n  (3,Carol)";
    static final String GOOD_FMT = "[{id:int,name:str}]:\n  (1,Alice),\n  (2,Bob),\n  (3,Carol)";

    @Test void testBadFormatAsListThrows() {
        assertThrows(Exception.class, () -> Ason.decodeList(BAD_FMT, Row.class),
            "should reject {schema}: for List decode");
    }

    @Test void testBadFormatTrailingRowsThrows() {
        // {schema}:(row),(row),... first struct decoded but trailing rows → error
        assertThrows(Exception.class, () -> Ason.decode(BAD_FMT, Row.class),
            "should reject trailing rows after single struct decode");
    }

    @Test void testGoodFormatAsList() {
        List<Row> rows = Ason.decodeList(GOOD_FMT, Row.class);
        assertEquals(3, rows.size());
        assertEquals(1L, rows.get(0).id);
        assertEquals("Alice", rows.get(0).name);
        assertEquals("Carol", rows.get(2).name);
    }

    @Test void testGoodFormatSingleStruct() {
        // {id:int,name:str}: (1, Alice) — single struct, one tuple: MUST succeed
        Row r = Ason.decode("{id:int,name:str}:(1,Alice)", Row.class);
        assertEquals(1L, r.id);
        assertEquals("Alice", r.name);
    }

    @Test void testGoodFormatVecSingleItem() {
        // [{id:int,name:str}]: (1, Alice) — array schema, one tuple: MUST succeed
        List<Row> rows = Ason.decodeList("[{id:int,name:str}]:(1,Alice)", Row.class);
        assertEquals(1, rows.size());
        assertEquals(1L, rows.get(0).id);
        assertEquals("Alice", rows.get(0).name);
    }

    @Test void testBadFormatExtraTuplesThrows() {
        String bad = "{id:int,name:str}:(10,Dave),(11,Eve)";
        assertThrows(Exception.class, () -> Ason.decode(bad, Row.class),
            "should reject trailing tuple after single struct");
    }

    @Test void testBadFormatNoVecWrapperManyRows() {
        String bad = "{id,name}:(1,A),(2,B),(3,C),(4,D),(5,E)";
        assertThrows(Exception.class, () -> Ason.decodeList(bad, Row.class),
            "should reject {schema}: without [] for list");
    }

    // ========================================================================
    // Additional encodePretty -> decode roundtrip tests (complex)
    // ========================================================================

    @Test void testPrettySliceRoundtrip() {
        List<Row> rows = List.of(new Row(1, "Alice"), new Row(2, "Bob"), new Row(3, "Carol"));
        String pretty = Ason.encodePretty(rows);
        assertTrue(pretty.contains("\n"), "expected multi-line for slice");
        List<Row> decoded = Ason.decodeList(pretty, Row.class);
        assertEquals(rows, decoded);
    }

    @Test void testPrettyTypedRoundtrip() {
        User u = new User(99, "Zara", false);
        String pretty = Ason.encodePrettyTyped(u);
        assertTrue(pretty.contains(":int") || pretty.contains(":str") || pretty.contains(":bool"));
        User decoded = Ason.decode(pretty, User.class);
        assertEquals(u, decoded);
    }

    @Test void testPrettyNestedRoundtrip() {
        Employee e = new Employee("Alice", new Dept("Engineering"));
        String pretty = Ason.encodePretty(e);
        Employee decoded = Ason.decode(pretty, Employee.class);
        assertEquals(e, decoded);
    }

    @Test void testPrettyComplexSliceRoundtrip() {
        List<Score> scores = List.of(new Score(1, 95.5), new Score(2, 72.3), new Score(3, 40.0));
        String pretty = Ason.encodePretty(scores);
        List<Score> decoded = Ason.decodeList(pretty, Score.class);
        assertEquals(scores, decoded);
    }

    @Test void testPrettyLargeSliceRoundtrip() {
        List<Row> rows = new ArrayList<>();
        for (int i = 1; i <= 50; i++) rows.add(new Row(i, "user" + i));
        String pretty = Ason.encodePretty(rows);
        List<Row> decoded = Ason.decodeList(pretty, Row.class);
        assertEquals(50, decoded.size());
        assertEquals(rows.get(49), decoded.get(49));
    }

    // ========================================================================
    // Binary roundtrip tests
    // ========================================================================

    @Test void testBinaryRoundtrip() {
        User u = new User(42, "Alice", true);
        byte[] bin = Ason.encodeBinary(u);
        User decoded = Ason.decodeBinary(bin, User.class);
        assertEquals(u, decoded);
    }

    @Test void testBinaryListRoundtrip() {
        List<Row> rows = List.of(new Row(1, "Alice"), new Row(2, "Bob"), new Row(3, "Carol"));
        byte[] bin = Ason.encodeBinary(rows);
        List<Row> decoded = Ason.decodeBinaryList(bin, Row.class);
        assertEquals(rows, decoded);
    }

    @Test void testBinaryScoreRoundtrip() {
        Score s = new Score(7, 99.9);
        byte[] bin = Ason.encodeBinary(s);
        Score decoded = Ason.decodeBinary(bin, Score.class);
        assertEquals(s, decoded);
    }

    // ─── Typed primitive list fields ────────────────────────────────────

    public static class WithBoolList {
        public String name;
        public List<Boolean> flags;
        public WithBoolList() { flags = new ArrayList<>(); }
    }

    public static class WithIntList {
        public String name;
        public List<Long> nums;
        public WithIntList() { nums = new ArrayList<>(); }
    }

    public static class WithStrList {
        public String name;
        public List<String> tags;
        public WithStrList() { tags = new ArrayList<>(); }
    }

    @Test void testEncodeTypedBoolListField() {
        WithBoolList v = new WithBoolList();
        v.name = "test";
        v.flags = List.of(true, false, true);
        String out = Ason.encodeTyped(v);
        assertTrue(out.contains("flags:[bool]"), "got: " + out);
    }

    @Test void testEncodeTypedIntListField() {
        WithIntList v = new WithIntList();
        v.name = "test";
        v.nums = List.of(1L, 2L, 3L);
        String out = Ason.encodeTyped(v);
        assertTrue(out.contains("nums:[int]"), "got: " + out);
    }

    @Test void testEncodeTypedStrListField() {
        WithStrList v = new WithStrList();
        v.name = "test";
        v.tags = List.of("a", "b");
        String out = Ason.encodeTyped(v);
        assertTrue(out.contains("tags:[str]"), "got: " + out);
    }

    @Test void testEncodeTypedEmptyBoolList() {
        WithBoolList v = new WithBoolList();
        v.name = "test";
        v.flags = new ArrayList<>();
        String out = Ason.encodeTyped(v);
        // Empty list should still get type annotation from generic reflection
        assertTrue(out.contains("flags:[bool]"), "got: " + out);
    }

    @Test void testEncodePrettyTypedBoolListField() {
        WithBoolList v = new WithBoolList();
        v.name = "test";
        v.flags = List.of(true, false);
        String out = Ason.encodePrettyTyped(v);
        assertTrue(out.contains("[bool]"), "got: " + out);
    }

    @Test void testEncodeEmptyTopLevelList() {
        List<Row> empty = new ArrayList<>();
        String out = Ason.encode(empty);
        assertEquals("[]", out);
    }

    @Test void testEncodeTypedEmptyTopLevelList() {
        List<Row> empty = new ArrayList<>();
        String out = Ason.encodeTyped(empty);
        assertEquals("[]", out);
    }

    // ─── Field names with +/- in decode ─────────────────────────────────

    @Test void testDecodeFieldNameWithPlusMinus() {
        // +/- in field names should not break the parser — unmatched fields are just ignored
        String input = "{lowPriorityEIR+CIR:int,name:str}:(42,Alice)";
        User decoded = Ason.decode(input, User.class);
        // "name" matches, "lowPriorityEIR+CIR" does not match any User field → ignored
        assertEquals("Alice", decoded.name);
    }
}
