// Package registry stores parsed definitions and resolves them the way TDL
// does: case-insensitive names, [#Kind: Name] modification of earlier
// definitions, and "Use : Other" attribute inheritance.
package registry

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/subhasundardass/accountant/internal/godl/parser"
)

// maxUseDepth bounds "Use" chains; a longer chain is treated as a cycle.
const maxUseDepth = 32

// Registry is safe for concurrent use. Stored definitions are immutable:
// a modification replaces the entry with a new copy, so a *Definition you
// obtained earlier never changes underneath you (safe for hot reload).
type Registry struct {
	mu    sync.RWMutex
	defs  map[string]*parser.Definition
	order []string // keys in first-load order
}

// New returns an empty registry.
func New() *Registry {
	return &Registry{defs: map[string]*parser.Definition{}}
}

// LoadSource parses src and loads the result. Syntax errors and load errors
// are returned together; everything that was valid is still loaded.
func (r *Registry) LoadSource(file, src string) error {
	defs, perr := parser.ParseFile(file, src)
	lerr := r.Load(defs)
	return errors.Join(perr, lerr)
}

// Load adds definitions in order, as TDL loads files:
//   - [Kind: Name]  creates the definition, or replaces an earlier one of the
//     same kind and name.
//   - [#Kind: Name] modifies the existing definition (see applyModify); it is
//     an error if the definition does not exist yet.
//
// All errors are collected and returned joined.
func (r *Registry) Load(defs []*parser.Definition) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var errs []error
	for _, d := range defs {
		key := d.Key()
		if !d.Modify {
			c := d.Clone()
			// A plain definition may not use Add/Delete/Replace; the parser already
			// rejects that, so attrs here are plain sets.
			if _, exists := r.defs[key]; !exists {
				r.order = append(r.order, key)
			}
			r.defs[key] = c
			continue
		}
		base, ok := r.defs[key]
		if !ok {
			errs = append(errs, locErr(d, "modify of undefined %s (load the base definition first)", d))
			continue
		}
		next, err := r.applyModify(base, d)
		r.defs[key] = next // next is valid even if err != nil
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

// applyModify returns a new definition: base with mod's attributes applied.
//
//	Attr : v            override (same as Replace)
//	Replace : Attr : v  override
//	Add : Attr : a, b   append items to the (possibly inherited) list value
//	Delete : Attr : a   remove items from the (possibly inherited) list value;
//	                    with no items listed, the attribute is cleared
func (r *Registry) applyModify(base, mod *parser.Definition) (*parser.Definition, error) {
	out := base.Clone()
	var errs []error
	for _, a := range mod.Attrs {
		switch a.Op {
		case parser.OpAdd, parser.OpDelete:
			cur, _ := r.attr(out, a.Key(), 0)
			items := parser.SplitTop(cur.Value, ',')
			if a.Op == parser.OpAdd {
				items = append(items, a.List()...)
			} else if len(a.List()) == 0 {
				items = nil
			} else {
				items = removeItems(items, a.List())
			}
			setAttr(out, parser.Attr{Name: a.Name, Value: strings.Join(items, ", "), Line: a.Line})
		default: // OpSet, OpReplace
			setAttr(out, parser.Attr{Name: a.Name, Value: a.Value, Line: a.Line})
		}
	}
	return out, errors.Join(errs...)
}

func removeItems(items, drop []string) []string {
	var kept []string
outer:
	for _, it := range items {
		for _, d := range drop {
			if parser.Norm(it) == parser.Norm(d) {
				continue outer
			}
		}
		kept = append(kept, it)
	}
	return kept
}

// setAttr overrides an existing local attribute in place, or appends one.
func setAttr(d *parser.Definition, a parser.Attr) {
	for i := range d.Attrs {
		if d.Attrs[i].Key() == a.Key() {
			d.Attrs[i] = a
			return
		}
	}
	d.Attrs = append(d.Attrs, a)
}

// Get returns the definition of the given kind and name.
func (r *Registry) Get(kind, name string) (*parser.Definition, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	d, ok := r.defs[parser.Key(kind, name)]
	return d, ok
}

// All returns every definition of a kind in load order. An empty kind
// returns everything.
func (r *Registry) All(kind string) []*parser.Definition {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []*parser.Definition
	for _, k := range r.order {
		d := r.defs[k]
		if kind == "" || parser.Norm(d.Kind) == parser.Norm(kind) {
			out = append(out, d)
		}
	}
	return out
}

// Attr resolves an attribute of d: the definition's own value if it has one
// (the last occurrence wins), otherwise the value inherited through "Use".
func (r *Registry) Attr(d *parser.Definition, name string) (parser.Attr, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.attr(d, parser.Norm(name), 0)
}

// Value is Attr returning only the raw value ("" when absent).
func (r *Registry) Value(d *parser.Definition, name string) string {
	a, _ := r.Attr(d, name)
	return a.Value
}

// List is Attr split on top-level commas.
func (r *Registry) List(d *parser.Definition, name string) []string {
	a, _ := r.Attr(d, name)
	return a.List()
}

// attr is Attr without locking; key must already be normalised.
func (r *Registry) attr(d *parser.Definition, key string, depth int) (parser.Attr, bool) {
	for i := len(d.Attrs) - 1; i >= 0; i-- {
		if d.Attrs[i].Key() == key {
			return d.Attrs[i], true
		}
	}
	if key == "use" || depth >= maxUseDepth {
		return parser.Attr{}, false
	}
	for i := len(d.Attrs) - 1; i >= 0; i-- {
		if d.Attrs[i].Key() == "use" {
			if parent, ok := r.defs[parser.Key(d.Kind, d.Attrs[i].Value)]; ok {
				return r.attr(parent, key, depth+1)
			}
			break
		}
	}
	return parser.Attr{}, false
}

// Formula returns the raw expression of a named formula, declared as
//
//	[System: Formula]
//	    MyFormula : $Name + " (" + $Parent + ")"
//
// and referenced from expressions as @MyFormula.
func (r *Registry) Formula(name string) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	key := parser.Norm(name)
	d, ok := r.defs[parser.Key("System", "Formula")]
	if !ok {
		return "", false
	}
	for i := len(d.Attrs) - 1; i >= 0; i-- {
		if d.Attrs[i].Key() == key {
			return d.Attrs[i].Value, true
		}
	}
	return "", false
}

// Validate checks every "Use" reference: the target must exist (same kind)
// and the chain must not loop. It returns all problems, joined.
func (r *Registry) Validate() error {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var errs []error
	for _, k := range r.order {
		d := r.defs[k]
		seen := map[string]bool{d.Key(): true}
		cur := d
		for {
			u, ok := localAttr(cur, "use")
			if !ok {
				break
			}
			next, found := r.defs[parser.Key(cur.Kind, u.Value)]
			if !found {
				errs = append(errs, locErr(cur, "%s: Use : %s refers to an undefined %s", cur, u.Value, cur.Kind))
				break
			}
			if seen[next.Key()] {
				errs = append(errs, locErr(d, "%s: Use chain loops back to %s", d, next))
				break
			}
			seen[next.Key()] = true
			cur = next
		}
	}
	return errors.Join(errs...)
}

func localAttr(d *parser.Definition, key string) (parser.Attr, bool) {
	for i := len(d.Attrs) - 1; i >= 0; i-- {
		if d.Attrs[i].Key() == key {
			return d.Attrs[i], true
		}
	}
	return parser.Attr{}, false
}

func locErr(d *parser.Definition, format string, a ...any) error {
	return &parser.Error{File: d.File, Line: d.Line, Msg: fmt.Sprintf(format, a...)}
}
