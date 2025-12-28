#include <gtest/gtest.h>
#include <ason/parser.hpp>

using namespace ason;

TEST(ParserTest, SimpleObject) {
    auto result = parse("{name,age}:(Alice,30)");
    ASSERT_TRUE(result.ok());
    
    auto& value = result.value();
    EXPECT_TRUE(value.is_object());
    
    auto name = value.get("name");
    ASSERT_NE(name, nullptr);
    EXPECT_EQ(*name->as_string(), "Alice");
    
    auto age = value.get("age");
    ASSERT_NE(age, nullptr);
    EXPECT_EQ(*age->as_integer(), 30);
}

TEST(ParserTest, MultipleObjects) {
    auto result = parse("{name,age}:(Alice,30),(Bob,25)");
    ASSERT_TRUE(result.ok());
    
    auto& value = result.value();
    EXPECT_TRUE(value.is_array());
    
    auto* arr = value.as_array();
    ASSERT_EQ(arr->size(), 2);
    
    EXPECT_EQ(*(*arr)[0].get("name")->as_string(), "Alice");
    EXPECT_EQ(*(*arr)[0].get("age")->as_integer(), 30);
    EXPECT_EQ(*(*arr)[1].get("name")->as_string(), "Bob");
    EXPECT_EQ(*(*arr)[1].get("age")->as_integer(), 25);
}

TEST(ParserTest, NestedObject) {
    auto result = parse("{name,addr{city,zip}}:(Alice,(NYC,10001))");
    ASSERT_TRUE(result.ok());
    
    auto& value = result.value();
    EXPECT_EQ(*value.get("name")->as_string(), "Alice");
    
    auto addr = value.get("addr");
    ASSERT_NE(addr, nullptr);
    EXPECT_EQ(*addr->get("city")->as_string(), "NYC");
    EXPECT_EQ(*addr->get("zip")->as_integer(), 10001);
}

TEST(ParserTest, SimpleArrayField) {
    auto result = parse("{name,scores[]}:(Alice,[90,85,95])");
    ASSERT_TRUE(result.ok());
    
    auto& value = result.value();
    EXPECT_EQ(*value.get("name")->as_string(), "Alice");
    
    auto scores = value.get("scores");
    ASSERT_NE(scores, nullptr);
    EXPECT_TRUE(scores->is_array());
    
    auto* arr = scores->as_array();
    ASSERT_EQ(arr->size(), 3);
    EXPECT_EQ(*(*arr)[0].as_integer(), 90);
    EXPECT_EQ(*(*arr)[1].as_integer(), 85);
    EXPECT_EQ(*(*arr)[2].as_integer(), 95);
}

TEST(ParserTest, ObjectArrayField) {
    auto result = parse("{name,friends[{id,name}]}:(Alice,[(1,Bob),(2,Carol)])");
    ASSERT_TRUE(result.ok());
    
    auto friends = result.value().get("friends");
    ASSERT_NE(friends, nullptr);
    
    auto* arr = friends->as_array();
    ASSERT_EQ(arr->size(), 2);
    
    EXPECT_EQ(*(*arr)[0].get("id")->as_integer(), 1);
    EXPECT_EQ(*(*arr)[0].get("name")->as_string(), "Bob");
    EXPECT_EQ(*(*arr)[1].get("id")->as_integer(), 2);
    EXPECT_EQ(*(*arr)[1].get("name")->as_string(), "Carol");
}

TEST(ParserTest, NullValues) {
    auto result = parse("{name,age,city}:(Alice,,NYC)");
    ASSERT_TRUE(result.ok());
    
    EXPECT_EQ(*result.value().get("name")->as_string(), "Alice");
    EXPECT_TRUE(result.value().get("age")->is_null());
    EXPECT_EQ(*result.value().get("city")->as_string(), "NYC");
}

TEST(ParserTest, EmptyString) {
    auto result = parse(R"({name,bio}:(Alice,""))");
    ASSERT_TRUE(result.ok());
    
    EXPECT_EQ(*result.value().get("bio")->as_string(), "");
}

TEST(ParserTest, Booleans) {
    auto result = parse("{name,active,verified}:(Alice,true,false)");
    ASSERT_TRUE(result.ok());
    
    EXPECT_EQ(*result.value().get("active")->as_bool(), true);
    EXPECT_EQ(*result.value().get("verified")->as_bool(), false);
}

TEST(ParserTest, Floats) {
    auto result = parse("{name,score}:(Alice,98.5)");
    ASSERT_TRUE(result.ok());
    
    EXPECT_DOUBLE_EQ(*result.value().get("score")->as_float(), 98.5);
}

TEST(ParserTest, StandaloneArray) {
    auto result = parse("[1,2,3]");
    ASSERT_TRUE(result.ok());
    
    EXPECT_TRUE(result.value().is_array());
    auto* arr = result.value().as_array();
    ASSERT_EQ(arr->size(), 3);
    EXPECT_EQ(*(*arr)[0].as_integer(), 1);
}

TEST(ParserTest, EmptyArray) {
    auto result = parse("{name,tags[]}:(Alice,[])");
    ASSERT_TRUE(result.ok());
    
    auto tags = result.value().get("tags");
    ASSERT_NE(tags, nullptr);
    EXPECT_TRUE(tags->as_array()->empty());
}

TEST(ParserTest, Comments) {
    auto result = parse("/* header */ {name,age} /* schema */ : /* data */ (Alice,30)");
    ASSERT_TRUE(result.ok());
    
    EXPECT_EQ(*result.value().get("name")->as_string(), "Alice");
    EXPECT_EQ(*result.value().get("age")->as_integer(), 30);
}

TEST(ParserTest, UnicodeStrings) {
    auto result = parse("{name,city}:(小明,北京)");
    ASSERT_TRUE(result.ok());
    
    EXPECT_EQ(*result.value().get("name")->as_string(), "小明");
    EXPECT_EQ(*result.value().get("city")->as_string(), "北京");
}

TEST(ParserTest, NegativeNumbers) {
    auto result = parse("{value}:(-42)");
    ASSERT_TRUE(result.ok());
    
    EXPECT_EQ(*result.value().get("value")->as_integer(), -42);
}

TEST(ParserTest, NegativeFloat) {
    auto result = parse("{value}:(-3.14)");
    ASSERT_TRUE(result.ok());
    
    EXPECT_DOUBLE_EQ(*result.value().get("value")->as_float(), -3.14);
}

