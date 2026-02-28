package main

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// ──────────────────────────────────────────────────────────────────────────────
// Hover
// ──────────────────────────────────────────────────────────────────────────────

// HoverInfo returns markdown hover text for the position.
func HoverInfo(root *Node, line, col int) string {
	n := findNodeAt(root, line, col)
	if n == nil {
		return ""
	}

	switch n.Kind {
	case NodeField:
		return hoverField(n)
	case NodeTypeAnnot:
		return hoverType(n.Value)
	case NodeSchema:
		return hoverSchema(n)
	case NodeArraySchema:
		return hoverArraySchema(n)
	case NodeTuple:
		return "**Data Tuple** `(...)`\n\nOrdered values matching the schema fields."
	case NodeArray:
		return "**Array** `[...]`"
	case NodeValue:
		return hoverValue(n)
	case NodeMapType:
		return "**Map Type** `map[K,V]`\n\nKey-value pair collection."
	}
	return ""
}

func hoverField(n *Node) string {
	var sb strings.Builder
	sb.WriteString("**Field** `")
	sb.WriteString(n.Value)
	sb.WriteString("`")
	if len(n.Children) > 0 {
		c := n.Children[0]
		switch c.Kind {
		case NodeTypeAnnot:
			sb.WriteString(" : `")
			sb.WriteString(c.Value)
			sb.WriteString("`")
		case NodeSchema:
			sb.WriteString(" : nested object")
		case NodeArraySchema:
			sb.WriteString(" : object array")
		}
	}
	return sb.String()
}

func hoverType(t string) string {
	switch t {
	case "int", "integer":
		return "**Type** `int`\n\nInteger value (e.g., `42`, `-100`)"
	case "float", "double":
		return "**Type** `float`\n\nFloating-point value (e.g., `3.14`)"
	case "str", "string":
		return "**Type** `str`\n\nString value (quoted or unquoted)"
	case "bool", "boolean":
		return "**Type** `bool`\n\nBoolean: `true` or `false`"
	}
	return "**Type** `" + t + "`"
}

func hoverSchema(n *Node) string {
	fields := n.SchemaFields()
	var sb strings.Builder
	sb.WriteString("**Schema** — ")
	sb.WriteString(fmt.Sprintf("%d", len(fields)))
	sb.WriteString(" field(s)\n\n")
	for _, f := range fields {
		sb.WriteString("- `")
		sb.WriteString(f.Value)
		sb.WriteString("`")
		if len(f.Children) > 0 {
			c := f.Children[0]
			if c.Kind == NodeTypeAnnot {
				sb.WriteString(" : ")
				sb.WriteString(c.Value)
			}
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

func hoverArraySchema(n *Node) string {
	fields := n.SchemaFields()
	var sb strings.Builder
	sb.WriteString("**Array Schema** — ")
	sb.WriteString(fmt.Sprintf("%d", len(fields)))
	sb.WriteString(" field(s) per element\n\n")
	for _, f := range fields {
		sb.WriteString("- `")
		sb.WriteString(f.Value)
		sb.WriteString("`")
		if len(f.Children) > 0 {
			c := f.Children[0]
			if c.Kind == NodeTypeAnnot {
				sb.WriteString(" : ")
				sb.WriteString(c.Value)
			}
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

func hoverValue(n *Node) string {
	switch n.Token.Type {
	case TokenNumber:
		return "**Number** `" + n.Value + "`"
	case TokenBool:
		return "**Boolean** `" + n.Value + "`"
	case TokenString:
		return "**Quoted String**"
	default:
		if strings.TrimSpace(n.Value) == "" {
			return "**Null** — empty value"
		}
		return "**String** `" + strings.TrimSpace(n.Value) + "`"
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Completion
// ──────────────────────────────────────────────────────────────────────────────

// CompletionItem is an LSP completion entry.
type CompletionItem struct {
	Label      string
	Kind       int // 1=Text 6=Variable 12=Value 14=Keyword
	Detail     string
	InsertText string
}

// Complete returns completion items for the given position.
func Complete(root *Node, src string, line, col int) []CompletionItem {
	// Find what context we're in
	ctx := findContext(root, line, col)

	switch ctx {
	case contextSchemaType:
		return typeCompletions()
	case contextSchemaField:
		return schemaKeywordCompletions()
	case contextDataValue:
		return dataValueCompletions()
	case contextTopLevel:
		return topLevelCompletions()
	}
	return nil
}

type completionContext int

const (
	contextUnknown     completionContext = iota
	contextSchemaType                    // after ':' in schema
	contextSchemaField                   // inside {  }
	contextDataValue                     // inside (  ) or [ ]
	contextTopLevel                      // at the start
)

func findContext(root *Node, line, col int) completionContext {
	n := findNodeAt(root, line, col)
	if n == nil {
		return contextTopLevel
	}
	// Walk up to determine context
	switch n.Kind {
	case NodeSchema:
		return contextSchemaField
	case NodeField:
		return contextSchemaField
	case NodeTypeAnnot:
		return contextSchemaType
	case NodeTuple, NodeValue:
		return contextDataValue
	case NodeArray:
		return contextDataValue
	}
	return contextTopLevel
}

func typeCompletions() []CompletionItem {
	return []CompletionItem{
		{Label: "int", Kind: 14, Detail: "Integer type", InsertText: "int"},
		{Label: "float", Kind: 14, Detail: "Float type", InsertText: "float"},
		{Label: "str", Kind: 14, Detail: "String type", InsertText: "str"},
		{Label: "bool", Kind: 14, Detail: "Boolean type", InsertText: "bool"},
		{Label: "string", Kind: 14, Detail: "String type (alias)", InsertText: "string"},
		{Label: "integer", Kind: 14, Detail: "Integer type (alias)", InsertText: "integer"},
		{Label: "double", Kind: 14, Detail: "Float type (alias)", InsertText: "double"},
		{Label: "boolean", Kind: 14, Detail: "Boolean type (alias)", InsertText: "boolean"},
		{Label: "map", Kind: 14, Detail: "Map type", InsertText: "map[str,str]"},
	}
}

func schemaKeywordCompletions() []CompletionItem {
	return []CompletionItem{
		{Label: "field", Kind: 6, Detail: "Add a field", InsertText: "field"},
	}
}

func dataValueCompletions() []CompletionItem {
	return []CompletionItem{
		{Label: "true", Kind: 12, Detail: "Boolean true", InsertText: "true"},
		{Label: "false", Kind: 12, Detail: "Boolean false", InsertText: "false"},
	}
}

func topLevelCompletions() []CompletionItem {
	return []CompletionItem{
		{Label: "{schema}:(data)", Kind: 15, Detail: "Single object", InsertText: "{$1}:($2)"},
		{Label: "[{schema}]:(data)", Kind: 15, Detail: "Object array", InsertText: "[{$1}]:($2)"},
		{Label: "[values]", Kind: 15, Detail: "Plain array", InsertText: "[$1]"},
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Formatting
// ──────────────────────────────────────────────────────────────────────────────

// nodeComplexity returns the estimated single-line character width of a node.
// Returns -1 if the node is "structurally complex" (contains nested schemas/tuples).
func nodeComplexity(n *Node) int {
	if n == nil {
		return 0
	}
	switch n.Kind {
	case NodeDocument:
		return -1 // always expand
	case NodeSingleObject:
		return -1 // always expand schema:tuple pair
	case NodeObjectArray:
		return -1 // always expand array items
	case NodeSchema:
		total := 2 // { }
		for i, c := range n.Children {
			if i > 0 {
				total += 2 // ", "
			}
			w := nodeComplexity(c)
			if w < 0 {
				return -1
			}
			total += w
		}
		return total
	case NodeField:
		total := len(n.Value)
		if len(n.Children) > 0 {
			c := n.Children[0]
			if c.Kind == NodeTypeAnnot {
				if c.Value == "array" && len(c.Children) > 0 {
					// [innerType]
					total += 1 + 1 + len(c.Children[0].Value) + 1 // :[type]
				} else {
					total += 1 + len(c.Value) // :type
				}
			} else if c.Kind == NodeSchema {
				w := nodeComplexity(c)
				if w < 0 {
					return -1
				}
				total += 1 + w // :{...}
			} else if c.Kind == NodeArraySchema {
				if len(c.Children) > 0 {
					w := nodeComplexity(c.Children[0])
					if w < 0 {
						return -1
					}
					total += 3 + w // :[{...}]
				}
			}
		}
		return total
	case NodeTuple:
		total := 2 // ( )
		for i, c := range n.Children {
			if i > 0 {
				total += 2 // ", "
			}
			w := nodeComplexity(c)
			if w < 0 {
				return -1
			}
			total += w
		}
		return total
	case NodeArray:
		total := 2 // [ ]
		for i, c := range n.Children {
			if i > 0 {
				total += 2 // ", "
			}
			w := nodeComplexity(c)
			if w < 0 {
				return -1
			}
			total += w
		}
		return total
	case NodeValue:
		return len(strings.TrimSpace(n.Value))
	case NodeMapType:
		return len(n.Token.Value)
	default:
		return 0
	}
}

// isComplex returns true if a node should be expanded to multiple lines.
// A node is complex if it contains nested structures (schemas, tuple-with-tuples)
// or its single-line representation would exceed maxWidth characters.
const maxLineWidth = 100

func isComplex(n *Node) bool {
	w := nodeComplexity(n)
	return w < 0 || w > maxLineWidth
}

// Format reformats the ASON source into canonical form.
// Simple structures stay inline; complex/nested ones expand with indentation.
func Format(src string) string {
	root, _ := Parse(src)
	if root == nil {
		return src
	}
	var sb strings.Builder
	formatNode(&sb, root, 0)
	return sb.String()
}

// formatInline writes a node as a single line (no newlines).
// writeTypeAnnot writes a type annotation, handling [innerType] for array types.
func writeTypeAnnot(sb *strings.Builder, c *Node) {
	if c.Value == "array" && len(c.Children) > 0 {
		sb.WriteString("[")
		inner := c.Children[0]
		if inner.Kind == NodeTypeAnnot {
			sb.WriteString(inner.Value)
		} else {
			formatInline(sb, inner)
		}
		sb.WriteString("]")
	} else {
		sb.WriteString(c.Value)
	}
}

func formatInline(sb *strings.Builder, n *Node) {
	switch n.Kind {
	case NodeSchema:
		sb.WriteString("{")
		for i, c := range n.Children {
			if i > 0 {
				sb.WriteString(", ")
			}
			formatInline(sb, c)
		}
		sb.WriteString("}")
	case NodeField:
		sb.WriteString(n.Value)
		if len(n.Children) > 0 {
			c := n.Children[0]
			if c.Kind == NodeTypeAnnot {
				sb.WriteString(":")
				writeTypeAnnot(sb, c)
			} else if c.Kind == NodeSchema {
				sb.WriteString(":")
				formatInline(sb, c)
			} else if c.Kind == NodeArraySchema {
				sb.WriteString(":[")
				if len(c.Children) > 0 {
					formatInline(sb, c.Children[0])
				}
				sb.WriteString("]")
			}
		}
	case NodeTuple:
		sb.WriteString("(")
		for i, c := range n.Children {
			if i > 0 {
				sb.WriteString(", ")
			}
			formatInline(sb, c)
		}
		sb.WriteString(")")
	case NodeArray:
		sb.WriteString("[")
		for i, c := range n.Children {
			if i > 0 {
				sb.WriteString(", ")
			}
			formatInline(sb, c)
		}
		sb.WriteString("]")
	case NodeValue:
		sb.WriteString(strings.TrimSpace(n.Value))
	case NodeMapType:
		sb.WriteString(n.Token.Value)
	}
}

func formatNode(sb *strings.Builder, n *Node, indent int) {
	switch n.Kind {
	case NodeDocument:
		for _, c := range n.Children {
			formatNode(sb, c, indent)
		}

	case NodeSingleObject:
		// schema
		if len(n.Children) >= 1 {
			formatNode(sb, n.Children[0], indent)
		}
		sb.WriteString(":")
		// data tuple
		if len(n.Children) >= 2 {
			tuple := n.Children[1]
			if !isComplex(tuple) {
				formatInline(sb, tuple)
			} else {
				sb.WriteString("\n")
				writeIndent(sb, indent+1)
				formatNode(sb, tuple, indent+1)
			}
		}

	case NodeObjectArray:
		// Array schema: [{...}]:
		sb.WriteString("[")
		if len(n.Children) >= 1 {
			arrSchema := n.Children[0]
			if arrSchema.Kind == NodeArraySchema && len(arrSchema.Children) > 0 {
				formatNode(sb, arrSchema.Children[0], indent)
			}
		}
		sb.WriteString("]:\n")
		// data tuples, one per line
		for i := 1; i < len(n.Children); i++ {
			writeIndent(sb, indent+1)
			formatNode(sb, n.Children[i], indent+1)
			if i < len(n.Children)-1 {
				sb.WriteString(",")
			}
			sb.WriteString("\n")
		}

	case NodeSchema:
		// If simple enough, inline
		if !isComplex(n) {
			formatInline(sb, n)
		} else {
			sb.WriteString("{\n")
			for i, c := range n.Children {
				writeIndent(sb, indent+1)
				formatNode(sb, c, indent+1)
				if i < len(n.Children)-1 {
					sb.WriteString(",")
				}
				sb.WriteString("\n")
			}
			writeIndent(sb, indent)
			sb.WriteString("}")
		}

	case NodeField:
		sb.WriteString(n.Value)
		if len(n.Children) > 0 {
			c := n.Children[0]
			if c.Kind == NodeTypeAnnot {
				sb.WriteString(":")
				writeTypeAnnot(sb, c)
			} else if c.Kind == NodeSchema {
				sb.WriteString(":")
				formatNode(sb, c, indent)
			} else if c.Kind == NodeArraySchema {
				sb.WriteString(":[")
				if len(c.Children) > 0 {
					formatNode(sb, c.Children[0], indent)
				}
				sb.WriteString("]")
			}
		}

	case NodeTuple:
		// If simple enough, inline
		if !isComplex(n) {
			formatInline(sb, n)
		} else {
			sb.WriteString("(\n")
			for i, c := range n.Children {
				writeIndent(sb, indent+1)
				formatNode(sb, c, indent+1)
				if i < len(n.Children)-1 {
					sb.WriteString(",")
				}
				sb.WriteString("\n")
			}
			writeIndent(sb, indent)
			sb.WriteString(")")
		}

	case NodeArray:
		// If simple enough, inline
		if !isComplex(n) {
			formatInline(sb, n)
		} else {
			sb.WriteString("[\n")
			for i, c := range n.Children {
				writeIndent(sb, indent+1)
				formatNode(sb, c, indent+1)
				if i < len(n.Children)-1 {
					sb.WriteString(",")
				}
				sb.WriteString("\n")
			}
			writeIndent(sb, indent)
			sb.WriteString("]")
		}

	case NodeValue:
		sb.WriteString(strings.TrimSpace(n.Value))

	case NodeMapType:
		sb.WriteString(n.Token.Value)
	}
}

func writeIndent(sb *strings.Builder, level int) {
	for i := 0; i < level; i++ {
		sb.WriteString("  ")
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Inlay Hints — show field names before data values
// ──────────────────────────────────────────────────────────────────────────────

// InlayHint represents a field annotation shown inline.
type InlayHint struct {
	Line  int
	Col   int
	Label string
}

// InlayHints returns inlay hints for field names before data values.
func InlayHints(root *Node) []InlayHint {
	if root == nil {
		return nil
	}
	var hints []InlayHint
	var walk func(n *Node)
	walk = func(n *Node) {
		if n == nil {
			return
		}
		switch n.Kind {
		case NodeSingleObject:
			if len(n.Children) >= 2 {
				schema := n.Children[0]
				tuple := n.Children[1]
				hints = append(hints, tupleHints(schema, tuple)...)
			}
		case NodeObjectArray:
			if len(n.Children) >= 1 {
				arrSchema := n.Children[0]
				for i := 1; i < len(n.Children); i++ {
					if n.Children[i].Kind == NodeTuple {
						hints = append(hints, tupleHints(arrSchema, n.Children[i])...)
					}
				}
			}
		}
		for _, c := range n.Children {
			walk(c)
		}
	}
	walk(root)
	return hints
}

// tupleHints generates hints for one tuple matched against a schema.
func tupleHints(schema *Node, tuple *Node) []InlayHint {
	fields := schema.SchemaFields()
	if len(fields) == 0 || tuple.Kind != NodeTuple {
		return nil
	}
	var hints []InlayHint
	for i, child := range tuple.Children {
		if i >= len(fields) {
			break
		}
		fieldName := fields[i].Value
		label := fieldName + ":"
		hints = append(hints, InlayHint{
			Line:  child.Token.Line,
			Col:   child.Token.Col,
			Label: label,
		})
		// Recurse into nested structures
		if len(fields[i].Children) > 0 {
			fc := fields[i].Children[0]
			switch fc.Kind {
			case NodeSchema:
				// field:{nested_schema} — data is a tuple
				if child.Kind == NodeTuple {
					hints = append(hints, tupleHints(fc, child)...)
				}
			case NodeArraySchema:
				// field:[{nested_schema}] — data is an array of tuples
				if child.Kind == NodeArray {
					for _, elem := range child.Children {
						if elem.Kind == NodeTuple {
							hints = append(hints, tupleHints(fc, elem)...)
						}
					}
				}
			}
		}
	}
	return hints
}

// ──────────────────────────────────────────────────────────────────────────────
// Compress — single-line minified output
// ──────────────────────────────────────────────────────────────────────────────

// Compress reformats ASON source into a single-line compact form.
func Compress(src string) string {
	root, _ := Parse(src)
	if root == nil {
		return src
	}
	var sb strings.Builder
	compressNode(&sb, root)
	return sb.String()
}

func compressNode(sb *strings.Builder, n *Node) {
	switch n.Kind {
	case NodeDocument:
		for _, c := range n.Children {
			compressNode(sb, c)
		}
	case NodeSingleObject:
		if len(n.Children) >= 1 {
			compressNode(sb, n.Children[0])
		}
		sb.WriteString(":")
		if len(n.Children) >= 2 {
			compressNode(sb, n.Children[1])
		}
	case NodeObjectArray:
		sb.WriteString("[")
		if len(n.Children) >= 1 {
			arrSchema := n.Children[0]
			if arrSchema.Kind == NodeArraySchema && len(arrSchema.Children) > 0 {
				compressNode(sb, arrSchema.Children[0])
			}
		}
		sb.WriteString("]:")
		for i := 1; i < len(n.Children); i++ {
			if i > 1 {
				sb.WriteString(",")
			}
			compressNode(sb, n.Children[i])
		}
	case NodeSchema:
		sb.WriteString("{")
		for i, c := range n.Children {
			if i > 0 {
				sb.WriteString(",")
			}
			compressNode(sb, c)
		}
		sb.WriteString("}")
	case NodeField:
		sb.WriteString(n.Value)
		if len(n.Children) > 0 {
			c := n.Children[0]
			if c.Kind == NodeTypeAnnot {
				sb.WriteString(":")
				writeTypeAnnot(sb, c)
			} else if c.Kind == NodeSchema {
				sb.WriteString(":")
				compressNode(sb, c)
			} else if c.Kind == NodeArraySchema {
				sb.WriteString(":[")
				if len(c.Children) > 0 {
					compressNode(sb, c.Children[0])
				}
				sb.WriteString("]")
			}
		}
	case NodeTuple:
		sb.WriteString("(")
		for i, c := range n.Children {
			if i > 0 {
				sb.WriteString(",")
			}
			compressNode(sb, c)
		}
		sb.WriteString(")")
	case NodeArray:
		sb.WriteString("[")
		for i, c := range n.Children {
			if i > 0 {
				sb.WriteString(",")
			}
			compressNode(sb, c)
		}
		sb.WriteString("]")
	case NodeValue:
		sb.WriteString(strings.TrimSpace(n.Value))
	case NodeMapType:
		sb.WriteString(n.Token.Value)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// ASON → JSON conversion
// ──────────────────────────────────────────────────────────────────────────────

// AsonToJSON converts ASON source to pretty-printed JSON.
func AsonToJSON(src string) (string, error) {
	root, diags := Parse(src)
	if root == nil {
		return "", fmt.Errorf("parse failed")
	}
	for _, d := range diags {
		if d.Severity == SeverityError {
			return "", fmt.Errorf("parse error: %s", d.Message)
		}
	}
	val := nodeToJSON(root)
	data, err := json.MarshalIndent(val, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func nodeToJSON(n *Node) interface{} {
	if n == nil {
		return nil
	}
	switch n.Kind {
	case NodeDocument:
		if len(n.Children) == 0 {
			return nil
		}
		return nodeToJSON(n.Children[0])

	case NodeSingleObject:
		if len(n.Children) < 2 {
			return map[string]interface{}{}
		}
		schema := n.Children[0]
		tuple := n.Children[1]
		return tupleToJSONObject(schema, tuple)

	case NodeObjectArray:
		if len(n.Children) < 1 {
			return []interface{}{}
		}
		arrSchema := n.Children[0]
		var result []interface{}
		for i := 1; i < len(n.Children); i++ {
			if n.Children[i].Kind == NodeTuple {
				result = append(result, tupleToJSONObject(arrSchema, n.Children[i]))
			}
		}
		if result == nil {
			result = []interface{}{}
		}
		return result

	case NodeArray:
		var result []interface{}
		for _, c := range n.Children {
			result = append(result, nodeToJSON(c))
		}
		if result == nil {
			result = []interface{}{}
		}
		return result

	case NodeValue:
		return valueToJSON(n)

	case NodeTuple:
		// Tuple without schema context → array
		var result []interface{}
		for _, c := range n.Children {
			result = append(result, nodeToJSON(c))
		}
		return result
	}
	return nil
}

func tupleToJSONObject(schema *Node, tuple *Node) map[string]interface{} {
	fields := schema.SchemaFields()
	obj := make(map[string]interface{}, len(fields))
	for i, field := range fields {
		if i >= len(tuple.Children) {
			obj[field.Value] = nil
			continue
		}
		child := tuple.Children[i]
		// Check if field has nested schema
		hasNestedSchema := false
		for _, fc := range field.Children {
			if fc.Kind == NodeSchema && child.Kind == NodeTuple {
				obj[field.Value] = tupleToJSONObject(fc, child)
				hasNestedSchema = true
				break
			}
			if fc.Kind == NodeArraySchema && child.Kind == NodeArray {
				// Array of objects: [{schema}] → [obj, obj, ...]
				innerSchema := fc
				var arr []interface{}
				for _, elem := range child.Children {
					if elem.Kind == NodeTuple {
						arr = append(arr, tupleToJSONObject(innerSchema, elem))
					} else {
						arr = append(arr, nodeToJSON(elem))
					}
				}
				if arr == nil {
					arr = []interface{}{}
				}
				obj[field.Value] = arr
				hasNestedSchema = true
				break
			}
		}
		if !hasNestedSchema {
			obj[field.Value] = nodeToJSON(child)
		}
	}
	return obj
}

func valueToJSON(n *Node) interface{} {
	v := strings.TrimSpace(n.Value)
	if v == "" {
		return nil
	}
	switch n.Token.Type {
	case TokenNumber:
		if strings.Contains(v, ".") {
			f, err := strconv.ParseFloat(v, 64)
			if err == nil {
				return f
			}
		}
		i, err := strconv.ParseInt(v, 10, 64)
		if err == nil {
			return i
		}
		return v
	case TokenBool:
		return v == "true"
	case TokenString:
		// Remove surrounding quotes and unescape
		if len(v) >= 2 && v[0] == '"' && v[len(v)-1] == '"' {
			unq, err := strconv.Unquote(v)
			if err == nil {
				return unq
			}
			return v[1 : len(v)-1]
		}
		return v
	default:
		// PlainStr - try to infer type
		if v == "true" {
			return true
		}
		if v == "false" {
			return false
		}
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			return i
		}
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
		return v
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// JSON → ASON conversion
// ──────────────────────────────────────────────────────────────────────────────

// JSONToASON converts JSON source to ASON format.
func JSONToASON(src string) (string, error) {
	var raw interface{}
	if err := json.Unmarshal([]byte(src), &raw); err != nil {
		return "", fmt.Errorf("invalid JSON: %v", err)
	}
	return jsonValToASON(raw), nil
}

func jsonValToASON(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case map[string]interface{}:
		return jsonObjectToASON(val)
	case []interface{}:
		return jsonArrayToASON(val)
	case string:
		if needsQuote(val) {
			return strconv.Quote(val)
		}
		return val
	case float64:
		if val == math.Trunc(val) && !math.IsInf(val, 0) && !math.IsNaN(val) {
			return strconv.FormatInt(int64(val), 10)
		}
		return strconv.FormatFloat(val, 'f', -1, 64)
	case bool:
		if val {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprintf("%v", val)
	}
}

func jsonObjectToASON(obj map[string]interface{}) string {
	if len(obj) == 0 {
		return "{}:()"
	}
	// Collect keys in order (Go maps are unordered, but we'll sort for consistency)
	keys := sortedKeys(obj)
	var schemaParts []string
	var dataParts []string
	for _, k := range keys {
		v := obj[k]
		fieldSchema, fieldData := jsonFieldToASON(k, v)
		schemaParts = append(schemaParts, fieldSchema)
		dataParts = append(dataParts, fieldData)
	}
	return "{" + strings.Join(schemaParts, ",") + "}:(" + strings.Join(dataParts, ",") + ")"
}

func jsonFieldToASON(key string, val interface{}) (schema string, data string) {
	qk := quoteKey(key)
	if val == nil {
		return qk + ":str", ""
	}
	switch v := val.(type) {
	case map[string]interface{}:
		// Nested object
		innerKeys := sortedKeys(v)
		var innerSchema []string
		var innerData []string
		for _, ik := range innerKeys {
			fs, fd := jsonFieldToASON(ik, v[ik])
			innerSchema = append(innerSchema, fs)
			innerData = append(innerData, fd)
		}
		schema = qk + ":{" + strings.Join(innerSchema, ",") + "}"
		data = "(" + strings.Join(innerData, ",") + ")"
		return
	case []interface{}:
		// Check if array of uniform objects
		if len(v) > 0 {
			if firstObj, ok := v[0].(map[string]interface{}); ok {
				// Object array → [{schema}]
				innerKeys := sortedKeys(firstObj)
				var innerSchema []string
				for _, ik := range innerKeys {
					fs, _ := jsonFieldToASON(ik, firstObj[ik])
					innerSchema = append(innerSchema, fs)
				}
				schema = qk + ":[{" + strings.Join(innerSchema, ",") + "}]"
				var tuples []string
				for _, elem := range v {
					if eObj, ok2 := elem.(map[string]interface{}); ok2 {
						var dParts []string
						for _, ik := range innerKeys {
							_, fd := jsonFieldToASON(ik, eObj[ik])
							dParts = append(dParts, fd)
						}
						tuples = append(tuples, "("+strings.Join(dParts, ",")+")")
					}
				}
				data = "[" + strings.Join(tuples, ",") + "]"
				return
			}
		}
		// Plain array
		elemType := inferArrayType(v)
		schema = qk + ":[" + elemType + "]"
		var elems []string
		for _, elem := range v {
			elems = append(elems, jsonValToASON(elem))
		}
		data = "[" + strings.Join(elems, ",") + "]"
		return
	case float64:
		if v == math.Trunc(v) && !math.IsInf(v, 0) && !math.IsNaN(v) {
			schema = qk + ":int"
			data = strconv.FormatInt(int64(v), 10)
		} else {
			schema = qk + ":float"
			data = strconv.FormatFloat(v, 'f', -1, 64)
		}
		return
	case bool:
		schema = qk + ":bool"
		if v {
			data = "true"
		} else {
			data = "false"
		}
		return
	case string:
		schema = qk + ":str"
		if needsQuote(v) {
			data = strconv.Quote(v)
		} else {
			data = v
		}
		return
	default:
		schema = qk + ":str"
		data = fmt.Sprintf("%v", v)
		return
	}
}

func jsonArrayToASON(arr []interface{}) string {
	if len(arr) == 0 {
		return "[str]"
	}
	// Check if array of objects with same schema
	if firstObj, ok := arr[0].(map[string]interface{}); ok {
		keys := sortedKeys(firstObj)
		var schemaParts []string
		for _, k := range keys {
			fs, _ := jsonFieldToASON(k, firstObj[k])
			schemaParts = append(schemaParts, fs)
		}
		var tuples []string
		for _, elem := range arr {
			if eObj, ok2 := elem.(map[string]interface{}); ok2 {
				var dParts []string
				for _, k := range keys {
					_, fd := jsonFieldToASON(k, eObj[k])
					dParts = append(dParts, fd)
				}
				tuples = append(tuples, "("+strings.Join(dParts, ",")+")")
			}
		}
		result := "[{" + strings.Join(schemaParts, ",") + "}]:\n"
		for i, t := range tuples {
			result += "  " + t
			if i < len(tuples)-1 {
				result += ","
			}
			result += "\n"
		}
		return strings.TrimRight(result, "\n")
	}
	// Plain array
	var elems []string
	for _, elem := range arr {
		elems = append(elems, jsonValToASON(elem))
	}
	return "[" + strings.Join(elems, ",") + "]"
}

func inferArrayType(arr []interface{}) string {
	if len(arr) == 0 {
		return "str"
	}
	switch arr[0].(type) {
	case float64:
		for _, v := range arr {
			if f, ok := v.(float64); ok {
				if f != math.Trunc(f) {
					return "float"
				}
			}
		}
		return "int"
	case bool:
		return "bool"
	case string:
		return "str"
	default:
		return "str"
	}
}

func needsQuote(s string) bool {
	if s == "" {
		return true
	}
	for _, c := range s {
		if c == ',' || c == ')' || c == '(' || c == '[' || c == ']' ||
			c == '{' || c == '}' || c == ':' || c == '"' || c == '\n' ||
			c == '\r' || c == '\t' || c == ' ' || c == '/' {
			return true
		}
	}
	return false
}

// needsKeyQuote returns true if a JSON key contains characters that are not
// valid ASON identifier characters [a-zA-Z0-9_+-].
func needsKeyQuote(s string) bool {
	if s == "" {
		return true
	}
	for _, c := range s {
		ok := (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '_' || c == '+' || c == '-'
		if !ok {
			return true
		}
	}
	return false
}

// quoteKey returns the key quoted if necessary for ASON field names.
func quoteKey(key string) string {
	if needsKeyQuote(key) {
		return strconv.Quote(key)
	}
	return key
}

func sortedKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	// Simple sort
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] > keys[j] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
	return keys
}

// ──────────────────────────────────────────────────────────────────────────────
// Find node at position
// ──────────────────────────────────────────────────────────────────────────────

func findNodeAt(root *Node, line, col int) *Node {
	if root == nil {
		return nil
	}
	var best *Node
	var walk func(n *Node)
	walk = func(n *Node) {
		if n == nil {
			return
		}
		// Check if position is within this node
		startOk := n.Token.Line < line || (n.Token.Line == line && n.Token.Col <= col)
		endLine := n.EndToken.EndLine
		endCol := n.EndToken.EndCol
		if n.EndToken.Type == 0 && n.Token.Type != 0 {
			endLine = n.Token.EndLine
			endCol = n.Token.EndCol
		}
		endOk := endLine > line || (endLine == line && endCol >= col)

		if startOk && endOk {
			best = n
		}
		for _, c := range n.Children {
			walk(c)
		}
	}
	walk(root)
	return best
}
