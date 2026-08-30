package selector

import (
	"errors"
	"fmt"
	"strings"

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

func runHuhInput(title, initial string) (string, error) {
	value := initial
	field := huh.NewInput().
		Title(title).
		Value(&value).
		Validate(func(value string) error {
			if strings.TrimSpace(value) == "" {
				return fmt.Errorf("a value is required")
			}
			return nil
		})
	keymap := huh.NewDefaultKeyMap()
	keymap.Quit.SetKeys("esc", "ctrl+c")
	err := huh.NewForm(huh.NewGroup(field)).
		WithKeyMap(keymap).
		WithAccessible(false).
		WithShowHelp(false).
		Run()
	if errors.Is(err, huh.ErrUserAborted) {
		return "", ErrCancelled
	}
	if err != nil {
		return "", fmt.Errorf("input %s: %w", title, err)
	}
	return value, nil
}
