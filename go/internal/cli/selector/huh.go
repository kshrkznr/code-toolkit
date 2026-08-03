package selector

import (
	"errors"
	"fmt"

	"charm.land/huh/v2"
)

func runHuh(title string, candidates []string) (string, error) {
	options := make([]huh.Option[string], 0, len(candidates))
	for _, candidate := range candidates {
		options = append(options, huh.NewOption(candidate, candidate))
	}

	var selected string
	err := huh.NewSelect[string]().
		Title(title).
		Options(options...).
		Value(&selected).
		Run()
	if errors.Is(err, huh.ErrUserAborted) {
		return "", ErrCancelled
	}
	if err != nil {
		return "", fmt.Errorf("select %s: %w", title, err)
	}
	return selected, nil
}
