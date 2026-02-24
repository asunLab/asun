package main

import (
	"fmt"
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

// Format reformats the ASON source into canonical form.
func Format(src string) string {
	root, _ := Parse(src)
	if root == nil {
		return src
	}
	var sb strings.Builder
	formatNode(&sb, root, 0)
	return sb.String()
}

func formatNode(sb *strings.Builder, n *Node, indent int) {
	switch n.Kind {
	case NodeDocument:
		for _, c := range n.Children {
			formatNode(sb, c, indent)
		}

	case NodeSingleObject:
		if len(n.Children) >= 1 {
			formatNode(sb, n.Children[0], indent) // schema
		}
		sb.WriteString(":")
		if len(n.Children) >= 2 {
			formatNode(sb, n.Children[1], indent) // tuple
		}

	case NodeObjectArray:
		sb.WriteString("[")
		if len(n.Children) >= 1 {
			arrSchema := n.Children[0]
			if arrSchema.Kind == NodeArraySchema && len(arrSchema.Children) > 0 {
				formatNode(sb, arrSchema.Children[0], indent) // inner schema
			}
		}
		sb.WriteString("]:\n")
		for i := 1; i < len(n.Children); i++ {
			writeIndent(sb, indent+1)
			formatNode(sb, n.Children[i], indent+1)
			if i < len(n.Children)-1 {
				sb.WriteString(",")
			}
			sb.WriteString("\n")
		}

	case NodeSchema:
		sb.WriteString("{")
		for i, c := range n.Children {
			if i > 0 {
				sb.WriteString(", ")
			}
			formatNode(sb, c, indent)
		}
		sb.WriteString("}")

	case NodeField:
		sb.WriteString(n.Value)
		if len(n.Children) > 0 {
			c := n.Children[0]
			if c.Kind == NodeTypeAnnot {
				sb.WriteString(":")
				sb.WriteString(c.Value)
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
		sb.WriteString("(")
		for i, c := range n.Children {
			if i > 0 {
				sb.WriteString(", ")
			}
			formatNode(sb, c, indent)
		}
		sb.WriteString(")")

	case NodeArray:
		sb.WriteString("[")
		for i, c := range n.Children {
			if i > 0 {
				sb.WriteString(", ")
			}
			formatNode(sb, c, indent)
		}
		sb.WriteString("]")

	case NodeValue:
		sb.WriteString(strings.TrimSpace(n.Value))

	case NodeMapType:
		// Just write the raw token value
		sb.WriteString(n.Token.Value)
	}
}

func writeIndent(sb *strings.Builder, level int) {
	for i := 0; i < level; i++ {
		sb.WriteString("  ")
	}
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
