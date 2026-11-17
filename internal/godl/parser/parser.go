package parser

import (
	"errors"
	"fmt"
	"strings"
)

// Error is a syntax error with its position.
type Error struct {
	File string
	Line int
	Msg  string
}

func (e *Error) Error() string {
	if e.File != "" {
		return fmt.Sprintf("%s:%d: %s", e.File, e.Line, e.Msg)
	}
	return fmt.Sprintf("line %d: %s", e.Line, e.Msg)
}

// Parse parses GoDL source.
func Parse(src string) ([]*Definition, error) {
	return ParseFile("", src)
}

// ParseFile parses GoDL source.
//
// Syntax:
//
//	; full-line comment
//	[Kind: Name]                   new definition
//	[#Kind: Name]                  modify an existing definition
//	    Attribute : value          attribute; value kept raw
//	    Add|Delete|Replace : Attribute : value
//	    value text \               continues on the next line
//
// Syntax errors are collected and returned together.
// Valid definitions are returned even when an error occurs.
func ParseFile(file, src string) ([]*Definition, error) {
	var (
		defs []*Definition
		errs []error
		cur  *Definition
	)

	fail := func(line int, format string, args ...any) {
		errs = append(errs, &Error{
			File: file,
			Line: line,
			Msg:  fmt.Sprintf(format, args...),
		})
	}

	src = strings.ReplaceAll(src, "\r\n", "\n")
	src = strings.ReplaceAll(src, "\r", "\n")

	lines := strings.Split(src, "\n")

	for i := 0; i < len(lines); i++ {
		lineNo := i + 1
		raw := lines[i]

		// Join backslash-continued lines.
		// Trim both sides so the boundary always contains one space.
		for hasContinuation(raw) {
			if i+1 >= len(lines) {
				fail(lineNo, "line continuation has no following line")
				raw = removeContinuation(raw)
				break
			}

			raw = strings.TrimSpace(removeContinuation(raw)) +
				" " +
				strings.TrimSpace(lines[i+1])

			i++
		}

		line := strings.TrimSpace(stripComment(raw))
		if line == "" {
			continue
		}

		// Definition header: [Kind: Name] or [#Kind: Name]
		if strings.HasPrefix(line, "[") {
			if !strings.HasSuffix(line, "]") {
				fail(lineNo, "unterminated definition header %q", line)
				cur = nil
				continue
			}

			head := strings.TrimSpace(line[1 : len(line)-1])

			modify := strings.HasPrefix(head, "#")
			if modify {
				head = strings.TrimSpace(strings.TrimPrefix(head, "#"))
			}

			kind, name, ok := strings.Cut(head, ":")
			kind = strings.TrimSpace(kind)
			name = strings.TrimSpace(name)

			if !ok || kind == "" || name == "" {
				fail(
					lineNo,
					"definition header must look like [Kind: Name], got %q",
					line,
				)
				cur = nil
				continue
			}

			cur = &Definition{
				Kind:   kind,
				Name:   name,
				Modify: modify,
				File:   file,
				Line:   lineNo,
			}

			defs = append(defs, cur)
			continue
		}

		// Attributes must belong to a definition.
		if cur == nil {
			fail(lineNo, "attribute outside any definition: %q", line)
			continue
		}

		idx := indexTop(line, ':')
		if idx < 0 {
			fail(lineNo, "expected 'Attribute : value', got %q", line)
			continue
		}

		name := strings.TrimSpace(line[:idx])
		value := strings.TrimSpace(line[idx+1:])

		if name == "" {
			fail(lineNo, "attribute name is empty in %q", line)
			continue
		}

		attr := Attr{
			Name:  name,
			Value: value,
			Line:  lineNo,
		}

		// Handle modification operations.
		switch Norm(name) {
		case "add", "delete", "replace":
			if !cur.Modify {
				fail(
					lineNo,
					"%s is only valid inside a [#Kind: Name] modify block",
					name,
				)
				continue
			}

			j := indexTop(value, ':')
			target := value
			rest := ""

			if j >= 0 {
				target = strings.TrimSpace(value[:j])
				rest = strings.TrimSpace(value[j+1:])
			}

			if target == "" {
				fail(
					lineNo,
					"%s needs a target attribute: '%s : Attribute : value'",
					name,
					name,
				)
				continue
			}

			attr.Op = Op(Norm(name))
			attr.Name = target
			attr.Value = rest
		}

		cur.Attrs = append(cur.Attrs, attr)
	}

	return defs, errors.Join(errs...)
}

// hasContinuation reports whether a line ends with a backslash,
// ignoring trailing spaces and tabs.
func hasContinuation(line string) bool {
	return strings.HasSuffix(strings.TrimRight(line, " \t"), `\`)
}

// removeContinuation removes the trailing continuation backslash
// and any spaces or tabs after it.
func removeContinuation(line string) string {
	line = strings.TrimRight(line, " \t")
	return strings.TrimSuffix(line, `\`)
}

// stripComment removes a ';' comment.
//
// A semicolon starts a comment when it is:
//   - At the beginning of the line, or
//   - Preceded by whitespace.
//
// Semicolons inside double-quoted strings are preserved.
func stripComment(s string) string {
	inQuote := false

	for i := 0; i < len(s); i++ {
		switch {
		case s[i] == '"':
			inQuote = !inQuote

		case s[i] == ';' &&
			!inQuote &&
			(i == 0 || s[i-1] == ' ' || s[i-1] == '\t'):
			return s[:i]
		}
	}

	return s
}
