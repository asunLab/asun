#include "ason_parser.h"
#include <stdlib.h>
#include <string.h>
#include <stdio.h>
#include <ctype.h>

/* Forward declarations */
typedef struct schema_field schema_field;

/* Schema field types */
typedef enum {
    SCHEMA_SIMPLE,      /* Simple value */
    SCHEMA_OBJECT,      /* Nested object */
    SCHEMA_ARRAY,       /* Array of simple values */
    SCHEMA_OBJECT_ARRAY /* Array of objects */
} schema_field_type;

/* Schema field */
struct schema_field {
    char* name;
    schema_field_type type;
    schema_field* children;     /* For nested objects */
    size_t children_count;
};

/* Schema */
typedef struct {
    schema_field* fields;
    size_t count;
} ason_schema;

/* Parser state */
typedef struct {
    ason_lexer lexer;
    ason_token current;
    ason_error error;
    char error_msg[256];
} parser_state;

static void parser_init(parser_state* p, const char* input) {
    ason_lexer_init(&p->lexer, input);
    p->current = ason_lexer_next(&p->lexer);
    p->error = ASON_OK;
    p->error_msg[0] = '\0';
}

static void parser_advance(parser_state* p) {
    p->current = ason_lexer_next(&p->lexer);
}

static bool parser_match(parser_state* p, ason_token_type type) {
    return p->current.type == type;
}

static bool parser_consume(parser_state* p, ason_token_type type) {
    if (p->current.type == type) {
        parser_advance(p);
        return true;
    }
    return false;
}

static void parser_error(parser_state* p, const char* msg) {
    if (p->error == ASON_OK) {
        p->error = ASON_ERROR_PARSE_ERROR;
        snprintf(p->error_msg, sizeof(p->error_msg), "%s at line %zu, column %zu",
                 msg, p->current.line, p->current.column);
    }
}

static void schema_free(ason_schema* schema) {
    if (!schema) return;
    for (size_t i = 0; i < schema->count; i++) {
        free(schema->fields[i].name);
        if (schema->fields[i].children) {
            for (size_t j = 0; j < schema->fields[i].children_count; j++) {
                free(schema->fields[i].children[j].name);
            }
            free(schema->fields[i].children);
        }
    }
    free(schema->fields);
    schema->fields = NULL;
    schema->count = 0;
}

/* Forward declarations for parsing functions */
static ason_schema parse_schema(parser_state* p);
static ason_value* parse_value(parser_state* p);
static ason_value* parse_tuple(parser_state* p, const ason_schema* schema);

/* Check if token is a number */
static bool is_number_token(const ason_token* token) {
    if (token->len == 0) return false;
    const char* s = token->start;
    size_t i = 0;
    
    /* Optional negative sign */
    if (s[i] == '-') i++;
    if (i >= token->len) return false;
    
    /* Must start with digit */
    if (!isdigit((unsigned char)s[i])) return false;
    
    bool has_dot = false;
    for (; i < token->len; i++) {
        if (s[i] == '.') {
            if (has_dot) return false;
            has_dot = true;
        } else if (!isdigit((unsigned char)s[i])) {
            return false;
        }
    }
    return true;
}

/* Parse a primitive value from an identifier token */
static ason_value* parse_primitive(parser_state* p) {
    ason_token token = p->current;
    parser_advance(p);
    
    if (token.type == ASON_TOKEN_STRING) {
        char* str = ason_token_to_string(&token);
        ason_value* v = ason_value_string(str);
        free(str);
        return v;
    }
    
    if (token.type != ASON_TOKEN_IDENT) {
        return ason_value_null();
    }
    
    /* Check for keywords */
    if (ason_token_equals(&token, "true")) {
        return ason_value_bool(true);
    }
    if (ason_token_equals(&token, "false")) {
        return ason_value_bool(false);
    }
    if (ason_token_equals(&token, "null")) {
        return ason_value_null();
    }
    
    /* Check for number */
    if (is_number_token(&token)) {
        char* str = ason_token_to_string(&token);
        bool is_float = strchr(str, '.') != NULL;
        ason_value* v;
        if (is_float) {
            v = ason_value_float(atof(str));
        } else {
            v = ason_value_integer(atoll(str));
        }
        free(str);
        return v;
    }
    
    /* Unquoted string */
    char* str = ason_token_to_string(&token);
    ason_value* v = ason_value_string(str);
    free(str);
    return v;
}

/* Parse a standalone array [1,2,3] */
static ason_value* parse_array(parser_state* p) {
    if (!parser_consume(p, ASON_TOKEN_LBRACKET)) {
        parser_error(p, "Expected '['");
        return NULL;
    }

    ason_value* arr = ason_value_array();
    if (!arr) return NULL;

    if (parser_match(p, ASON_TOKEN_RBRACKET)) {
        parser_advance(p);
        return arr;
    }

    while (1) {
        ason_value* item = parse_value(p);
        if (!item) {
            ason_value_free(arr);
            return NULL;
        }
        ason_array_push(arr, item);

        if (parser_match(p, ASON_TOKEN_RBRACKET)) {
            parser_advance(p);
            break;
        }
        if (!parser_consume(p, ASON_TOKEN_COMMA)) {
            parser_error(p, "Expected ',' or ']'");
            ason_value_free(arr);
            return NULL;
        }
    }

    return arr;
}

/* Parse schema fields */
static ason_schema parse_schema(parser_state* p) {
    ason_schema schema = {NULL, 0};

    if (!parser_consume(p, ASON_TOKEN_LBRACE)) {
        parser_error(p, "Expected '{'");
        return schema;
    }

    size_t capacity = 4;
    schema.fields = (schema_field*)calloc(capacity, sizeof(schema_field));
    if (!schema.fields) return schema;

    while (!parser_match(p, ASON_TOKEN_RBRACE) && !parser_match(p, ASON_TOKEN_EOF)) {
        if (schema.count >= capacity) {
            capacity *= 2;
            schema_field* new_fields = (schema_field*)realloc(schema.fields, capacity * sizeof(schema_field));
            if (!new_fields) {
                schema_free(&schema);
                return schema;
            }
            schema.fields = new_fields;
        }

        if (!parser_match(p, ASON_TOKEN_IDENT)) {
            parser_error(p, "Expected field name");
            schema_free(&schema);
            return schema;
        }

        schema_field* field = &schema.fields[schema.count];
        memset(field, 0, sizeof(schema_field));
        field->name = ason_token_to_string(&p->current);
        field->type = SCHEMA_SIMPLE;
        parser_advance(p);

        /* Check for nested object */
        if (parser_match(p, ASON_TOKEN_LBRACE)) {
            ason_schema nested = parse_schema(p);
            if (p->error != ASON_OK) {
                free(field->name);
                schema_free(&schema);
                return schema;
            }
            field->type = SCHEMA_OBJECT;
            field->children = nested.fields;
            field->children_count = nested.count;
        }
        /* Check for array */
        else if (parser_match(p, ASON_TOKEN_LBRACKET)) {
            parser_advance(p); /* [ */

            if (parser_match(p, ASON_TOKEN_LBRACE)) {
                /* Array of objects: field[{...}] */
                ason_schema nested = parse_schema(p);
                if (p->error != ASON_OK) {
                    free(field->name);
                    schema_free(&schema);
                    return schema;
                }
                field->type = SCHEMA_OBJECT_ARRAY;
                field->children = nested.fields;
                field->children_count = nested.count;
            } else {
                /* Simple array: field[] */
                field->type = SCHEMA_ARRAY;
            }

            if (!parser_consume(p, ASON_TOKEN_RBRACKET)) {
                parser_error(p, "Expected ']'");
                free(field->name);
                schema_free(&schema);
                return schema;
            }
        }

        schema.count++;

        if (parser_match(p, ASON_TOKEN_COMMA)) {
            parser_advance(p);
        }
    }

    if (!parser_consume(p, ASON_TOKEN_RBRACE)) {
        parser_error(p, "Expected '}'");
        schema_free(&schema);
    }

    return schema;
}

/* Parse a tuple with schema: (value1,value2,...) */
static ason_value* parse_tuple(parser_state* p, const ason_schema* schema) {
    if (!parser_consume(p, ASON_TOKEN_LPAREN)) {
        parser_error(p, "Expected '('");
        return NULL;
    }

    ason_value* obj = ason_value_object();
    if (!obj) return NULL;

    for (size_t i = 0; i < schema->count; i++) {
        const schema_field* field = &schema->fields[i];
        ason_value* val = NULL;

        /* Check for empty value (null) */
        if (parser_match(p, ASON_TOKEN_COMMA) || parser_match(p, ASON_TOKEN_RPAREN)) {
            val = ason_value_null();
        }
        else if (field->type == SCHEMA_OBJECT && parser_match(p, ASON_TOKEN_LPAREN)) {
            /* Nested object */
            ason_schema nested = {(schema_field*)field->children, field->children_count};
            val = parse_tuple(p, &nested);
        }
        else if (field->type == SCHEMA_ARRAY && parser_match(p, ASON_TOKEN_LBRACKET)) {
            /* Simple array */
            val = parse_array(p);
        }
        else if (field->type == SCHEMA_OBJECT_ARRAY && parser_match(p, ASON_TOKEN_LBRACKET)) {
            /* Array of objects */
            parser_advance(p); /* [ */
            val = ason_value_array();
            if (!val) {
                ason_value_free(obj);
                return NULL;
            }

            while (!parser_match(p, ASON_TOKEN_RBRACKET) && !parser_match(p, ASON_TOKEN_EOF)) {
                ason_schema nested = {(schema_field*)field->children, field->children_count};
                ason_value* item = parse_tuple(p, &nested);
                if (!item) {
                    ason_value_free(val);
                    ason_value_free(obj);
                    return NULL;
                }
                ason_array_push(val, item);

                if (parser_match(p, ASON_TOKEN_COMMA)) {
                    parser_advance(p);
                }
            }

            if (!parser_consume(p, ASON_TOKEN_RBRACKET)) {
                parser_error(p, "Expected ']'");
                ason_value_free(val);
                ason_value_free(obj);
                return NULL;
            }
        }
        else {
            val = parse_primitive(p);
        }

        if (!val) {
            ason_value_free(obj);
            return NULL;
        }

        ason_object_set(obj, field->name, val);

        if (i < schema->count - 1) {
            if (!parser_consume(p, ASON_TOKEN_COMMA)) {
                /* Allow missing trailing values */
                break;
            }
        }
    }

    if (!parser_consume(p, ASON_TOKEN_RPAREN)) {
        parser_error(p, "Expected ')'");
        ason_value_free(obj);
        return NULL;
    }

    return obj;
}

/* Parse any value */
static ason_value* parse_value(parser_state* p) {
    if (parser_match(p, ASON_TOKEN_LBRACKET)) {
        return parse_array(p);
    }
    if (parser_match(p, ASON_TOKEN_LPAREN)) {
        /* Tuple without schema - parse as array */
        parser_advance(p);
        ason_value* arr = ason_value_array();
        if (!arr) return NULL;

        while (!parser_match(p, ASON_TOKEN_RPAREN) && !parser_match(p, ASON_TOKEN_EOF)) {
            ason_value* item = parse_value(p);
            if (!item) {
                ason_value_free(arr);
                return NULL;
            }
            ason_array_push(arr, item);

            if (parser_match(p, ASON_TOKEN_COMMA)) {
                parser_advance(p);
            }
        }

        if (!parser_consume(p, ASON_TOKEN_RPAREN)) {
            parser_error(p, "Expected ')'");
            ason_value_free(arr);
            return NULL;
        }
        return arr;
    }

    return parse_primitive(p);
}

/* Main parse function */
ason_parse_result ason_parse(const char* input) {
    ason_parse_result result = {NULL, ASON_OK, "", 0, 0};

    if (!input) {
        result.error = ASON_ERROR_NULL_POINTER;
        snprintf(result.error_msg, sizeof(result.error_msg), "Null input");
        return result;
    }

    parser_state p;
    parser_init(&p, input);

    /* Empty input */
    if (parser_match(&p, ASON_TOKEN_EOF)) {
        result.value = ason_value_null();
        return result;
    }

    /* Check for schema definition */
    if (parser_match(&p, ASON_TOKEN_LBRACE)) {
        ason_schema schema = parse_schema(&p);
        if (p.error != ASON_OK) {
            result.error = p.error;
            strncpy(result.error_msg, p.error_msg, sizeof(result.error_msg) - 1);
            result.error_line = p.current.line;
            result.error_column = p.current.column;
            schema_free(&schema);
            return result;
        }

        /* Expect colon */
        if (!parser_consume(&p, ASON_TOKEN_COLON)) {
            result.error = ASON_ERROR_UNEXPECTED_TOKEN;
            snprintf(result.error_msg, sizeof(result.error_msg), "Expected ':' after schema");
            schema_free(&schema);
            return result;
        }

        /* Parse first tuple */
        ason_value* first = parse_tuple(&p, &schema);
        if (p.error != ASON_OK || !first) {
            result.error = p.error;
            strncpy(result.error_msg, p.error_msg, sizeof(result.error_msg) - 1);
            result.error_line = p.current.line;
            result.error_column = p.current.column;
            schema_free(&schema);
            return result;
        }

        /* Check for multiple tuples */
        if (parser_match(&p, ASON_TOKEN_COMMA)) {
            ason_value* arr = ason_value_array();
            ason_array_push(arr, first);

            while (parser_consume(&p, ASON_TOKEN_COMMA)) {
                ason_value* next = parse_tuple(&p, &schema);
                if (!next) {
                    result.error = p.error;
                    strncpy(result.error_msg, p.error_msg, sizeof(result.error_msg) - 1);
                    ason_value_free(arr);
                    schema_free(&schema);
                    return result;
                }
                ason_array_push(arr, next);
            }

            result.value = arr;
        } else {
            result.value = first;
        }

        schema_free(&schema);
    }
    else if (parser_match(&p, ASON_TOKEN_LBRACKET)) {
        /* Standalone array */
        result.value = parse_array(&p);
    }
    else {
        /* Single value */
        result.value = parse_value(&p);
    }

    if (p.error != ASON_OK) {
        result.error = p.error;
        strncpy(result.error_msg, p.error_msg, sizeof(result.error_msg) - 1);
        result.error_line = p.current.line;
        result.error_column = p.current.column;
    }

    return result;
}

void ason_parse_result_free(ason_parse_result* result) {
    if (result && result->value) {
        ason_value_free(result->value);
        result->value = NULL;
    }
}
