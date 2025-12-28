#include <gtest/gtest.h>
#include <ason/value.hpp>

using namespace ason;

TEST(ValueTest, NullValue) {
    Value v;
    EXPECT_TRUE(v.is_null());
    EXPECT_FALSE(v.is_bool());
    EXPECT_FALSE(v.is_integer());
}

TEST(ValueTest, BoolValue) {
    Value v(true);
    EXPECT_TRUE(v.is_bool());
    EXPECT_EQ(*v.as_bool(), true);
    
    Value v2(false);
    EXPECT_EQ(*v2.as_bool(), false);
}

TEST(ValueTest, IntegerValue) {
    Value v(42);
    EXPECT_TRUE(v.is_integer());
    EXPECT_EQ(*v.as_integer(), 42);
    
    Value v2(-100);
    EXPECT_EQ(*v2.as_integer(), -100);
}

TEST(ValueTest, FloatValue) {
    Value v(3.14);
    EXPECT_TRUE(v.is_float());
    EXPECT_DOUBLE_EQ(*v.as_float(), 3.14);
}

TEST(ValueTest, StringValue) {
    Value v("hello");
    EXPECT_TRUE(v.is_string());
    EXPECT_EQ(*v.as_string(), "hello");
    
    Value v2(std::string("world"));
    EXPECT_EQ(*v2.as_string(), "world");
}

TEST(ValueTest, ArrayValue) {
    Array arr;
    arr.push_back(Value(1));
    arr.push_back(Value(2));
    arr.push_back(Value(3));
    
    Value v(std::move(arr));
    EXPECT_TRUE(v.is_array());
    
    auto* a = v.as_array();
    ASSERT_NE(a, nullptr);
    EXPECT_EQ(a->size(), 3);
    EXPECT_EQ(*(*a)[0].as_integer(), 1);
    EXPECT_EQ(*(*a)[1].as_integer(), 2);
    EXPECT_EQ(*(*a)[2].as_integer(), 3);
}

TEST(ValueTest, ObjectValue) {
    Object obj;
    obj["name"] = Value("Alice");
    obj["age"] = Value(30);
    
    Value v(std::move(obj));
    EXPECT_TRUE(v.is_object());
    
    auto* o = v.as_object();
    ASSERT_NE(o, nullptr);
    EXPECT_EQ(o->size(), 2);
    
    auto name = v.get("name");
    ASSERT_NE(name, nullptr);
    EXPECT_EQ(*name->as_string(), "Alice");
    
    auto age = v.get("age");
    ASSERT_NE(age, nullptr);
    EXPECT_EQ(*age->as_integer(), 30);
}

TEST(ValueTest, GetNonExistent) {
    Object obj;
    obj["name"] = Value("Alice");
    Value v(std::move(obj));
    
    EXPECT_EQ(v.get("unknown"), nullptr);
}

TEST(ValueTest, GetIndex) {
    Array arr;
    arr.push_back(Value(10));
    arr.push_back(Value(20));
    Value v(std::move(arr));
    
    auto* first = v.get_index(0);
    ASSERT_NE(first, nullptr);
    EXPECT_EQ(*first->as_integer(), 10);
    
    auto* second = v.get_index(1);
    ASSERT_NE(second, nullptr);
    EXPECT_EQ(*second->as_integer(), 20);
    
    EXPECT_EQ(v.get_index(10), nullptr);
}

TEST(ValueTest, CopyValue) {
    Value v1("hello");
    Value v2 = v1;
    
    EXPECT_EQ(*v1.as_string(), "hello");
    EXPECT_EQ(*v2.as_string(), "hello");
}

TEST(ValueTest, MoveValue) {
    Value v1("hello");
    Value v2 = std::move(v1);
    
    EXPECT_EQ(*v2.as_string(), "hello");
}

TEST(ValueTest, NestedStructure) {
    Object inner;
    inner["city"] = Value("NYC");
    inner["zip"] = Value(10001);
    
    Object outer;
    outer["name"] = Value("Alice");
    outer["addr"] = Value(std::move(inner));
    
    Value v(std::move(outer));
    
    auto addr = v.get("addr");
    ASSERT_NE(addr, nullptr);
    EXPECT_TRUE(addr->is_object());
    
    auto city = addr->get("city");
    ASSERT_NE(city, nullptr);
    EXPECT_EQ(*city->as_string(), "NYC");
}

