package repository

import "testing"

func TestBuiltInRepositoryConnectors(t *testing.T) {
	want := map[string]Connector{
		"visual-studio-marketplace": ConnectorVisualStudioMarketplace,
		"open-vsx":                  ConnectorOpenVSX,
		"cursor-marketplace":        ConnectorCursorMarketplace,
		"windsurf-marketplace":      ConnectorWindsurfMarketplace,
	}
	for id, connector := range want {
		definition, err := Lookup(id)
		if err != nil {
			t.Fatalf("Lookup(%q): %v", id, err)
		}
		if definition.ID != id || definition.Connector != connector {
			t.Errorf("Lookup(%q) = %#v", id, definition)
		}
	}
}

func TestLookupRejectsUnknownRepository(t *testing.T) {
	if _, err := Lookup("unknown"); err == nil {
		t.Fatal("expected unknown Repository error")
	}
}
