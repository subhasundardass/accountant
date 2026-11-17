package parser

import (
	"fmt"
	"strings"
)

// Op is the modification operator of an attribute inside a [#Kind: Name] block.
type Op string

const (
	OpSet     Op = ""        // plain "Attr : value" (a bare override in a modify block)
	OpAdd     Op = "add"     // "Add : Attr : items"     append items to a list attribute
	OpDelete  Op = "delete"  // "Delete : Attr : items"  remove items from a list attribute
	OpReplace Op = "replace" // "Replace : Attr : value" replace the whole value
)

// Attr is one "Name : value" line of a definition.
type Attr struct {
	Name  string // as written, e.g. "Set as"
	Value string // raw text after the first top-level ':'; expressions stay unparsed
	Op    Op
	Line  int
}

// Key is the normalised attribute name used for lookups.
func (a Attr) Key() string { return Norm(a.Name) }

// List splits the value on top-level commas: "A, B" -> ["A","B"].
func (a Attr) List() []string { return SplitTop(a.Value, ',') }

// Parts splits the value on top-level colons: "Line : Coll" -> ["Line","Coll"].
// Only use this for structural attributes (Repeat, ...), never for
// expressions, where ':' is the $$Func argument separator.
func (a Attr) Parts() []string { return SplitTop(a.Value, ':') }

// Definition is one "[Kind: Name]" block.
type Definition struct {
	Kind   string
	Name   string
	Modify bool // written as [#Kind: Name]
	Attrs  []Attr
	File   string
	Line   int
}

// Key is the registry key: normalised "kind/name".
func (d *Definition) Key() string { return Key(d.Kind, d.Name) }

func (d *Definition) String() string {
	hash := ""
	if d.Modify {
		hash = "#"
	}
	return fmt.Sprintf("[%s%s: %s]", hash, d.Kind, d.Name)
}

// Clone returns a copy that shares no slices with d.
func (d *Definition) Clone() *Definition {
	c := *d
	c.Attrs = append([]Attr(nil), d.Attrs...)
	return &c
}

// Norm lower-cases and collapses whitespace. TDL names are case-insensitive.
func Norm(s string) string {
	return strings.ToLower(strings.Join(strings.Fields(s), " "))
}

// Key builds the registry key for a kind and name.
func Key(kind, name string) string { return Norm(kind) + "/" + Norm(name) }

// SplitTop splits s on sep, ignoring separators inside "double quotes"
// (a doubled "" is an escaped quote) or (parentheses). Parts are trimmed and
// empty parts are dropped.
func SplitTop(s string, sep rune) []string {
	var out []string
	var b strings.Builder
	inQuote, depth := false, 0
	flush := func() {
		if p := strings.TrimSpace(b.String()); p != "" {
			out = append(out, p)
		}
		b.Reset()
	}
	for _, r := range s {
		switch {
		case r == '"':
			inQuote = !inQuote
			b.WriteRune(r)
		case inQuote:
			b.WriteRune(r)
		case r == '(':
			depth++
			b.WriteRune(r)
		case r == ')':
			if depth > 0 {
				depth--
			}
			b.WriteRune(r)
		case r == sep && depth == 0:
			flush()
		default:
			b.WriteRune(r)
		}
	}
	flush()
	return out
}

// indexTop returns the byte index of the first sep outside quotes, or -1.
func indexTop(s string, sep byte) int {
	inQuote := false
	for i := 0; i < len(s); i++ {
		switch {
		case s[i] == '"':
			inQuote = !inQuote
		case s[i] == sep && !inQuote:
			return i
		}
	}
	return -1
}
