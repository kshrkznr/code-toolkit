package selector

import (
	"errors"
	"fmt"
	"sort"
)

var (
	ErrCancelled    = errors.New("selection cancelled")
	ErrNoCandidates = errors.New("no selectable candidates")
)

type Selector interface {
	Select(title string, candidates []string) (string, error)
}

type Native struct {
	run  func(title string, candidates []string) (string, error)
	read func(title, initial string) (string, error)
}

func New() *Native {
	return &Native{run: runHuh, read: runHuhInput}
}

func (s *Native) Select(title string, candidates []string) (string, error) {
	values := append([]string(nil), candidates...)
	sort.Strings(values)
	values = compact(values)

	switch len(values) {
	case 0:
		return "", fmt.Errorf("%w: %s", ErrNoCandidates, title)
	case 1:
		return values[0], nil
	default:
		return s.run(title, values)
	}
}

// Input reads one editable value. initial is accepted when the user submits
// without changing it; Escape returns ErrCancelled.
func (s *Native) Input(title, initial string) (string, error) {
	read := s.read
	if read == nil {
		read = runHuhInput
	}
	return read(title, initial)
}

func compact(values []string) []string {
	result := values[:0]
	for _, value := range values {
		if value == "" || len(result) > 0 && result[len(result)-1] == value {
			continue
		}
		result = append(result, value)
	}
	return result
}
