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

// Parse parses GoDL source. See ParseFile.
func Parse(src string) ([]*Definition, error) { return ParseFile("", src) }

// ParseFile parses GoDL source; file is only used in error messages and is
// recorded on each Definition.
//
// Syntax:
//
//	; full-line comment            (also "  ; trailing" outside quotes)
//	[Kind: Name]                   new definition
//	[#Kind: Name]                  modify an existing definition
//	    Attribute : value          attribute; value kept raw
//	    Add|Delete|Replace : Attribute : value    (only inside [#...])
//	    value text \               trailing backslash continues on next line
//
// All syntax errors are collected (parsing continues after an error) and
// returned joined, so one run reports every problem. The returned definitions
// are usable even when err != nil: broken lines are skipped.
func ParseFile(file, src string) ([]*Definition, error) {
	var (
		defs []*Definition
		errs []error
		cur  *Definition
	)
	fail := func(line int, format string, a ...any) {
		errs = append(errs, &Error{File: file, Line: line, Msg: fmt.Sprintf(format, a...)})
	}

	lines := strings.Split(strings.ReplaceAll(src, "\r\n", "\n"), "\n")
	for i := 0; i < len(lines); i++ {
		lineNo := i + 1
		raw := lines[i]

		// join backslash continuations
		for strings.HasSuffix(strings.TrimRight(raw, " \t"), `\`) && i+1 < len(lines) {
			raw = strings.TrimSuffix(strings.TrimRight(raw, " \t"), `\`) + " " + strings.TrimSpace(lines[i+1])
			i++
		}

		line := strings.TrimSpace(stripComment(raw))
		if line == "" {
			continue
		}

		// ---- header: [Kind: Name] / [#Kind: Name]
		if strings.HasPrefix(line, "[") {
			if !strings.HasSuffix(line, "]") {
				fail(lineNo, "unterminated definition header %q", line)
				cur = nil
				continue
			}
			head := strings.TrimSpace(line[1 : len(line)-1])
			modify := strings.HasPrefix(head, "#")
			head = strings.TrimSpace(strings.TrimPrefix(head, "#"))
			kind, name, ok := strings.Cut(head, ":")
			kind, name = strings.TrimSpace(kind), strings.TrimSpace(name)
			if !ok || kind == "" || name == "" {
				fail(lineNo, "definition header must look like [Kind: Name], got %q", line)
				cur = nil
				continue
			}
			cur = &Definition{Kind: kind, Name: name, Modify: modify, File: file, Line: lineNo}
			defs = append(defs, cur)
			continue
		}

		// ---- attribute
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

		attr := Attr{Name: name, Value: value, Line: lineNo}
		switch Norm(name) {
		case "add", "delete", "replace":
			if !cur.Modify {
				fail(lineNo, "%s is only valid inside a [#Kind: Name] modify block", name)
				continue
			}
			j := indexTop(value, ':')
			target, rest := value, ""
			if j >= 0 {
				target, rest = strings.TrimSpace(value[:j]), strings.TrimSpace(value[j+1:])
			}
			if target == "" {
				fail(lineNo, "%s needs a target attribute: '%s : Attribute : value'", name, name)
				continue
			}
			attr.Op = Op(Norm(name))
			attr.Name, attr.Value = target, rest
		}
		cur.Attrs = append(cur.Attrs, attr)
	}
	return defs, errors.Join(errs...)
}

// stripComment removes a ';' comment: a ';' at the start of the line, or one
// preceded by whitespace, outside double quotes.
func stripComment(s string) string {
	inQuote := false
	for i := 0; i < len(s); i++ {
		switch c := s[i]; {
		case c == '"':
			inQuote = !inQuote
		case c == ';' && !inQuote && (i == 0 || s[i-1] == ' ' || s[i-1] == '\t'):
			return s[:i]
		}
	}
	return s
}
