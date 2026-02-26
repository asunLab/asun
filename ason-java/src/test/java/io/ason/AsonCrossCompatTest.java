package io.ason;

import org.junit.jupiter.api.Test;
import java.util.*;
import static org.junit.jupiter.api.Assertions.*;

public class AsonCrossCompatTest {

    // ========================================================================
    // Struct definitions
    // ========================================================================

    // Dimension 1: Trailing fields
    public static class FullUser {
        public long id;
        public String name;
        public int age;
        public boolean active;
        public double score;
        public FullUser() {}
        public FullUser(long id, String name, int age, boolean active, double score) {
            this.id = id; this.name = name; this.age = age; this.active = active; this.score = score;
        }
    }

    public static class MiniUser {
        public long id;
        public String name;
        public MiniUser() {}
    }

    // Dimension 2: Complex trailing fields
    public static class RichProfile {
        public long id;
        public String name;
        public List<String> tags;
        public List<Long> scores;
        public RichProfile() { tags = new ArrayList<>(); scores = new ArrayList<>(); }
        public RichProfile(long id, String name, List<String> tags, List<Long> scores) {
            this.id = id; this.name = name; this.tags = tags; this.scores = scores;
        }
    }

    public static class ThinProfile {
        public long id;
        public String name;
        public ThinProfile() {}
    }

    // Dimension 3: Nested struct
    public static class InnerFull {
        public long x;
        public long y;
        public double z;
        public boolean w;
        public InnerFull() {}
        public InnerFull(long x, long y, double z, boolean w) {
            this.x = x; this.y = y; this.z = z; this.w = w;
        }
    }

    public static class OuterFull {
        public String name;
        public InnerFull inner;
        public boolean flag;
        public OuterFull() {}
        public OuterFull(String name, InnerFull inner, boolean flag) {
            this.name = name; this.inner = inner; this.flag = flag;
        }
    }

    public static class InnerThin {
        public long x;
        public long y;
        public InnerThin() {}
    }

    public static class OuterThin {
        public String name;
        public InnerThin inner;
        public OuterThin() {}
    }

    // Dimension 4: Vec of nested structs
    public static class TaskFull {
        public String title;
        public boolean done;
        public long priority;
        public double weight;
        public TaskFull() {}
        public TaskFull(String title, boolean done, long priority, double weight) {
            this.title = title; this.done = done; this.priority = priority; this.weight = weight;
        }
    }

    public static class ProjectFull {
        public String name;
        public List<TaskFull> tasks;
        public ProjectFull() { tasks = new ArrayList<>(); }
        public ProjectFull(String name, List<TaskFull> tasks) {
            this.name = name; this.tasks = tasks;
        }
    }

    public static class TaskThin {
        public String title;
        public boolean done;
        public TaskThin() {}
    }

    public static class ProjectThin {
        public String name;
        public List<TaskThin> tasks;
        public ProjectThin() { tasks = new ArrayList<>(); }
    }

    // Dimension 5: Deep nesting
    public static class L3Full {
        public long a;
        public String b;
        public boolean c;
        public L3Full() {}
        public L3Full(long a, String b, boolean c) { this.a = a; this.b = b; this.c = c; }
    }

    public static class L2Full {
        public String name;
        public L3Full sub;
        public long code;
        public List<String> tags;
        public L2Full() { tags = new ArrayList<>(); }
        public L2Full(String name, L3Full sub, long code, List<String> tags) {
            this.name = name; this.sub = sub; this.code = code; this.tags = tags;
        }
    }

    public static class L1Full {
        public long id;
        public L2Full child;
        public String extra;
        public L1Full() {}
        public L1Full(long id, L2Full child, String extra) {
            this.id = id; this.child = child; this.extra = extra;
        }
    }

    public static class L3Thin {
        public long a;
        public L3Thin() {}
    }

    public static class L2Thin {
        public String name;
        public L3Thin sub;
        public L2Thin() {}
    }

    public static class L1Thin {
        public long id;
        public L2Thin child;
        public L1Thin() {}
    }

    // Dimension 6: Field reorder
    public static class OrderABC {
        public long a;
        public String b;
        public boolean c;
        public OrderABC() {}
        public OrderABC(long a, String b, boolean c) { this.a = a; this.b = b; this.c = c; }
    }

    public static class OrderCAB {
        public boolean c;
        public long a;
        public String b;
        public OrderCAB() {}
    }

    // Dimension 7: Reorder + drop
    public static class BigRecord {
        public long id;
        public String name;
        public double score;
        public boolean active;
        public long level;
        public BigRecord() {}
        public BigRecord(long id, String name, double score, boolean active, long level) {
            this.id = id; this.name = name; this.score = score; this.active = active; this.level = level;
        }
    }

    public static class SmallReordered {
        public double score;
        public long id;
        public SmallReordered() {}
    }

    // Dimension 8: Target has extra fields
    public static class SrcSmall {
        public long id;
        public String name;
        public SrcSmall() {}
        public SrcSmall(long id, String name) { this.id = id; this.name = name; }
    }

    public static class DstBig {
        public long id;
        public String name;
        public boolean missing;
        public double extra;
        public DstBig() {}
    }

    // Dimension 9: Optional fields
    public static class SrcWithOptionals {
        public long id;
        public Optional<String> label;
        public Optional<Double> score;
        public boolean flag;
        public SrcWithOptionals() { label = Optional.empty(); score = Optional.empty(); }
        public SrcWithOptionals(long id, Optional<String> label, Optional<Double> score, boolean flag) {
            this.id = id; this.label = label; this.score = score; this.flag = flag;
        }
    }

    public static class DstFewerOptionals {
        public long id;
        public Optional<String> label;
        public DstFewerOptionals() { label = Optional.empty(); }
    }

    // Dimension 10: Special strings trailing
    public static class SrcSpecialStr {
        public long id;
        public String name;
        public String bio;
        public SrcSpecialStr() {}
        public SrcSpecialStr(long id, String name, String bio) {
            this.id = id; this.name = name; this.bio = bio;
        }
    }

    public static class DstNoStr {
        public long id;
        public DstNoStr() {}
    }

    // Dimension 11: Nested arrays trailing
    public static class SrcNestedArray {
        public long id;
        public List<Long> matrix;
        public List<String> tags;
        public SrcNestedArray() { matrix = new ArrayList<>(); tags = new ArrayList<>(); }
        public SrcNestedArray(long id, List<Long> matrix, List<String> tags) {
            this.id = id; this.matrix = matrix; this.tags = tags;
        }
    }

    public static class DstNestedArrayThin {
        public long id;
        public DstNestedArrayThin() {}
    }

    // Dimension 13: Float roundtrip
    public static class SrcFloat {
        public long id;
        public double value;
        public SrcFloat() {}
        public SrcFloat(long id, double value) { this.id = id; this.value = value; }
    }

    public static class DstFloat {
        public long id;
        public double value;
        public DstFloat() {}
    }

    // Dimension 14: Negative numbers
    public static class SrcNegative {
        public long a;
        public long b;
        public double c;
        public String d;
        public SrcNegative() {}
        public SrcNegative(long a, long b, double c, String d) {
            this.a = a; this.b = b; this.c = c; this.d = d;
        }
    }

    public static class DstNegativeThin {
        public long a;
        public long b;
        public DstNegativeThin() {}
    }

    // Dimension 15: Empty strings
    public static class SrcEmpty {
        public long id;
        public String name;
        public String bio;
        public SrcEmpty() {}
        public SrcEmpty(long id, String name, String bio) {
            this.id = id; this.name = name; this.bio = bio;
        }
    }

    public static class DstEmptyThin {
        public long id;
        public DstEmptyThin() {}
    }

    // Dimension 16: Map trailing
    public static class SrcWithMap {
        public long id;
        public String name;
        public Map<String, Long> meta;
        public SrcWithMap() { meta = new LinkedHashMap<>(); }
        public SrcWithMap(long id, String name, Map<String, Long> meta) {
            this.id = id; this.name = name; this.meta = meta;
        }
    }

    public static class DstNoMap {
        public long id;
        public String name;
        public DstNoMap() {}
    }

    // Dimension 19: Nested vec struct + trailing outer fields
    public static class DetailFull {
        public long id;
        public String name;
        public int age;
        public boolean gender;
        public DetailFull() {}
        public DetailFull(long id, String name, int age, boolean gender) {
            this.id = id; this.name = name; this.age = age; this.gender = gender;
        }
    }

    public static class UserFull {
        public List<DetailFull> details;
        public long code;
        public String label;
        public UserFull() { details = new ArrayList<>(); }
        public UserFull(List<DetailFull> details, long code, String label) {
            this.details = details; this.code = code; this.label = label;
        }
    }

    public static class PersonThin {
        public long id;
        public String name;
        public PersonThin() {}
    }

    public static class HumanThin {
        public List<PersonThin> details;
        public HumanThin() { details = new ArrayList<>(); }
    }

    // Dimension 20: All-bool trailing
    public static class SrcBools {
        public long id;
        public boolean a;
        public boolean b;
        public boolean c;
        public SrcBools() {}
        public SrcBools(long id, boolean a, boolean b, boolean c) {
            this.id = id; this.a = a; this.b = b; this.c = c;
        }
    }

    public static class DstBoolsThin {
        public long id;
        public DstBoolsThin() {}
    }

    // Dimension 21/22: Pick middle/last
    public static class SrcFiveFields {
        public long a;
        public String b;
        public double c;
        public boolean d;
        public long e;
        public SrcFiveFields() {}
        public SrcFiveFields(long a, String b, double c, boolean d, long e) {
            this.a = a; this.b = b; this.c = c; this.d = d; this.e = e;
        }
    }

    public static class DstMiddleOnly {
        public double c;
        public DstMiddleOnly() {}
    }

    public static class DstLastOnly {
        public long e;
        public DstLastOnly() {}
    }

    // Dimension 23: No overlap
    public static class SrcAlpha {
        public long x;
        public String y;
        public SrcAlpha() {}
        public SrcAlpha(long x, String y) { this.x = x; this.y = y; }
    }

    public static class DstBeta {
        public long p;
        public String q;
        public DstBeta() {}
    }

    // Dimension 24: Nested array of structs
    public static class WorkerFull {
        public String name;
        public List<String> skills;
        public long yearsXp;
        public double rating;
        public WorkerFull() { skills = new ArrayList<>(); }
        public WorkerFull(String name, List<String> skills, long yearsXp, double rating) {
            this.name = name; this.skills = skills; this.yearsXp = yearsXp; this.rating = rating;
        }
    }

    public static class TeamFull {
        public String lead;
        public List<WorkerFull> workers;
        public double budget;
        public TeamFull() { workers = new ArrayList<>(); }
        public TeamFull(String lead, List<WorkerFull> workers, double budget) {
            this.lead = lead; this.workers = workers; this.budget = budget;
        }
    }

    public static class WorkerThin {
        public String name;
        public List<String> skills;
        public WorkerThin() { skills = new ArrayList<>(); }
    }

    public static class TeamThin {
        public String lead;
        public List<WorkerThin> workers;
        public TeamThin() { workers = new ArrayList<>(); }
    }

    // Dimension 25: Typed mixed
    public static class SrcTyped {
        public long a;
        public String b;
        public double c;
        public boolean d;
        public SrcTyped() {}
        public SrcTyped(long a, String b, double c, boolean d) {
            this.a = a; this.b = b; this.c = c; this.d = d;
        }
    }

    public static class DstMixed {
        public String b;
        public boolean d;
        public long extra;
        public double more;
        public DstMixed() {}
    }

    // Dimension 26: Many trailing fields
    public static class SrcWide {
        public long f1;
        public String f2;
        public boolean f3;
        public long f4;
        public String f5;
        public boolean f6;
        public long f7;
        public String f8;
        public boolean f9;
        public long f10;
        public SrcWide() {}
    }

    public static class DstNarrow {
        public long f1;
        public DstNarrow() {}
    }

    // Dimension 28: ASON-like strings
    public static class SrcAsonLike {
        public long id;
        public String data;
        public String code;
        public SrcAsonLike() {}
        public SrcAsonLike(long id, String data, String code) {
            this.id = id; this.data = data; this.code = code;
        }
    }

    public static class DstAsonLikeThin {
        public long id;
        public DstAsonLikeThin() {}
    }

    // Dimension 29: Unicode
    public static class SrcUnicode {
        public long id;
        public String name;
        public String bio;
        public SrcUnicode() {}
        public SrcUnicode(long id, String name, String bio) {
            this.id = id; this.name = name; this.bio = bio;
        }
    }

    public static class DstUnicodeThin {
        public long id;
        public DstUnicodeThin() {}
    }

    // Dimension 30: Roundtrip ABBA
    public static class VersionA {
        public long id;
        public String name;
        public boolean active;
        public VersionA() {}
        public VersionA(long id, String name, boolean active) {
            this.id = id; this.name = name; this.active = active;
        }
    }

    public static class VersionB {
        public long id;
        public String name;
        public VersionB() {}
    }

    // Dimension 31: Empty arrays in middle
    public static class SrcWithArr {
        public long id;
        public List<String> items;
        public long score;
        public SrcWithArr() { items = new ArrayList<>(); }
        public SrcWithArr(long id, List<String> items, long score) {
            this.id = id; this.items = items; this.score = score;
        }
    }

    public static class DstWithArrThin {
        public long id;
        public List<String> items;
        public DstWithArrThin() { items = new ArrayList<>(); }
    }

    // Dimension 32: Skip nested struct as tuple
    public static class SrcInner {
        public long a;
        public String b;
        public SrcInner() {}
        public SrcInner(long a, String b) { this.a = a; this.b = b; }
    }

    public static class SrcWithNested {
        public long id;
        public SrcInner inner;
        public String tail;
        public SrcWithNested() {}
        public SrcWithNested(long id, SrcInner inner, String tail) {
            this.id = id; this.inner = inner; this.tail = tail;
        }
    }

    public static class DstFlat {
        public long id;
        public DstFlat() {}
    }

    // ========================================================================
    // Tests
    // ========================================================================

    // Dimension 1: Extra trailing fields — vec
    @Test void testTrailingFieldsDroppedVec() {
        List<FullUser> src = List.of(
            new FullUser(1, "Alice", 30, true, 95.5),
            new FullUser(2, "Bob", 25, false, 87.0)
        );
        String data = Ason.encode(src);
        List<MiniUser> dst = Ason.decodeList(data, MiniUser.class);
        assertEquals(2, dst.size());
        assertEquals(1, dst.get(0).id);
        assertEquals("Alice", dst.get(0).name);
        assertEquals(2, dst.get(1).id);
        assertEquals("Bob", dst.get(1).name);
    }

    // Dimension 1: Extra trailing fields — single struct
    @Test void testTrailingFieldsDroppedSingle() {
        FullUser src = new FullUser(99, "Zara", 40, true, 100.0);
        String data = Ason.encode(src);
        MiniUser dst = Ason.decode(data, MiniUser.class);
        assertEquals(99, dst.id);
        assertEquals("Zara", dst.name);
    }

    // Dimension 2: Skip trailing complex types
    @Test void testSkipTrailingArrayAndMap() {
        RichProfile src = new RichProfile(1, "Alice", List.of("go", "rust"), List.of(90L, 85L, 92L));
        String data = Ason.encode(src);
        ThinProfile dst = Ason.decode(data, ThinProfile.class);
        assertEquals(1, dst.id);
        assertEquals("Alice", dst.name);
    }

    // Dimension 3: Nested struct fewer fields
    @Test void testNestedStructFewerFields() {
        OuterFull src = new OuterFull("test", new InnerFull(10, 20, 3.14, true), true);
        String data = Ason.encode(src);
        OuterThin dst = Ason.decode(data, OuterThin.class);
        assertEquals("test", dst.name);
        assertEquals(10, dst.inner.x);
        assertEquals(20, dst.inner.y);
    }

    // Dimension 4: Vec of nested structs
    @Test void testVecNestedStructSkipExtra() {
        List<ProjectFull> src = List.of(
            new ProjectFull("Alpha", List.of(
                new TaskFull("Design", true, 1, 0.5),
                new TaskFull("Code", false, 2, 0.8)
            )),
            new ProjectFull("Beta", List.of(
                new TaskFull("Test", false, 3, 1.0)
            ))
        );
        String data = Ason.encode(src);
        List<ProjectThin> dst = Ason.decodeList(data, ProjectThin.class);
        assertEquals(2, dst.size());
        assertEquals("Alpha", dst.get(0).name);
        assertEquals(2, dst.get(0).tasks.size());
        assertEquals("Design", dst.get(0).tasks.get(0).title);
        assertTrue(dst.get(0).tasks.get(0).done);
        assertEquals("Code", dst.get(0).tasks.get(1).title);
        assertFalse(dst.get(0).tasks.get(1).done);
        assertEquals("Beta", dst.get(1).name);
        assertEquals(1, dst.get(1).tasks.size());
    }

    // Dimension 5: Deep nesting 3 levels
    @Test void testDeepNesting3Levels() {
        L1Full src = new L1Full(1,
            new L2Full("mid", new L3Full(42, "hello", true), 7, List.of("x", "y")),
            "dropped");
        String data = Ason.encode(src);
        L1Thin dst = Ason.decode(data, L1Thin.class);
        assertEquals(1, dst.id);
        assertEquals("mid", dst.child.name);
        assertEquals(42, dst.child.sub.a);
    }

    // Dimension 6: Field reorder
    @Test void testFieldReorder() {
        OrderABC src = new OrderABC(1, "hi", true);
        String data = Ason.encode(src);
        OrderCAB dst = Ason.decode(data, OrderCAB.class);
        assertEquals(1, dst.a);
        assertEquals("hi", dst.b);
        assertTrue(dst.c);
    }

    // Dimension 7: Reorder + drop trailing
    @Test void testReorderPlusDropTrailing() {
        List<BigRecord> src = List.of(
            new BigRecord(1, "A", 9.5, true, 3),
            new BigRecord(2, "B", 8.0, false, 1)
        );
        String data = Ason.encode(src);
        List<SmallReordered> dst = Ason.decodeList(data, SmallReordered.class);
        assertEquals(2, dst.size());
        assertEquals(1, dst.get(0).id);
        assertEquals(9.5, dst.get(0).score);
        assertEquals(2, dst.get(1).id);
        assertEquals(8.0, dst.get(1).score);
    }

    // Dimension 8: Target has extra fields
    @Test void testTargetHasExtraFields() {
        SrcSmall src = new SrcSmall(42, "Alice");
        String data = Ason.encode(src);
        DstBig dst = Ason.decode(data, DstBig.class);
        assertEquals(42, dst.id);
        assertEquals("Alice", dst.name);
        assertFalse(dst.missing);
        assertEquals(0.0, dst.extra);
    }

    // Dimension 9: Optional fields skip trailing
    @Test void testOptionalFieldsSkipTrailing() {
        SrcWithOptionals src = new SrcWithOptionals(1, Optional.of("hello"), Optional.of(95.5), true);
        String data = Ason.encode(src);
        DstFewerOptionals dst = Ason.decode(data, DstFewerOptionals.class);
        assertEquals(1, dst.id);
        assertEquals(Optional.of("hello"), dst.label);
    }

    // Dimension 9: Optional nil skip trailing
    @Test void testOptionalNilSkipTrailing() {
        SrcWithOptionals src = new SrcWithOptionals(2, Optional.empty(), Optional.empty(), false);
        String data = Ason.encode(src);
        DstFewerOptionals dst = Ason.decode(data, DstFewerOptionals.class);
        assertEquals(2, dst.id);
        assertEquals(Optional.empty(), dst.label);
    }

    // Dimension 10: Skip quoted string with special chars
    @Test void testSkipQuotedStringWithSpecialChars() {
        SrcSpecialStr src = new SrcSpecialStr(1, "comma,here", "paren(test) and \"quotes\"");
        String data = Ason.encode(src);
        DstNoStr dst = Ason.decode(data, DstNoStr.class);
        assertEquals(1, dst.id);
    }

    // Dimension 11: Skip trailing array fields
    @Test void testSkipTrailingArrayFields() {
        List<SrcNestedArray> src = List.of(
            new SrcNestedArray(1, List.of(1L, 2L, 3L), List.of("a", "b")),
            new SrcNestedArray(2, List.of(4L, 5L), List.of("c"))
        );
        String data = Ason.encode(src);
        List<DstNestedArrayThin> dst = Ason.decodeList(data, DstNestedArrayThin.class);
        assertEquals(2, dst.size());
        assertEquals(1, dst.get(0).id);
        assertEquals(2, dst.get(1).id);
    }

    // Dimension 12: Int widening (Java long covers all)
    @Test void testIntWidening() {
        SrcSmall src = new SrcSmall(100, "wide");
        String data = Ason.encode(src);
        SrcSmall dst = Ason.decode(data, SrcSmall.class);
        assertEquals(100, dst.id);
        assertEquals("wide", dst.name);
    }

    // Dimension 13: Float precision roundtrip
    @Test void testFloatRoundtrip() {
        SrcFloat src = new SrcFloat(1, 3.14159);
        String data = Ason.encode(src);
        DstFloat dst = Ason.decode(data, DstFloat.class);
        assertTrue(Math.abs(dst.value - 3.14159) < 1e-10);
    }

    // Dimension 14: Negative numbers skip trailing
    @Test void testNegativeNumbersSkipTrailing() {
        SrcNegative src = new SrcNegative(-1, -999999, -3.14, "neg");
        String data = Ason.encode(src);
        DstNegativeThin dst = Ason.decode(data, DstNegativeThin.class);
        assertEquals(-1, dst.a);
        assertEquals(-999999, dst.b);
    }

    // Dimension 15: Empty string fields
    @Test void testEmptyStringFields() {
        SrcEmpty src = new SrcEmpty(1, "", "");
        String data = Ason.encode(src);
        DstEmptyThin dst = Ason.decode(data, DstEmptyThin.class);
        assertEquals(1, dst.id);
    }

    // Dimension 16: Skip trailing map field
    @Test void testSkipTrailingMapField() {
        Map<String, Long> meta = new LinkedHashMap<>();
        meta.put("age", 30L);
        meta.put("score", 95L);
        SrcWithMap src = new SrcWithMap(1, "Alice", meta);
        String data = Ason.encode(src);
        DstNoMap dst = Ason.decode(data, DstNoMap.class);
        assertEquals(1, dst.id);
        assertEquals("Alice", dst.name);
    }

    // Dimension 17: Typed schema vec decode
    @Test void testTypedSchemaVecDecode() {
        List<FullUser> src = List.of(
            new FullUser(1, "Alice", 30, true, 95.5)
        );
        String data = Ason.encodeTyped(src);
        List<MiniUser> dst = Ason.decodeList(data, MiniUser.class);
        assertEquals(1, dst.size());
        assertEquals(1, dst.get(0).id);
        assertEquals("Alice", dst.get(0).name);
    }

    // Dimension 18: Typed schema single decode
    @Test void testTypedSchemaSingleDecode() {
        FullUser src = new FullUser(42, "Bob", 25, false, 88.0);
        String data = Ason.encodeTyped(src);
        MiniUser dst = Ason.decode(data, MiniUser.class);
        assertEquals(42, dst.id);
        assertEquals("Bob", dst.name);
    }

    // Dimension 19: Nested vec struct + trailing outer fields
    @Test void testNestedVecStructPlusTrailingOuterFields() {
        List<UserFull> src = List.of(
            new UserFull(List.of(
                new DetailFull(1, "Alice", 30, true),
                new DetailFull(2, "Bob", 25, false)
            ), 42, "test")
        );
        String data = Ason.encode(src);
        List<HumanThin> dst = Ason.decodeList(data, HumanThin.class);
        assertEquals(1, dst.size());
        assertEquals(2, dst.get(0).details.size());
        assertEquals(1, dst.get(0).details.get(0).id);
        assertEquals("Alice", dst.get(0).details.get(0).name);
        assertEquals(2, dst.get(0).details.get(1).id);
        assertEquals("Bob", dst.get(0).details.get(1).name);
    }

    // Dimension 20: Skip trailing bools
    @Test void testSkipTrailingBools() {
        List<SrcBools> src = List.of(
            new SrcBools(1, true, false, true),
            new SrcBools(2, false, true, false)
        );
        String data = Ason.encode(src);
        List<DstBoolsThin> dst = Ason.decodeList(data, DstBoolsThin.class);
        assertEquals(2, dst.size());
        assertEquals(1, dst.get(0).id);
        assertEquals(2, dst.get(1).id);
    }

    // Dimension 21: Pick middle field only
    @Test void testPickMiddleFieldOnly() {
        SrcFiveFields src = new SrcFiveFields(1, "hi", 3.14, true, 99);
        String data = Ason.encode(src);
        DstMiddleOnly dst = Ason.decode(data, DstMiddleOnly.class);
        assertEquals(3.14, dst.c);
    }

    // Dimension 22: Pick last field only
    @Test void testPickLastFieldOnly() {
        SrcFiveFields src = new SrcFiveFields(1, "hi", 3.14, true, 42);
        String data = Ason.encode(src);
        DstLastOnly dst = Ason.decode(data, DstLastOnly.class);
        assertEquals(42, dst.e);
    }

    // Dimension 23: No overlapping fields
    @Test void testNoOverlappingFields() {
        SrcAlpha src = new SrcAlpha(1, "hello");
        String data = Ason.encode(src);
        DstBeta dst = Ason.decode(data, DstBeta.class);
        assertEquals(0, dst.p);
        assertNull(dst.q);
    }

    // Dimension 24: Nested array of structs with extra fields
    @Test void testNestedArrayOfStructsWithExtraFields() {
        TeamFull src = new TeamFull("Alice", List.of(
            new WorkerFull("Bob", List.of("go", "rust"), 5, 4.5),
            new WorkerFull("Carol", List.of("python"), 3, 3.8)
        ), 100000.0);
        String data = Ason.encode(src);
        TeamThin dst = Ason.decode(data, TeamThin.class);
        assertEquals("Alice", dst.lead);
        assertEquals(2, dst.workers.size());
        assertEquals("Bob", dst.workers.get(0).name);
        assertEquals(List.of("go", "rust"), dst.workers.get(0).skills);
        assertEquals("Carol", dst.workers.get(1).name);
        assertEquals(List.of("python"), dst.workers.get(1).skills);
    }

    // Dimension 25: Typed schema mixed fields (subset + reorder + extra target)
    @Test void testTypedSchemaMixedFields() {
        SrcTyped src = new SrcTyped(1, "test", 2.5, true);
        String data = Ason.encodeTyped(src);
        DstMixed dst = Ason.decode(data, DstMixed.class);
        assertEquals("test", dst.b);
        assertTrue(dst.d);
        assertEquals(0, dst.extra);
        assertEquals(0.0, dst.more);
    }

    // Dimension 26: Many trailing fields (stress skip)
    @Test void testManyTrailingFields() {
        SrcWide src = new SrcWide();
        src.f1 = 42; src.f2 = "a"; src.f3 = true; src.f4 = 4; src.f5 = "b";
        src.f6 = false; src.f7 = 7; src.f8 = "c"; src.f9 = true; src.f10 = 10;
        String data = Ason.encode(src);
        DstNarrow dst = Ason.decode(data, DstNarrow.class);
        assertEquals(42, dst.f1);
    }

    // Dimension 27: Vec with single row
    @Test void testVecSingleRow() {
        List<FullUser> src = List.of(new FullUser(1, "Alice", 30, true, 95.5));
        String data = Ason.encode(src);
        List<MiniUser> dst = Ason.decodeList(data, MiniUser.class);
        assertEquals(1, dst.size());
        assertEquals(1, dst.get(0).id);
        assertEquals("Alice", dst.get(0).name);
    }

    // Dimension 28: Skip string containing ASON-like syntax
    @Test void testSkipStringContainingAsonSyntax() {
        SrcAsonLike src = new SrcAsonLike(1, "{a,b}:(1,2)", "[(x,y),(z,w)]");
        String data = Ason.encode(src);
        DstAsonLikeThin dst = Ason.decode(data, DstAsonLikeThin.class);
        assertEquals(1, dst.id);
    }

    // Dimension 29: Unicode in trailing fields
    @Test void testSkipUnicodeInTrailing() {
        SrcUnicode src = new SrcUnicode(1, "日本語テスト", "中文描述，包含逗号");
        String data = Ason.encode(src);
        DstUnicodeThin dst = Ason.decode(data, DstUnicodeThin.class);
        assertEquals(1, dst.id);
    }

    // Dimension 30: Roundtrip ABBA
    @Test void testRoundtripABBA() {
        // A -> B
        VersionA srcA = new VersionA(1, "test", true);
        String dataA = Ason.encode(srcA);
        VersionB dstB = Ason.decode(dataA, VersionB.class);
        assertEquals(1, dstB.id);
        assertEquals("test", dstB.name);

        // B -> A (missing fields = zero)
        String dataB = Ason.encode(dstB);
        VersionA dstA = Ason.decode(dataB, VersionA.class);
        assertEquals(1, dstA.id);
        assertEquals("test", dstA.name);
        assertFalse(dstA.active);
    }

    // Dimension 31: Empty arrays in middle field
    @Test void testEmptyArrayInMiddleField() {
        List<SrcWithArr> src = List.of(
            new SrcWithArr(1, List.of(), 10),
            new SrcWithArr(2, List.of("a", "b"), 20),
            new SrcWithArr(3, new ArrayList<>(), 30)
        );
        String data = Ason.encode(src);
        List<DstWithArrThin> dst = Ason.decodeList(data, DstWithArrThin.class);
        assertEquals(3, dst.size());
        assertEquals(1, dst.get(0).id);
        assertEquals(0, dst.get(0).items.size());
        assertEquals(2, dst.get(1).id);
        assertEquals(List.of("a", "b"), dst.get(1).items);
        assertEquals(3, dst.get(2).id);
    }

    // Dimension 32: Skip nested struct as tuple
    @Test void testSkipNestedStructAsTuple() {
        SrcWithNested src = new SrcWithNested(1, new SrcInner(10, "nested"), "end");
        String data = Ason.encode(src);
        DstFlat dst = Ason.decode(data, DstFlat.class);
        assertEquals(1, dst.id);
    }

    // Dimension 33: Vec with many rows — stress
    @Test void testManyRows() {
        List<FullUser> src = new ArrayList<>();
        for (int i = 0; i < 100; i++) {
            src.add(new FullUser(i, "user", i, i % 2 == 0, i * 0.1));
        }
        String data = Ason.encode(src);
        List<MiniUser> dst = Ason.decodeList(data, MiniUser.class);
        assertEquals(100, dst.size());
        for (int i = 0; i < 100; i++) {
            assertEquals(i, dst.get(i).id);
            assertEquals("user", dst.get(i).name);
        }
    }

    // Dimension 34: Typed encode, target has subset + reorder
    @Test void testTypedEncodeSubsetReorder() {
        List<BigRecord> src = List.of(
            new BigRecord(1, "A", 9.5, true, 3),
            new BigRecord(2, "B", 8.0, false, 1)
        );
        String data = Ason.encodeTyped(src);
        List<SmallReordered> dst = Ason.decodeList(data, SmallReordered.class);
        assertEquals(2, dst.size());
        assertEquals(9.5, dst.get(0).score);
        assertEquals(1, dst.get(0).id);
    }

    // Dimension 35: Zero-value source fields
    @Test void testZeroValueSourceFields() {
        FullUser src = new FullUser(0, "", 0, false, 0.0);
        String data = Ason.encode(src);
        MiniUser dst = Ason.decode(data, MiniUser.class);
        assertEquals(0, dst.id);
        assertEquals("", dst.name);
    }
}
