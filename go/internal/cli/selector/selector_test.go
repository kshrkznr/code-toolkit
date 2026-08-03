package selector

import (
	"errors"
	"reflect"
	"testing"
)

func TestSelectCandidateRules(t *testing.T) {
	t.Run("none", func(t *testing.T) {
		s := &Native{run: func(string, []string) (string, error) {
			t.Fatal("interactive selector called")
			return "", nil
		}}
		if _, err := s.Select("Distribution", nil); !errors.Is(err, ErrNoCandidates) {
			t.Fatalf("Select() error = %v, want ErrNoCandidates", err)
		}
	})

	t.Run("one", func(t *testing.T) {
		s := &Native{run: func(string, []string) (string, error) {
			t.Fatal("interactive selector called")
			return "", nil
		}}
		got, err := s.Select("Distribution", []string{"vscode-golang"})
		if err != nil {
			t.Fatal(err)
		}
		if got != "vscode-golang" {
			t.Fatalf("Select() = %q", got)
		}
	})

	t.Run("many", func(t *testing.T) {
		var gotCandidates []string
		s := &Native{run: func(_ string, candidates []string) (string, error) {
			gotCandidates = append([]string(nil), candidates...)
			return candidates[1], nil
		}}
		got, err := s.Select("Distribution", []string{"z", "a", "z", ""})
		if err != nil {
			t.Fatal(err)
		}
		if got != "z" {
			t.Fatalf("Select() = %q, want z", got)
		}
		if !reflect.DeepEqual(gotCandidates, []string{"a", "z"}) {
			t.Fatalf("candidates = %v", gotCandidates)
		}
	})
}
