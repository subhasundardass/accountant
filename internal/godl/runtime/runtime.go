package runtime

import (
	"fmt"
	"strings"

	"github.com/subhasundardass/accountant/internal/godl/expr"
	"github.com/subhasundardass/accountant/internal/godl/parser"
	"github.com/subhasundardass/accountant/internal/godl/registry"
)

// Runtime owns the loaded GoDL definitions and provides evaluation.
type Runtime struct {
	registry *registry.Registry
}

// New creates an empty GoDL runtime.
func New() *Runtime {
	return &Runtime{
		registry: registry.New(),
	}
}

// LoadSource parses and loads a GoDL source file.
func (r *Runtime) LoadSource(file, source string) error {
	if r == nil {
		return fmt.Errorf("GoDL runtime is nil")
	}
	return r.registry.LoadSource(file, source)
}

// LoadFiles loads multiple named source files in the supplied order.
// Later definitions can modify definitions loaded earlier.
func (r *Runtime) LoadFiles(files map[string]string, order []string) error {
	for _, name := range order {
		source, ok := files[name]
		if !ok {
			return fmt.Errorf("GoDL source %q was not provided", name)
		}
		if err := r.LoadSource(name, source); err != nil {
			return err
		}
	}
	return nil
}

// Validate checks references such as Use after all source files are loaded.
func (r *Runtime) Validate() error {
	if r == nil {
		return fmt.Errorf("GoDL runtime is nil")
	}
	return r.registry.Validate()
}

// Definition returns a loaded definition.
func (r *Runtime) Definition(kind, name string) (*parser.Definition, bool) {
	if r == nil {
		return nil, false
	}
	return r.registry.Get(kind, name)
}

// Definitions returns all definitions of a kind in load order.
// An empty kind returns all definitions.
func (r *Runtime) Definitions(kind string) []*parser.Definition {
	if r == nil {
		return nil
	}
	return r.registry.All(kind)
}

// Content resolves a definition and its Use inheritance into Content.
func (r *Runtime) Content(kind, name string) (*Content, error) {
	if r == nil {
		return nil, fmt.Errorf("GoDL runtime is nil")
	}

	def, ok := r.registry.Get(kind, name)
	if !ok {
		return nil, fmt.Errorf("GoDL definition not found: [%s: %s]", kind, name)
	}

	attrs, err := r.resolveAttributes(def, map[string]bool{})
	if err != nil {
		return nil, err
	}

	return &Content{
		Kind:       def.Kind,
		Name:       def.Name,
		Attributes: attrs,
	}, nil
}

func (r *Runtime) resolveAttributes(
	def *parser.Definition,
	seen map[string]bool,
) ([]parser.Attr, error) {
	key := def.Key()
	if seen[key] {
		return nil, fmt.Errorf("GoDL Use cycle detected at %s", def)
	}
	seen[key] = true
	defer delete(seen, key)

	var attrs []parser.Attr

	// Resolve inherited attributes first.
	if use, ok := r.registry.Attr(def, "Use"); ok {
		parent, found := r.registry.Get(def.Kind, use.Value)
		if !found {
			return nil, fmt.Errorf("%s: Use refers to missing %s %q",
				def, def.Kind, use.Value)
		}

		inherited, err := r.resolveAttributes(parent, seen)
		if err != nil {
			return nil, err
		}
		attrs = append(attrs, inherited...)
	}

	// Apply local attributes. Repeated attributes are retained in order.
	for _, attr := range def.Attrs {
		if strings.EqualFold(attr.Name, "Use") {
			continue
		}

		switch attr.Op {
		case parser.OpAdd:
			attrs = appendListAttribute(attrs, attr)
		case parser.OpDelete:
			attrs = deleteListAttribute(attrs, attr)
		default:
			attrs = setContentAttribute(attrs, attr)
		}
	}

	return attrs, nil
}

func setContentAttribute(attrs []parser.Attr, incoming parser.Attr) []parser.Attr {
	key := parser.Norm(incoming.Name)

	// Preserve repeated collection attributes in declaration order.
	switch key {
	case "line", "field", "part", "button", "key", "action", "event":
		return append(attrs, incoming)
	}

	// Ordinary attributes use last-value-wins behavior.
	for i := len(attrs) - 1; i >= 0; i-- {
		if parser.Norm(attrs[i].Name) == key {
			attrs[i] = incoming
			return attrs
		}
	}

	return append(attrs, incoming)
}

func appendListAttribute(attrs []parser.Attr, attr parser.Attr) []parser.Attr {
	key := attr.Key()

	for i := range attrs {
		if attrs[i].Key() == key {
			items := attrs[i].List()
			items = append(items, attr.List()...)
			attrs[i].Value = strings.Join(items, ", ")
			return attrs
		}
	}

	return append(attrs, attr)
}

func deleteListAttribute(attrs []parser.Attr, attr parser.Attr) []parser.Attr {
	key := attr.Key()
	drop := map[string]bool{}

	for _, item := range attr.List() {
		drop[parser.Norm(item)] = true
	}

	for i := range attrs {
		if attrs[i].Key() != key {
			continue
		}

		if len(drop) == 0 {
			attrs[i].Value = ""
			return attrs
		}

		var kept []string
		for _, item := range attrs[i].List() {
			if !drop[parser.Norm(item)] {
				kept = append(kept, item)
			}
		}
		attrs[i].Value = strings.Join(kept, ", ")
		return attrs
	}

	return attrs
}

// Eval evaluates an expression using the supplied evaluation context.
func (r *Runtime) Eval(source string, ctx *Context) (expr.Value, error) {
	if r == nil {
		return nil, fmt.Errorf("GoDL runtime is nil")
	}

	evalContext := &expressionContext{
		runtime: r,
		context: ctx,
	}
	return expr.Evaluate(source, evalContext)
}

// Context contains the current data available to GoDL expressions.
type Context struct {
	Fields    map[string]expr.Value
	Variables map[string]expr.Value

	// Functions contains application-provided GoDL functions.
	Functions map[string]Function
}

// Function is a Go function callable from a GoDL expression.
type Function func(args []expr.Value) (expr.Value, error)

// NewContext creates an empty evaluation context.
func NewContext() *Context {
	return &Context{
		Fields:    make(map[string]expr.Value),
		Variables: make(map[string]expr.Value),
		Functions: make(map[string]Function),
	}
}

// SetField assigns a field value.
func (c *Context) SetField(name string, value expr.Value) {
	if c.Fields == nil {
		c.Fields = make(map[string]expr.Value)
	}
	c.Fields[parser.Norm(name)] = value
}

// SetVariable assigns a variable value.
func (c *Context) SetVariable(name string, value expr.Value) {
	if c.Variables == nil {
		c.Variables = make(map[string]expr.Value)
	}
	c.Variables[parser.Norm(name)] = value
}

// RegisterFunction registers an application-provided expression function.
func (c *Context) RegisterFunction(name string, fn Function) {
	if c.Functions == nil {
		c.Functions = make(map[string]Function)
	}
	c.Functions[parser.Norm(name)] = fn
}

type expressionContext struct {
	runtime *Runtime
	context *Context
}

func (c *expressionContext) Field(name string) (expr.Value, bool) {
	if c.context == nil {
		return nil, false
	}
	value, ok := c.context.Fields[parser.Norm(name)]
	return value, ok
}

func (c *expressionContext) Variable(name string) (expr.Value, bool) {
	if c.context == nil {
		return nil, false
	}
	value, ok := c.context.Variables[parser.Norm(name)]
	return value, ok
}

func (c *expressionContext) Formula(name string) (string, bool) {
	return c.runtime.registry.Formula(name)
}

func (c *expressionContext) CallFunction(
	name string,
	args []expr.Value,
) (expr.Value, error) {
	if c.context == nil {
		return nil, fmt.Errorf("function $$%s is not registered", name)
	}

	fn, ok := c.context.Functions[parser.Norm(name)]
	if !ok {
		return nil, fmt.Errorf("unknown GoDL function $$%s", name)
	}
	return fn(args)
}
