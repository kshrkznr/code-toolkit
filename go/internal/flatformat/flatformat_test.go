package flatformat

import (
	"strings"
	"testing"
)

func TestEncodePreservesDottedKeysAndOrderedArrays(t *testing.T) {
	data, err := Encode(map[string]any{
		"editor.fontSize": float64(16),
		"files":           map[string]any{"exclude": []any{"one", "two"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, expected := range []string{
		`["editor.fontSize"]=16`,
		`["files", "exclude"]=[]`,
		`["files", "exclude", [0]]="one"`,
		`["files", "exclude", [1]]="two"`,
	} {
		if !strings.Contains(text, expected) {
			t.Fatalf("missing %q in:\n%s", expected, text)
		}
	}
}

func TestParseAcceptsWildcardAndRejectsDuplicate(t *testing.T) {
	assignments, err := Parse([]byte("[\"key\"]=[]\r\n[\"key\", [*]]=100\r\n[\"objects\", [@first], \"name\"]=\"one\"\r\n"))
	if err != nil || len(assignments) != 3 {
		t.Fatalf("Parse() = %#v, %v", assignments, err)
	}
	path, err := DecodePath(assignments[2].Path)
	if err != nil || len(path) != 3 || path[1].Kind != UnionNamed || path[1].Name != "first" {
		t.Fatalf("DecodePath() = %#v, %v", path, err)
	}
	if _, err := Parse([]byte("[] = {}\n[] = {}\n")); err == nil {
		t.Fatal("expected duplicate path error")
	}
}

func TestEncodeDoesNotHTMLEscapeWorkbenchSymbols(t *testing.T) {
	data, err := Encode(map[string]any{
		"vim.handleKeys": map[string]any{"<C-d>": false, "<C-w>": false},
		"label":          "one & two",
	})
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, expected := range []string{
		`["vim.handleKeys", "<C-d>"]=false`,
		`["vim.handleKeys", "<C-w>"]=false`,
		`["label"]="one & two"`,
	} {
		if !strings.Contains(text, expected) {
			t.Fatalf("missing %q in:\n%s", expected, text)
		}
	}
	for _, escaped := range []string{`\u003c`, `\u003e`, `\u0026`} {
		if strings.Contains(text, escaped) {
			t.Fatalf("Workbench output contains HTML escape %q:\n%s", escaped, text)
		}
	}
}
