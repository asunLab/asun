#include <gtest/gtest.h>
#include <ason/serializer.hpp>

using namespace ason;

TEST(SerializerTest, Primitives) {
    EXPECT_EQ(to_string(Value(true)), "true");
    EXPECT_EQ(to_string(Value(false)), "false");
    EXPECT_EQ(to_string(Value(42)), "42");
}

TEST(SerializerTest, StringNoQuotes) {
    EXPECT_EQ(to_string(Value("hello")), "hello");
    EXPECT_EQ(to_string(Value("Alice")), "Alice");
}

TEST(SerializerTest, StringWithQuotes) {
    // Empty string needs quotes
    EXPECT_EQ(to_string(Value("")), R"("")");
    
    // String that looks like boolean
    EXPECT_EQ(to_string(Value("true")), R"("true")");
    
    // String that looks like number
    EXPECT_EQ(to_string(Value("123")), R"("123")");
}

TEST(SerializerTest, Array) {
    Array arr;
    arr.push_back(Value(1));
    arr.push_back(Value(2));
    arr.push_back(Value(3));
    
    EXPECT_EQ(to_string(Value(std::move(arr))), "[1,2,3]");
}

TEST(SerializerTest, MixedArray) {
    Array arr;
    arr.push_back(Value(1));
    arr.push_back(Value("hello"));
    arr.push_back(Value(true));
    
    EXPECT_EQ(to_string(Value(std::move(arr))), "[1,hello,true]");
}

TEST(SerializerTest, Object) {
    Object obj;
    obj["name"] = Value("Alice");
    obj["age"] = Value(30);

    // Note: std::map orders by key, so "age" comes before "name"
    EXPECT_EQ(to_string(Value(std::move(obj))), "(30,Alice)");
}

TEST(SerializerTest, NestedObject) {
    Object inner;
    inner["city"] = Value("NYC");
    inner["zip"] = Value(10001);

    Object outer;
    outer["name"] = Value("Alice");
    outer["addr"] = Value(std::move(inner));

    // Note: std::map orders by key, so "addr" comes before "name", and "city" before "zip"
    EXPECT_EQ(to_string(Value(std::move(outer))), "((NYC,10001),Alice)");
}

TEST(SerializerTest, NullInObject) {
    Object obj;
    obj["name"] = Value("Alice");
    obj["age"] = Value();

    // Note: std::map orders by key, so "age" comes before "name"
    EXPECT_EQ(to_string(Value(std::move(obj))), "(,Alice)");
}

TEST(SerializerTest, NeedsQuotes) {
    // Should NOT need quotes
    EXPECT_FALSE(needs_quotes("hello"));
    EXPECT_FALSE(needs_quotes("Alice"));
    EXPECT_FALSE(needs_quotes("hello-world"));
    EXPECT_FALSE(needs_quotes("foo_bar"));
    
    // Should need quotes
    EXPECT_TRUE(needs_quotes(""));
    EXPECT_TRUE(needs_quotes("true"));
    EXPECT_TRUE(needs_quotes("false"));
    EXPECT_TRUE(needs_quotes("123"));
    EXPECT_TRUE(needs_quotes("3.14"));
    EXPECT_TRUE(needs_quotes("  hello"));
    EXPECT_TRUE(needs_quotes("hello  "));
}

TEST(SerializerTest, LooksLikeNumber) {
    EXPECT_TRUE(looks_like_number("123"));
    EXPECT_TRUE(looks_like_number("-123"));
    EXPECT_TRUE(looks_like_number("3.14"));
    EXPECT_TRUE(looks_like_number("-3.14"));
    EXPECT_TRUE(looks_like_number("1e10"));
    EXPECT_TRUE(looks_like_number("1.5e-3"));
    
    EXPECT_FALSE(looks_like_number(""));
    EXPECT_FALSE(looks_like_number("abc"));
    EXPECT_FALSE(looks_like_number("123abc"));
    EXPECT_FALSE(looks_like_number("-"));
    EXPECT_FALSE(looks_like_number("1."));
}

TEST(SerializerTest, PrettyArrayOfObjects) {
    Object obj1;
    obj1["name"] = Value("Alice");
    obj1["age"] = Value(30);
    
    Object obj2;
    obj2["name"] = Value("Bob");
    obj2["age"] = Value(25);
    
    Array arr;
    arr.push_back(Value(std::move(obj1)));
    arr.push_back(Value(std::move(obj2)));
    
    auto pretty = to_string_pretty(Value(std::move(arr)));
    EXPECT_NE(pretty.find('\n'), std::string::npos);
    EXPECT_NE(pretty.find("  "), std::string::npos); // indentation
    EXPECT_EQ(pretty.front(), '[');
    EXPECT_EQ(pretty.back(), ']');
}

TEST(SerializerTest, PrettySimpleStaysCompact) {
    Array arr;
    arr.push_back(Value(1));
    arr.push_back(Value(2));
    arr.push_back(Value(3));
    
    auto pretty = to_string_pretty(Value(std::move(arr)));
    EXPECT_EQ(pretty, "[1,2,3]");
}

TEST(SerializerTest, PrettySimpleObjectStaysCompact) {
    Object obj;
    obj["name"] = Value("Alice");
    obj["age"] = Value(30);

    auto pretty = to_string_pretty(Value(std::move(obj)));
    // Note: std::map orders by key, so "age" comes before "name"
    EXPECT_EQ(pretty, "(30,Alice)");
}

