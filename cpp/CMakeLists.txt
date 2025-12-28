cmake_minimum_required(VERSION 3.14)
project(ason VERSION 0.1.0 LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_CXX_EXTENSIONS OFF)

# Options
option(ASON_BUILD_TESTS "Build tests" ON)
option(ASON_BUILD_EXAMPLES "Build examples" ON)

# Library sources
set(ASON_SOURCES
    src/value.cpp
    src/lexer.cpp
    src/parser.cpp
    src/serializer.cpp
)

# Library headers
set(ASON_HEADERS
    include/ason/ason.hpp
    include/ason/error.hpp
    include/ason/value.hpp
    include/ason/lexer.hpp
    include/ason/parser.hpp
    include/ason/serializer.hpp
)

# Create library
add_library(ason ${ASON_SOURCES})
target_include_directories(ason PUBLIC
    $<BUILD_INTERFACE:${CMAKE_CURRENT_SOURCE_DIR}/include>
    $<INSTALL_INTERFACE:include>
)

# Set compile warnings
if(MSVC)
    target_compile_options(ason PRIVATE /W4)
else()
    target_compile_options(ason PRIVATE -Wall -Wextra -Wpedantic)
endif()

# Examples
if(ASON_BUILD_EXAMPLES)
    add_executable(basic_example examples/basic.cpp)
    target_link_libraries(basic_example PRIVATE ason)
endif()

# Tests
if(ASON_BUILD_TESTS)
    enable_testing()
    
    # Fetch GoogleTest
    include(FetchContent)
    FetchContent_Declare(
        googletest
        GIT_REPOSITORY https://github.com/google/googletest.git
        GIT_TAG v1.14.0
    )
    # For Windows: Prevent overriding the parent project's compiler/linker settings
    set(gtest_force_shared_crt ON CACHE BOOL "" FORCE)
    FetchContent_MakeAvailable(googletest)
    
    add_executable(ason_test
        tests/test_lexer.cpp
        tests/test_parser.cpp
        tests/test_serializer.cpp
        tests/test_value.cpp
    )
    target_link_libraries(ason_test PRIVATE ason GTest::gtest_main)
    
    include(GoogleTest)
    gtest_discover_tests(ason_test)
endif()

# Install
install(TARGETS ason
    EXPORT asonTargets
    LIBRARY DESTINATION lib
    ARCHIVE DESTINATION lib
    RUNTIME DESTINATION bin
)
install(DIRECTORY include/ DESTINATION include)

