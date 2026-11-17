package runtime

import (
	"github.com/subhasundardass/accountant/internal/godl/parser"
)

// Content is a resolved GoDL definition.
//
// Attributes preserve their original order. The runtime does not interpret
// UI-specific attributes here; the renderer or application can interpret them.
type Content struct {
	Kind       string
	Name       string
	Attributes []parser.Attr
}

// Attribute returns the last matching attribute.
func (c *Content) Attribute(name string) (parser.Attr, bool) {
	if c == nil {
		return parser.Attr{}, false
	}

	key := parser.Norm(name)

	for i := len(c.Attributes) - 1; i >= 0; i-- {
		if c.Attributes[i].Key() == key {
			return c.Attributes[i], true
		}
	}

	return parser.Attr{}, false
}

// Value returns the raw value of the last matching attribute.
func (c *Content) Value(name string) string {
	attr, ok := c.Attribute(name)
	if !ok {
		return ""
	}

	return attr.Value
}

// List returns values from all matching attributes,
// preserving their declaration order.
func (c *Content) List(name string) []string {
	if c == nil {
		return nil
	}

	key := parser.Norm(name)
	var values []string

	for _, attr := range c.Attributes {
		if attr.Key() != key {
			continue
		}

		values = append(values, attr.List()...)
	}

	return values
}
