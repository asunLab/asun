/**
 * @file ason_parser.h
 * @brief ASON Parser - parses tokens into values
 */

#ifndef ASON_PARSER_H
#define ASON_PARSER_H

#include "ason_value.h"
#include "ason_lexer.h"

#ifdef __cplusplus
extern "C" {
#endif

/* Parse result */
typedef struct {
    ason_value* value;      /* Parsed value (NULL on error) */
    ason_error error;       /* Error code */
    char error_msg[256];    /* Error message */
    size_t error_line;      /* Line number of error */
    size_t error_column;    /* Column number of error */
} ason_parse_result;

/**
 * Parse an ASON string into a value.
 * 
 * @param input The ASON string to parse
 * @return Parse result containing value or error
 * 
 * The caller is responsible for freeing the returned value with ason_value_free().
 * 
 * Examples:
 *   "{name,age}:(Alice,30)" -> Object with name and age
 *   "{name,age}:(Alice,30),(Bob,25)" -> Array of objects
 *   "[1,2,3]" -> Array of integers
 */
ason_parse_result ason_parse(const char* input);

/**
 * Free a parse result (frees the value if present)
 */
void ason_parse_result_free(ason_parse_result* result);

#ifdef __cplusplus
}
#endif

#endif /* ASON_PARSER_H */

