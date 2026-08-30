package docbundle

import (
	"fmt"
	"sort"
	"strings"
)

type markdownDocument struct {
	lines    []string
	headings []*markdownHeading
}

type markdownHeading struct {
	Level      int
	Text       string
	Anchor     string
	Raw        string
	Start      int
	BodyEnd    int
	SectionEnd int
	Parent     *markdownHeading
	Children   []*markdownHeading
}

func parseMarkdown(content string) markdownDocument {
	document := markdownDocument{lines: strings.SplitAfter(content, "\n")}
	anchors := map[string]int{}
	stack := []*markdownHeading{}
	var previous *markdownHeading
	fenceCharacter := byte(0)
	fenceLength := 0

	for index, line := range document.lines {
		value := strings.TrimRight(line, "\r\n")
		if character, length, closing := markdownFence(value, fenceCharacter, fenceLength); character != 0 {
			if closing {
				fenceCharacter = 0
				fenceLength = 0
			} else if fenceCharacter == 0 {
				fenceCharacter = character
				fenceLength = length
			}
			continue
		}
		if fenceCharacter != 0 {
			continue
		}

		match := headingPattern.FindStringSubmatch(value)
		if match == nil {
			continue
		}
		if previous != nil {
			previous.BodyEnd = index
		}
		for len(stack) > 0 && stack[len(stack)-1].Level >= len(match[1]) {
			stack[len(stack)-1].SectionEnd = index
			stack = stack[:len(stack)-1]
		}
		text := strings.TrimSpace(match[2])
		base := markdownAnchor(text)
		anchor := base
		if duplicate := anchors[base]; duplicate > 0 {
			anchor = fmt.Sprintf("%s-%d", base, duplicate)
		}
		anchors[base]++
		heading := &markdownHeading{
			Level:  len(match[1]),
			Text:   text,
			Anchor: anchor,
			Raw:    value,
			Start:  index,
		}
		if len(stack) > 0 {
			heading.Parent = stack[len(stack)-1]
			heading.Parent.Children = append(heading.Parent.Children, heading)
		}
		document.headings = append(document.headings, heading)
		stack = append(stack, heading)
		previous = heading
	}

	if previous != nil {
		previous.BodyEnd = len(document.lines)
	}
	for _, heading := range stack {
		heading.SectionEnd = len(document.lines)
	}
	return document
}

func markdownFence(line string, openCharacter byte, openLength int) (byte, int, bool) {
	trimmed := strings.TrimLeft(line, " ")
	if len(line)-len(trimmed) > 3 || trimmed == "" {
		return 0, 0, false
	}
	character := trimmed[0]
	if character != '`' && character != '~' {
		return 0, 0, false
	}
	length := 0
	for length < len(trimmed) && trimmed[length] == character {
		length++
	}
	if length < 3 {
		return 0, 0, false
	}
	if openCharacter == 0 {
		return character, length, false
	}
	if character != openCharacter || length < openLength || strings.TrimSpace(trimmed[length:]) != "" {
		return 0, 0, false
	}
	return character, length, true
}

func (document markdownDocument) heading(fragment string) (*markdownHeading, error) {
	for _, heading := range document.headings {
		if heading.Anchor == fragment {
			return heading, nil
		}
	}
	return nil, fmt.Errorf("heading not found: #%s", fragment)
}

func (document markdownDocument) section(fragment string) (string, error) {
	heading, err := document.heading(fragment)
	if err != nil {
		return "", err
	}
	return strings.Join(document.lines[heading.Start:heading.SectionEnd], ""), nil
}

func (document markdownDocument) project(fragment string, minimum, maximum int) (string, error) {
	if minimum > 0 || maximum < 0 || minimum > maximum {
		return "", fmt.Errorf("depth range must contain 0: %d..%d", minimum, maximum)
	}
	target, err := document.heading(fragment)
	if err != nil {
		return "", err
	}
	selected := map[*markdownHeading]bool{target: true}
	ancestor := target.Parent
	for depth := -1; ancestor != nil && depth >= minimum; depth-- {
		selected[ancestor] = true
		ancestor = ancestor.Parent
	}
	var addDescendants func(*markdownHeading, int)
	addDescendants = func(parent *markdownHeading, depth int) {
		if depth > maximum {
			return
		}
		for _, child := range parent.Children {
			selected[child] = true
			addDescendants(child, depth+1)
		}
	}
	addDescendants(target, 1)

	headings := make([]*markdownHeading, 0, len(selected))
	for heading := range selected {
		headings = append(headings, heading)
	}
	sort.Slice(headings, func(left, right int) bool { return headings[left].Start < headings[right].Start })
	var output strings.Builder
	for _, heading := range headings {
		output.WriteString(strings.Join(document.lines[heading.Start:heading.BodyEnd], ""))
	}
	return output.String(), nil
}

func (document markdownDocument) toc(reference string) string {
	var output strings.Builder
	for _, heading := range document.headings {
		if heading.Start == 0 && heading.Level == 1 && isCanonicalIdentity(heading.Text) {
			continue
		}
		depth := 0
		for parent := heading.Parent; parent != nil; parent = parent.Parent {
			if !(parent.Start == 0 && parent.Level == 1 && isCanonicalIdentity(parent.Text)) {
				depth++
			}
		}
		output.WriteString(strings.Repeat("  ", depth))
		label := strings.NewReplacer(`\`, `\\`, `[`, `\[`, `]`, `\]`).Replace(heading.Text)
		fmt.Fprintf(&output, "- [%s](%s#%s)\n", label, reference, heading.Anchor)
	}
	return output.String()
}
