package parser

import (
	"strings"
	"testing"
)

func TestParseDefinitions(t *testing.T) {
	src := `
; Company report
[Report: Company Report]
    Form : Company Form
    Title : "Company; Details"

[#Report: Company Report]
    Add : Line : Company Name Line, Company State Line
`
	defs, err := ParseFile("company.tdl", src)
	if err != nil {
		t.Fatalf("ParseFile() error = %v", err)
	}
	if len(defs) != 2 {
		t.Fatalf("expected 2 definitions, got %d", len(defs))
	}

	base := defs[0]
	if base.Kind != "Report" || base.Name != "Company Report" {
		t.Fatalf("unexpected definition: %#v", base)
	}
	if base.File != "company.tdl" {
		t.Fatalf("unexpected source file: %q", base.File)
	}
	if got := base.Attrs[1].Value; got != `"Company; Details"` {
		t.Fatalf("quoted semicolon was removed: %q", got)
	}

	mod := defs[1]
	if !mod.Modify {
		t.Fatal("expected modify definition")
	}
	if len(mod.Attrs) != 1 {
		t.Fatalf("expected one modify attribute, got %d", len(mod.Attrs))
	}
	if mod.Attrs[0].Op != OpAdd {
		t.Fatalf("expected OpAdd, got %q", mod.Attrs[0].Op)
	}
	if mod.Attrs[0].Name != "Line" {
		t.Fatalf("expected target attribute Line, got %q", mod.Attrs[0].Name)
	}
}

func TestParseContinuedLine(t *testing.T) {
	src := `[Report: Test]
    Title : First part \
        second part
`
	defs, err := Parse(src)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if got, want := defs[0].Attrs[0].Value, "First part second part"; got != want {
		t.Fatalf("value = %q, want %q", got, want)
	}
}

func TestParseCollectsErrors(t *testing.T) {
	src := `
    Title : Outside
[Report: Broken
[Report: Valid]
    Form : Main
`
	defs, err := ParseFile("test.tdl", src)
	if err == nil {
		t.Fatal("expected syntax errors")
	}
	if !strings.Contains(err.Error(), "test.tdl:2:") {
		t.Fatalf("error should include file and line: %v", err)
	}
	if len(defs) != 1 || defs[0].Name != "Valid" {
		t.Fatalf("parser should continue after errors: %#v", defs)
	}
}

func TestSplitTop(t *testing.T) {
	got := SplitTop(`A, "B, C", $$Func:(D, E), F`, ',')
	want := []string{"A", `"B, C"`, "$$Func:(D, E)", "F"}

	if len(got) != len(want) {
		t.Fatalf("SplitTop() = %#v, want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("item %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestDefinitionKeyIsCaseInsensitive(t *testing.T) {
	a := &Definition{Kind: "Report", Name: "Company Report"}
	b := &Definition{Kind: " report ", Name: "COMPANY   REPORT"}

	if a.Key() != b.Key() {
		t.Fatalf("keys differ: %q != %q", a.Key(), b.Key())
	}
}
