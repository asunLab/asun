package ason

// ---------------------------------------------------------------------------
// Pretty-print encoder — smart indentation for ASON output
// ---------------------------------------------------------------------------
//
// Simple structures stay inline:
//   {name:str, age:int}:(Alice, 30)
//
// Complex structures expand with 2-space indentation:
//   {
//     id:str,
//     name:str,
//     projects:[{
//       projectId:str,
//       tasks:[{taskId:str, status:str}]
//     }]
//   }:
//     (
//       E001,
//       John,
//       [(P001, [(T001, done), (T002, pending)])]
//     )

const prettyMaxWidth = 100

// EncodePretty serializes a struct or slice of structs to pretty-formatted ASON.
func EncodePretty(v any) ([]byte, error) {
	compact, err := Encode(v)
	if err != nil {
		return nil, err
	}
	return PrettyFormat(compact), nil
}

// EncodePrettyTyped serializes with type annotations in pretty format.
func EncodePrettyTyped(v any) ([]byte, error) {
	compact, err := EncodeTyped(v)
	if err != nil {
		return nil, err
	}
	return PrettyFormat(compact), nil
}

// PrettyFormat reformats compact ASON bytes with smart indentation.
// Simple structures stay inline; complex ones expand with 2-space indentation.
func PrettyFormat(src []byte) []byte {
	n := len(src)
	if n == 0 {
		return src
	}

	// Build matching bracket table
	match := buildMatchTable(src)
	f := &prettyFmt{src: src, match: match, out: make([]byte, 0, n*2)}
	f.writeTop()
	return f.out
}

type prettyFmt struct {
	src   []byte
	match []int
	out   []byte
	pos   int
	depth int
}

func buildMatchTable(src []byte) []int {
	n := len(src)
	match := make([]int, n)
	for i := range match {
		match[i] = -1
	}
	var stack []int
	inQuote := false
	for i := 0; i < n; i++ {
		if inQuote {
			if src[i] == '\\' && i+1 < n {
				i++
				continue
			}
			if src[i] == '"' {
				inQuote = false
			}
			continue
		}
		switch src[i] {
		case '"':
			inQuote = true
		case '{', '(', '[':
			stack = append(stack, i)
		case '}', ')', ']':
			if len(stack) > 0 {
				j := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				match[j] = i
				match[i] = j
			}
		}
	}
	return match
}

// writeTop handles the top-level ASON structure
func (f *prettyFmt) writeTop() {
	if f.pos >= len(f.src) {
		return
	}

	if f.src[f.pos] == '[' && f.pos+1 < len(f.src) && f.src[f.pos+1] == '{' {
		f.writeArrayTop()
	} else if f.src[f.pos] == '{' {
		f.writeObjectTop()
	} else {
		f.out = append(f.out, f.src[f.pos:]...)
	}
}

// writeObjectTop formats {schema}:(data)
func (f *prettyFmt) writeObjectTop() {
	f.writeGroup()
	if f.pos < len(f.src) && f.src[f.pos] == ':' {
		f.out = append(f.out, ':')
		f.pos++
		if f.pos < len(f.src) {
			closePos := f.match[f.pos]
			if closePos >= 0 && closePos-f.pos+1 <= prettyMaxWidth {
				f.writeInline(f.pos, closePos+1)
				f.pos = closePos + 1
			} else {
				f.out = append(f.out, '\n')
				f.depth++
				f.writeIndent()
				f.writeGroup()
				f.depth--
			}
		}
	}
}

// writeArrayTop formats [{schema}]:(t1),(t2),...
func (f *prettyFmt) writeArrayTop() {
	f.out = append(f.out, '[')
	f.pos++ // skip [
	f.writeGroup()
	if f.pos < len(f.src) && f.src[f.pos] == ']' {
		f.out = append(f.out, ']')
		f.pos++
	}
	if f.pos < len(f.src) && f.src[f.pos] == ':' {
		f.out = append(f.out, ':', '\n')
		f.pos++
	}

	f.depth++
	first := true
	for f.pos < len(f.src) {
		if f.src[f.pos] == ',' {
			f.pos++
		}
		if f.pos >= len(f.src) {
			break
		}
		if !first {
			f.out = append(f.out, ',', '\n')
		}
		first = false
		f.writeIndent()
		f.writeGroup()
	}
	f.out = append(f.out, '\n')
	f.depth--
}

// writeGroup formats a bracket group ({...}, (...), [...])
func (f *prettyFmt) writeGroup() {
	if f.pos >= len(f.src) {
		return
	}
	ch := f.src[f.pos]
	if ch != '{' && ch != '(' && ch != '[' {
		f.writeValue()
		return
	}

	// Special case: [{...}] array schema — fuse brackets
	if ch == '[' && f.pos+1 < len(f.src) && f.src[f.pos+1] == '{' {
		closeBrace := f.match[f.pos+1]
		closeBracket := f.match[f.pos]
		if closeBrace >= 0 && closeBracket >= 0 && closeBrace+1 == closeBracket {
			width := closeBracket - f.pos + 1
			if width <= prettyMaxWidth {
				f.writeInline(f.pos, closeBracket+1)
				f.pos = closeBracket + 1
				return
			}
			f.out = append(f.out, '[')
			f.pos++
			f.writeGroup() // {schema}
			f.out = append(f.out, ']')
			f.pos++ // skip ]
			return
		}
	}

	closePos := f.match[f.pos]
	if closePos < 0 {
		f.out = append(f.out, ch)
		f.pos++
		return
	}

	width := closePos - f.pos + 1
	if width <= prettyMaxWidth {
		f.writeInline(f.pos, closePos+1)
		f.pos = closePos + 1
		return
	}

	// Expanded form
	closeCh := f.src[closePos]
	f.out = append(f.out, ch)
	f.pos++

	if f.pos >= closePos {
		f.out = append(f.out, closeCh)
		f.pos = closePos + 1
		return
	}

	f.out = append(f.out, '\n')
	f.depth++

	first := true
	for f.pos < closePos {
		if f.src[f.pos] == ',' {
			f.pos++
		}
		if !first {
			f.out = append(f.out, ',', '\n')
		}
		first = false
		f.writeIndent()
		f.writeElement(closePos)
	}

	f.out = append(f.out, '\n')
	f.depth--
	f.writeIndent()
	f.out = append(f.out, closeCh)
	f.pos = closePos + 1
}

// writeElement writes one comma-delimited element within a bracket group
func (f *prettyFmt) writeElement(boundary int) {
	for f.pos < boundary && f.src[f.pos] != ',' {
		ch := f.src[f.pos]
		if ch == '{' || ch == '(' || ch == '[' {
			f.writeGroup()
		} else if ch == '"' {
			f.writeQuoted()
		} else {
			f.out = append(f.out, ch)
			f.pos++
		}
	}
}

// writeValue writes a non-bracket value
func (f *prettyFmt) writeValue() {
	for f.pos < len(f.src) {
		ch := f.src[f.pos]
		if ch == ',' || ch == ')' || ch == '}' || ch == ']' {
			break
		}
		if ch == '"' {
			f.writeQuoted()
		} else {
			f.out = append(f.out, ch)
			f.pos++
		}
	}
}

// writeQuoted writes a quoted string including quotes
func (f *prettyFmt) writeQuoted() {
	f.out = append(f.out, '"')
	f.pos++
	for f.pos < len(f.src) {
		ch := f.src[f.pos]
		f.out = append(f.out, ch)
		f.pos++
		if ch == '\\' && f.pos < len(f.src) {
			f.out = append(f.out, f.src[f.pos])
			f.pos++
		} else if ch == '"' {
			break
		}
	}
}

// writeInline writes src[start:end] with spaces after top-level commas
func (f *prettyFmt) writeInline(start, end int) {
	depth := 0
	inQuote := false
	for i := start; i < end; i++ {
		ch := f.src[i]
		if inQuote {
			f.out = append(f.out, ch)
			if ch == '\\' && i+1 < end {
				i++
				f.out = append(f.out, f.src[i])
			} else if ch == '"' {
				inQuote = false
			}
			continue
		}
		switch ch {
		case '"':
			inQuote = true
			f.out = append(f.out, ch)
		case '{', '(', '[':
			depth++
			f.out = append(f.out, ch)
		case '}', ')', ']':
			depth--
			f.out = append(f.out, ch)
		case ',':
			f.out = append(f.out, ',')
			if depth == 1 {
				f.out = append(f.out, ' ')
			}
		default:
			f.out = append(f.out, ch)
		}
	}
}

// writeIndent writes 2-space indentation for the current depth
func (f *prettyFmt) writeIndent() {
	for i := 0; i < f.depth; i++ {
		f.out = append(f.out, ' ', ' ')
	}
}
