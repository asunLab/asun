cmake_minimum_required(VERSION 3.14)
project(ason_c VERSION 0.1.0 LANGUAGES C)

set(CMAKE_C_STANDARD 11)
set(CMAKE_C_STANDARD_REQUIRED ON)

# Options
option(ASON_BUILD_TESTS "Build tests" ON)
option(ASON_BUILD_EXAMPLES "Build examples" ON)

# Library sources
set(ASON_SOURCES
    src/ason_value.c
    src/ason_lexer.c
    src/ason_parser.c
    src/ason_serializer.c
)

# Static library
add_library(ason_c STATIC ${ASON_SOURCES})
target_include_directories(ason_c PUBLIC ${CMAKE_CURRENT_SOURCE_DIR}/include)
target_compile_options(ason_c PRIVATE -Wall -Wextra -Wpedantic)

# Examples
if(ASON_BUILD_EXAMPLES)
    add_executable(basic_example_c examples/basic.c)
    target_link_libraries(basic_example_c ason_c)
endif()

# Tests
if(ASON_BUILD_TESTS)
    enable_testing()
    
    add_executable(ason_c_test
        tests/test_main.c
        tests/test_value.c
        tests/test_lexer.c
        tests/test_parser.c
        tests/test_serializer.c
    )
    target_link_libraries(ason_c_test ason_c)
    target_compile_options(ason_c_test PRIVATE -Wall -Wextra)
    
    add_test(NAME ason_c_test COMMAND ason_c_test)
endif()

# Install
install(TARGETS ason_c ARCHIVE DESTINATION lib)
install(DIRECTORY include/ DESTINATION include)

