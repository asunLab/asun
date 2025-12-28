#pragma once

/// @file ason.hpp
/// @brief Main header file for the ASON library.
///
/// ASON (Array-Schema Object Notation) is a compact data format
/// that separates schema from data for efficient representation.

#include "error.hpp"
#include "value.hpp"
#include "lexer.hpp"
#include "parser.hpp"
#include "serializer.hpp"

namespace ason {

/// Library version
constexpr const char* VERSION = "0.1.0";

} // namespace ason

