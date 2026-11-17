package runtime

import (
	"testing"

	"github.com/subhasundardass/accountant/internal/godl/expr"
)

func TestLoadSource(t *testing.T) {
	rt := New()

	source := `
[Style: Default Style]
    Type : Text
    Color : Blue

[Field: Company Name]
    Type : Text
    Label : Company Name
    Required : Yes

[Field: Company GSTIN]
    Use : Company Name
    Label : GSTIN
`

	if err := rt.LoadSource("company.tdl", source); err != nil {
		t.Fatalf("LoadSource() error = %v", err)
	}

	def, ok := rt.Definition("Field", "Company Name")
	if !ok {
		t.Fatal("Company Name definition not found")
	}

	if def.Name != "Company Name" {
		t.Errorf("definition name = %q, want %q", def.Name, "Company Name")
	}
}

func TestContent(t *testing.T) {
	rt := New()

	source := `
[Field: Company Name]
    Type : Text
    Label : Company Name
    Required : Yes
`

	if err := rt.LoadSource("company.tdl", source); err != nil {
		t.Fatal(err)
	}

	content, err := rt.Content("Field", "Company Name")
	if err != nil {
		t.Fatalf("Content() error = %v", err)
	}

	if content.Kind != "Field" {
		t.Errorf("Kind = %q, want Field", content.Kind)
	}

	if content.Name != "Company Name" {
		t.Errorf("Name = %q, want Company Name", content.Name)
	}

	if got := content.Value("Type"); got != "Text" {
		t.Errorf("Type = %q, want Text", got)
	}

	if got := content.Value("Label"); got != "Company Name" {
		t.Errorf("Label = %q, want Company Name", got)
	}

	if got := content.Value("Required"); got != "Yes" {
		t.Errorf("Required = %q, want Yes", got)
	}
}

func TestContentInheritance(t *testing.T) {
	rt := New()

	source := `
[Field: Base Field]
    Type : Text
    Required : Yes
    Color : Blue

[Field: Company Name]
    Use : Base Field
    Label : Company Name
    Color : Green
`

	if err := rt.LoadSource("inheritance.tdl", source); err != nil {
		t.Fatal(err)
	}

	content, err := rt.Content("Field", "Company Name")
	if err != nil {
		t.Fatalf("Content() error = %v", err)
	}

	tests := []struct {
		name string
		want string
	}{
		{"Type", "Text"},
		{"Required", "Yes"},
		{"Label", "Company Name"},
		{"Color", "Green"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := content.Value(tt.name); got != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}

func TestContentList(t *testing.T) {
	rt := New()

	source := `
[Part: Company Details]
    Line : Company Name Line
    Line : Company State Line
    Line : Company GSTIN Line
`

	if err := rt.LoadSource("company.tdl", source); err != nil {
		t.Fatal(err)
	}

	content, err := rt.Content("Part", "Company Details")
	if err != nil {
		t.Fatal(err)
	}

	got := content.List("Line")

	want := []string{
		"Company Name Line",
		"Company State Line",
		"Company GSTIN Line",
	}

	if len(got) != len(want) {
		t.Fatalf("List() length = %d, want %d; got %v", len(got), len(want), got)
	}

	for i := range want {
		if got[i] != want[i] {
			t.Errorf("List()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestEvalFields(t *testing.T) {
	rt := New()
	ctx := NewContext()

	ctx.SetField("Amount", 100)
	ctx.SetField("Tax", 18)

	got, err := rt.Eval("$Amount + $Tax", ctx)
	if err != nil {
		t.Fatalf("Eval() error = %v", err)
	}

	if value, ok := expr.ToNumber(got); !ok || value != 118 {
		t.Errorf("Eval() = %v, want 118", got)
	}
}

func TestEvalVariables(t *testing.T) {
	rt := New()
	ctx := NewContext()

	ctx.SetVariable("Discount", 10)

	got, err := rt.Eval("##Discount + 5", ctx)
	if err != nil {
		t.Fatalf("Eval() error = %v", err)
	}

	if value, ok := expr.ToNumber(got); !ok || value != 15 {
		t.Errorf("Eval() = %v, want 15", got)
	}
}

func TestEvalFunction(t *testing.T) {
	rt := New()
	ctx := NewContext()

	ctx.RegisterFunction("Double", func(args []expr.Value) (expr.Value, error) {
		if len(args) != 1 {
			t.Fatalf("Double() received %d arguments, want 1", len(args))
		}

		value, ok := expr.ToNumber(args[0])
		if !ok {
			t.Fatalf("Double() argument is not numeric: %v", args[0])
		}

		return value * 2, nil
	})

	got, err := rt.Eval("$$Double(25)", ctx)
	if err != nil {
		t.Fatalf("Eval() error = %v", err)
	}

	if value, ok := expr.ToNumber(got); !ok || value != 50 {
		t.Errorf("Eval() = %v, want 50", got)
	}
}

func TestEvalFormula(t *testing.T) {
	rt := New()

	source := `
[System: Formula]
    FullName : $FirstName + " " + $LastName
`

	if err := rt.LoadSource("formula.tdl", source); err != nil {
		t.Fatal(err)
	}

	ctx := NewContext()
	ctx.SetField("FirstName", "Subha")
	ctx.SetField("LastName", "Das")

	got, err := rt.Eval("@FullName", ctx)
	if err != nil {
		t.Fatalf("Eval() error = %v", err)
	}

	if got != "Subha Das" {
		t.Errorf("Eval() = %v, want %q", got, "Subha Das")
	}
}

func TestDefinitionNotFound(t *testing.T) {
	rt := New()

	_, err := rt.Content("Field", "Missing Field")
	if err == nil {
		t.Fatal("Content() expected an error for missing definition")
	}
}
