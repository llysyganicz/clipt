// Package frontmatter defines what front matter is for notes and templates:
// a YAML block between "---" fences at the very top of a file. It parses the
// "tags" key and merges tags into a file's text line by line, leaving
// everything it does not own exactly as written.
package frontmatter

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/goccy/go-yaml"
)

// Fence is the line that opens and closes front matter.
const Fence = "---"

// ErrUnclosed reports front matter whose closing fence is missing.
var ErrUnclosed = errors.New("front matter is not closed: no closing --- line")

// Block is the front matter of a file. Lines are the YAML lines between the
// fences; Start and End are the line indexes of the opening and closing
// fence (End is -1 when the block is not closed).
type Block struct {
	Present bool
	Lines   []string
	Start   int
	End     int
}

// YAML returns the block's YAML text.
func (b Block) YAML() string { return strings.Join(b.Lines, "\n") }

// Split splits content into its front matter and body. Windows line endings
// are normalized first. Front matter is present when the first line is
// exactly "---" and ends at the next such line; the body is everything after
// the closing line. Without front matter the body is the whole content. When
// the block is not closed the body is the whole content and the error is
// ErrUnclosed.
func Split(content string) (Block, string, error) {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	lines := strings.Split(content, "\n")
	if lines[0] != Fence {
		return Block{End: -1}, content, nil
	}
	for i := 1; i < len(lines); i++ {
		if lines[i] == Fence {
			return Block{Present: true, Lines: lines[1:i], Start: 0, End: i}, strings.Join(lines[i+1:], "\n"), nil
		}
	}
	return Block{Present: true, Lines: lines[1:], Start: 0, End: -1}, content, ErrUnclosed
}

// Tags parses the "tags" key of the YAML front matter text raw. The value
// must be a sequence of scalars; each tag is a value's text with surrounding
// whitespace removed. Empty values and values equal to an earlier one,
// compared case-insensitively, are dropped. A missing key or an empty list
// gives no tags; other keys are ignored.
func Tags(raw string) ([]string, error) {
	var shape struct {
		Tags any `yaml:"tags"`
	}
	if err := yaml.Unmarshal([]byte(raw), &shape); err != nil {
		return nil, err
	}
	if shape.Tags == nil {
		return nil, nil
	}
	items, ok := shape.Tags.([]any)
	if !ok {
		return nil, errors.New("tags must be a list")
	}
	for i, it := range items {
		switch it.(type) {
		case map[string]any, []any:
			return nil, fmt.Errorf("tags[%d] must be a plain value", i)
		}
	}
	var text struct {
		Tags []string `yaml:"tags"`
	}
	if err := yaml.Unmarshal([]byte(raw), &text); err != nil {
		return nil, err
	}
	return Clean(text.Tags), nil
}

// Clean trims every tag and drops empty tags and tags equal to an earlier
// one, compared case-insensitively, keeping order and case.
func Clean(tags []string) []string {
	var out []string
	seen := make(map[string]bool, len(tags))
	for _, t := range tags {
		t = strings.TrimSpace(t)
		k := strings.ToLower(t)
		if t == "" || seen[k] {
			continue
		}
		seen[k] = true
		out = append(out, t)
	}
	return out
}

// Parse returns the tags of a file's content: none without front matter, an
// error when the front matter is invalid (not closed, not YAML, or tags not
// a sequence of scalars).
func Parse(content string) ([]string, error) {
	fm, _, err := Split(content)
	if err != nil {
		return nil, err
	}
	if !fm.Present {
		return nil, nil
	}
	return Tags(fm.YAML())
}

// Flow returns tags as a YAML flow sequence, quoting values that need it,
// so that it parses back to the same strings.
func Flow(tags []string) string {
	if len(tags) == 0 {
		return "[]"
	}
	b, err := yaml.MarshalWithOptions(tags, yaml.Flow(true))
	if err != nil {
		// A []string always marshals; keep a valid fallback anyway.
		return "[]"
	}
	return strings.TrimRight(string(b), "\n")
}

// Line returns the front matter line for tags in flow form.
func Line(tags []string) string { return "tags: " + Flow(tags) }

// Display returns tags as shown in the UI: each as "#tag", space-separated.
func Display(tags []string) string {
	parts := make([]string, len(tags))
	for i, t := range tags {
		parts[i] = "#" + t
	}
	return strings.Join(parts, " ")
}

// MergeTags adds to content every tag of add it does not already have,
// compared case-insensitively, after the existing tags. Only front matter
// lines change: without front matter a block holding just the tags is
// prepended; without a "tags" key one is added as the last front matter
// line; otherwise the lines of the top-level "tags" entry are replaced by
// one flow line. delta is the number of lines added (or removed, when
// negative). When nothing needs adding the content is returned unchanged.
// Invalid front matter, or a result that does not parse back to the merged
// tags, refuses the merge with an error and returns content unchanged.
func MergeTags(content string, add []string) (merged string, delta int, err error) {
	fm, _, err := Split(content)
	if err != nil {
		return content, 0, err
	}
	var have []string
	if fm.Present {
		if have, err = Tags(fm.YAML()); err != nil {
			return content, 0, err
		}
	}
	all := Clean(append(slices.Clone(have), add...))
	if len(all) == len(have) {
		return content, 0, nil
	}
	line := Line(all)

	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	var out []string
	switch {
	case !fm.Present:
		out = append([]string{Fence, line, Fence}, lines...)
	default:
		start := -1
		for i := 1; i < fm.End; i++ {
			if strings.HasPrefix(lines[i], "tags:") {
				start = i
				break
			}
		}
		if start < 0 {
			out = slices.Concat(lines[:fm.End], []string{line}, lines[fm.End:])
			break
		}
		end := start + 1
		for end < fm.End && belongsToEntry(lines[end]) {
			end++
		}
		out = slices.Concat(lines[:start], []string{line}, lines[end:])
	}
	merged = strings.Join(out, "\n")
	if got, err := Parse(merged); err != nil || !slices.Equal(got, all) {
		return content, 0, errors.New("cannot merge tags: unsupported tags entry")
	}
	return merged, len(out) - len(lines), nil
}

// belongsToEntry reports whether a front matter line continues the entry
// above it: indented, or a block sequence item at column 0.
func belongsToEntry(line string) bool {
	return strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") ||
		strings.HasPrefix(line, "- ") || line == "-"
}
