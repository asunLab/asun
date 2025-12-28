/**
 * @file ason_serializer.h
 * @brief ASON Serializer - converts values to strings
 */

#ifndef ASON_SERIALIZER_H
#define ASON_SERIALIZER_H

#include "ason_value.h"

#ifdef __cplusplus
extern "C" {
#endif

/**
 * Serialize a value to an ASON string.
 * 
 * @param value The value to serialize
 * @return Newly allocated string (caller must free), or NULL on error
 */
char* ason_to_string(const ason_value* value);

/**
 * Serialize a value to a pretty-printed ASON string.
 * 
 * @param value The value to serialize
 * @return Newly allocated string (caller must free), or NULL on error
 */
char* ason_to_string_pretty(const ason_value* value);

#ifdef __cplusplus
}
#endif

#endif /* ASON_SERIALIZER_H */

