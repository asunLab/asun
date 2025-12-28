#include <ason/ason.hpp>
#include <iostream>

int main() {
    std::cout << "=== ASON C++ Library Demo ===" << std::endl;
    std::cout << "Version: " << ason::VERSION << std::endl << std::endl;
    
    // Example 1: Parse a simple object
    std::cout << "1. Parse simple object:" << std::endl;
    auto result = ason::parse("{name,age}:(Alice,30)");
    if (result.ok()) {
        auto& value = result.value();
        std::cout << "   name = " << *value.get("name")->as_string() << std::endl;
        std::cout << "   age = " << *value.get("age")->as_integer() << std::endl;
    } else {
        std::cerr << "   Error: " << result.error().what() << std::endl;
    }
    std::cout << std::endl;
    
    // Example 2: Parse multiple objects
    std::cout << "2. Parse multiple objects:" << std::endl;
    result = ason::parse("{name,age}:(Alice,30),(Bob,25)");
    if (result.ok()) {
        auto& arr = *result.value().as_array();
        for (size_t i = 0; i < arr.size(); ++i) {
            std::cout << "   [" << i << "] name = " << *arr[i].get("name")->as_string()
                      << ", age = " << *arr[i].get("age")->as_integer() << std::endl;
        }
    }
    std::cout << std::endl;
    
    // Example 3: Nested objects
    std::cout << "3. Parse nested object:" << std::endl;
    result = ason::parse("{name,addr{city,zip}}:(Alice,(NYC,10001))");
    if (result.ok()) {
        auto& value = result.value();
        std::cout << "   name = " << *value.get("name")->as_string() << std::endl;
        auto addr = value.get("addr");
        std::cout << "   addr.city = " << *addr->get("city")->as_string() << std::endl;
        std::cout << "   addr.zip = " << *addr->get("zip")->as_integer() << std::endl;
    }
    std::cout << std::endl;
    
    // Example 4: Arrays
    std::cout << "4. Parse array field:" << std::endl;
    result = ason::parse("{name,scores[]}:(Alice,[90,85,95])");
    if (result.ok()) {
        auto& value = result.value();
        std::cout << "   name = " << *value.get("name")->as_string() << std::endl;
        auto& scores = *value.get("scores")->as_array();
        std::cout << "   scores = [";
        for (size_t i = 0; i < scores.size(); ++i) {
            if (i > 0) std::cout << ", ";
            std::cout << *scores[i].as_integer();
        }
        std::cout << "]" << std::endl;
    }
    std::cout << std::endl;
    
    // Example 5: Serialize
    std::cout << "5. Serialize value:" << std::endl;
    ason::Object obj;
    obj["name"] = ason::Value("Bob");
    obj["active"] = ason::Value(true);
    obj["score"] = ason::Value(95.5);
    ason::Value v(std::move(obj));
    std::cout << "   " << ason::to_string(v) << std::endl;
    std::cout << std::endl;
    
    // Example 6: Unicode support
    std::cout << "6. Unicode support:" << std::endl;
    result = ason::parse("{name,city}:(Ming,Beijing)");
    if (result.ok()) {
        auto& value = result.value();
        std::cout << "   name = " << *value.get("name")->as_string() << std::endl;
        std::cout << "   city = " << *value.get("city")->as_string() << std::endl;
    }

    std::cout << std::endl << "=== Demo Complete ===" << std::endl;
    return 0;
}

