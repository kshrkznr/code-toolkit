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
	run func(title string, candidates []string) (string, error)
}

func New() *Native {
	return &Native{run: runHuh}
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
