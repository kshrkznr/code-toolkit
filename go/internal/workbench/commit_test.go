package workbench

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kshrkznr/code-toolkit/go/internal/mergerules"
	"github.com/kshrkznr/code-toolkit/go/internal/settings"
)

func TestCommitAllowsPartialSettingsAndWritesUnionRule(t *testing.T) {
	root := t.TempDir()
	draft := filepath.Join(root, "draft")
	ingredient := filepath.Join(root, "ingredient")
	if err := os.MkdirAll(draft, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(ingredient, 0o755); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(ingredient, "runtime.test.settings.json")
	if err := os.WriteFile(target, []byte(`{"old":"value","items":[0,9]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	draftData := `# Settings Draft

## Difference

### runtime.test.settings.json

#### items

` + "```diff\n" + `- ["old"]="wrong"
- ["items", [0]]=0
- ["items", [1]]=9
- ["items"]=[]
+ ["items"]=[]
+ ["items", [*]]=1
+ ["items", [*]]=2
+ ["objects"]=[]
+ ["objects", [@first]]={}
+ ["objects", [@first], "name"]="one"
` + "```\n"
	if err := os.WriteFile(filepath.Join(draft, "settings.draft.md"), []byte(draftData), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := (Service{CookbookRoot: root}).Commit(false)
	if err != nil {
		t.Fatal(err)
	}
	if result.Unresolved != 1 {
		t.Fatalf("result = %#v", result)
	}
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	document, err := settings.Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	items := document["items"].([]any)
	if len(items) != 2 || items[0].(float64) != 1 || items[1].(float64) != 2 {
		t.Fatalf("items = %#v", items)
	}
	objects := document["objects"].([]any)
	if len(objects) != 1 || objects[0].(map[string]any)["name"] != "one" {
		t.Fatalf("objects = %#v", objects)
	}
	if document["old"] != "value" {
		t.Fatalf("mismatched removal changed document: %#v", document)
	}
	rules, err := mergerules.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if !rules.Union[mergerules.Key([]string{"items"})] || !rules.Union[mergerules.Key([]string{"objects"})] {
		t.Fatalf("rules = %#v", rules)
	}
	if _, err := os.Stat(filepath.Join(draft, "settings.draft.md")); err != nil {
		t.Fatal("Draft must be retained")
	}
}

func TestCommitAllowsExtensionsWithoutRecipeDraft(t *testing.T) {
	root := t.TempDir()
	draft := filepath.Join(root, "draft")
	ingredient := filepath.Join(root, "ingredient")
	for _, path := range []string{draft, ingredient} {
		if err := os.MkdirAll(path, 0755); err != nil {
			t.Fatal(err)
		}
	}
	content := "# Extensions Draft\n\n## Difference\n\n### runtime.test.extensions\n\n```diff\n- old.id\n+ new.id\n```\n"
	if err := os.WriteFile(filepath.Join(draft, "extensions.draft.md"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ingredient, "runtime.test.extensions"), []byte("old.id\nkeep.id\n"), 0644); err != nil {
		t.Fatal(err)
	}
	result, err := (Service{CookbookRoot: root}).Commit(false)
	if err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(filepath.Join(ingredient, "runtime.test.extensions"))
	text := string(data)
	if strings.Contains(text, "old.id") || !strings.Contains(text, "new.id") || !strings.Contains(text, "keep.id") {
		t.Fatalf("extensions = %q", text)
	}
	if len(result.Files) != 1 {
		t.Fatalf("files = %#v", result.Files)
	}
}

func TestCommitRejectsCommentedJSONCWithoutForce(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, "draft"), 0755)
	os.MkdirAll(filepath.Join(root, "ingredient"), 0755)
	os.WriteFile(filepath.Join(root, "ingredient", "runtime.test.settings.jsonc"), []byte("{/* keep */\"a\":1}"), 0644)
	os.WriteFile(filepath.Join(root, "draft", "settings.draft.md"), []byte("## Difference\n### runtime.test.settings.jsonc\n```diff\n+ [\"b\"]=2\n```\n"), 0644)
	if _, err := (Service{CookbookRoot: root}).Commit(false); err == nil {
		t.Fatal("expected comment preservation gate")
	}
	if _, err := (Service{CookbookRoot: root}).Commit(true); err != nil {
		t.Fatal(err)
	}
}

func TestCommitRuntimeArtifactDocuments(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, "draft"), 0755)
	os.MkdirAll(filepath.Join(root, "ingredient"), 0755)
	draft := "# Keybindings Draft\n\n## Difference\n\n### runtime.test.keybindings.json\n\n```diff\n+ [\n+   {\"key\":\"ctrl+x\",\"command\":\"probe\"}\n+ ]\n```\n"
	if err := os.WriteFile(filepath.Join(root, "draft", "keybindings.draft.md"), []byte(draft), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := (Service{CookbookRoot: root}).Commit(false); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, "ingredient", "runtime.test.keybindings.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "probe") {
		t.Fatalf("keybindings = %s", data)
	}
}
