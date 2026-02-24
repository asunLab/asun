package main

import "fmt"

// Analyze runs additional semantic checks on the AST and returns diagnostics.
func Analyze(root *Node, src string) []Diagnostic {
	a := &analyzer{src: src}
	a.walk(root, nil)
	return a.diags
}

type analyzer struct {
	src   string
	diags []Diagnostic
}

func (a *analyzer) addDiag(n *Node, msg string, sev DiagSeverity) {
	a.diags = append(a.diags, Diagnostic{
		StartLine: n.Token.Line,
		StartCol:  n.Token.Col,
		EndLine:   n.EndToken.EndLine,
		EndCol:    n.EndToken.EndCol,
		Message:   msg,
		Severity:  sev,
	})
}

func (a *analyzer) walk(n *Node, parent *Node) {
	if n == nil {
		return
	}

	switch n.Kind {
	case NodeField:
		a.checkFieldName(n)

	case NodeSingleObject:
		a.checkSingleObject(n)

	case NodeObjectArray:
		a.checkObjectArray(n)

	case NodeValue:
		// Check for unquoted strings that look suspiciously like numbers with trailing text
		if n.Token.Type == TokenPlainStr {
			v := n.Value
			if len(v) > 0 && (v[0] >= '0' && v[0] <= '9' || v[0] == '-') {
				// Might be a badly typed number like "123abc"
				if !isNumber(v) && v != "true" && v != "false" {
					a.diags = append(a.diags, Diagnostic{
						StartLine: n.Token.Line,
						StartCol:  n.Token.Col,
						EndLine:   n.Token.EndLine,
						EndCol:    n.Token.EndCol,
						Message:   fmt.Sprintf("value %q starts with a digit but is not a valid number", v),
						Severity:  SeverityWarning,
					})
				}
			}
		}
	}

	for _, child := range n.Children {
		a.walk(child, n)
	}
}

// checkFieldName validates field name characters.
func (a *analyzer) checkFieldName(n *Node) {
	name := n.Value
	if name == "" {
		return
	}
	for i := 0; i < len(name); i++ {
		b := name[i]
		if !isIdentChar(b) {
			a.diags = append(a.diags, Diagnostic{
				StartLine: n.Token.Line,
				StartCol:  n.Token.Col + i,
				EndLine:   n.Token.Line,
				EndCol:    n.Token.Col + i + 1,
				Message:   fmt.Sprintf("illegal character '%c' in field name", b),
				Severity:  SeverityError,
			})
		}
	}
}

// checkSingleObject validates data tuple against schema.
func (a *analyzer) checkSingleObject(n *Node) {
	if len(n.Children) < 2 {
		return
	}
	schema := n.Children[0]
	tuple := n.Children[1]
	a.checkDataAgainstSchema(schema, tuple)
}

// checkObjectArray validates each data tuple in array against schema.
func (a *analyzer) checkObjectArray(n *Node) {
	if len(n.Children) < 1 {
		return
	}
	arrSchema := n.Children[0]
	for i := 1; i < len(n.Children); i++ {
		tuple := n.Children[i]
		if tuple.Kind == NodeTuple {
			a.checkDataAgainstSchema(arrSchema, tuple)
		}
	}
}

// checkDataAgainstSchema checks nested tuple structures match nested schemas.
func (a *analyzer) checkDataAgainstSchema(schema *Node, tuple *Node) {
	fields := schema.SchemaFields()
	if len(fields) == 0 || tuple.Kind != NodeTuple {
		return
	}

	// Already checked count in parser; do nested checks here.
	for i, field := range fields {
		if i >= len(tuple.Children) {
			break
		}
		child := tuple.Children[i]
		// If field has a nested schema, check the tuple recursively
		for _, fc := range field.Children {
			if fc.Kind == NodeSchema && child.Kind == NodeTuple {
				subFields := fc.SchemaFields()
				subCount := len(child.Children)
				if len(subFields) > 0 && subCount != len(subFields) {
					a.diags = append(a.diags, Diagnostic{
						StartLine: child.Token.Line,
						StartCol:  child.Token.Col,
						EndLine:   child.EndToken.EndLine,
						EndCol:    child.EndToken.EndCol,
						Message: fmt.Sprintf("nested field count mismatch: schema has %d fields, data has %d values",
							len(subFields), subCount),
						Severity: SeverityError,
					})
				}
			}
		}
	}
}
