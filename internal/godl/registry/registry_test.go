package registry

import (
	"testing"

	"github.com/subhasundardass/accountant/internal/godl/parser"
)

func TestLoadAndGet(t *testing.T) {
	r := New()

	err := r.LoadSource("base.tdl", `
[Report: Company Report]
    Form : Company Form
    Title : Company Details
`)
	if err != nil {
		t.Fatal(err)
	}

	d, ok := r.Get("report", "company report")
	if !ok {
		t.Fatal("definition not found")
	}
	if got := r.Value(d, "Form"); got != "Company Form" {
		t.Fatalf("Form = %q", got)
	}
}

func TestModifyAddAndDelete(t *testing.T) {
	r := New()

	err := r.LoadSource("test.tdl", `
[Report: Company Report]
    Line : Name Line, State Line

[#Report: Company Report]
    Add : Line : GSTIN Line
    Delete : Line : State Line
`)
	if err != nil {
		t.Fatal(err)
	}

	d, ok := r.Get("Report", "Company Report")
	if !ok {
		t.Fatal("definition not found")
	}

	got := r.List(d, "Line")
	want := []string{"Name Line", "GSTIN Line"}

	if len(got) != len(want) {
		t.Fatalf("Line = %#v, want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Line[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestUseInheritance(t *testing.T) {
	r := New()

	err := r.LoadSource("inheritance.tdl", `
[Form: Base Form]
    Width : 80
    Height : 25

[Form: Company Form]
    Use : Base Form
    Width : 100
`)
	if err != nil {
		t.Fatal(err)
	}

	d, _ := r.Get("Form", "Company Form")

	if got := r.Value(d, "Width"); got != "100" {
		t.Fatalf("Width = %q, want 100", got)
	}
	if got := r.Value(d, "Height"); got != "25" {
		t.Fatalf("Height = %q, want 25", got)
	}
}

func TestValidateUndefinedUse(t *testing.T) {
	r := New()

	err := r.LoadSource("bad.tdl", `
[Form: Company Form]
    Use : Missing Form
`)
	if err != nil {
		t.Fatal(err)
	}

	if err := r.Validate(); err == nil {
		t.Fatal("expected undefined Use error")
	}
}

func TestAllPreservesLoadOrder(t *testing.T) {
	r := New()

	defs := []*parser.Definition{
		{Kind: "Form", Name: "First"},
		{Kind: "Form", Name: "Second"},
		{Kind: "Report", Name: "Third"},
	}
	if err := r.Load(defs); err != nil {
		t.Fatal(err)
	}

	all := r.All("")
	if len(all) != 3 {
		t.Fatalf("expected 3 definitions, got %d", len(all))
	}

	if all[0].Name != "First" ||
		all[1].Name != "Second" ||
		all[2].Name != "Third" {
		t.Fatalf("unexpected load order: %#v", all)
	}
}
