// Package templates defines the template file format and where templates
// live: markdown files under the notes directory's .templates subdirectory,
// with optional YAML front matter declaring a small form whose values fill
// {{name}} placeholders in the body.
package templates

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"github.com/goccy/go-yaml"

	"github.com/llysyganicz/clipt/internal/frontmatter"
	"github.com/llysyganicz/clipt/internal/notes"
)

// dirName is the subdirectory of the notes directory that holds templates.
const dirName = ".templates"

// Skeleton is the content of a template freshly created from clipt without
// tags.
var Skeleton = SkeletonWithTags(nil)

// SkeletonWithTags is the skeleton with tags in its front matter.
func SkeletonWithTags(tags []string) string {
	return "---\n" + frontmatter.Line(frontmatter.Clean(tags)) +
		"\nform:\n  - name: title\n    label: Title\n    type: text\n---\n# {{title}}\n"
}

// Dir returns the templates directory for the notes directory.
func Dir(notesDir string) string { return filepath.Join(notesDir, dirName) }

// Store returns a notes store over the templates directory. The store is
// used as-is; listing a missing directory returns fs.ErrNotExist.
func Store(notesDir string) notes.Store { return notes.Store{Dir: Dir(notesDir)} }

// EnsureDir creates the templates directory when it is missing.
func EnsureDir(notesDir string) error { return os.MkdirAll(Dir(notesDir), 0o755) }

// Field types.
const (
	TypeText   = "text"
	TypeNumber = "number"
	TypeOption = "option"
)

// Field is one input of a template's form.
type Field struct {
	Name    string
	Label   string
	Type    string
	Options []string // TypeOption only
}

// Default returns the field's initial value: empty for text, 0 for number
// and the first option for option.
func (f Field) Default() string {
	switch f.Type {
	case TypeNumber:
		return "0"
	case TypeOption:
		if len(f.Options) > 0 {
			return f.Options[0]
		}
	}
	return ""
}

// Template is a parsed template: its form fields, tags and markdown body.
type Template struct {
	Fields []Field
	Tags   []string
	Body   string
}

type rawField struct {
	Name    string   `yaml:"name"`
	Label   string   `yaml:"label"`
	Type    string   `yaml:"type"`
	Options []string `yaml:"options"`
}

type frontMatter struct {
	Form []rawField `yaml:"form"`
}

// Parse parses a template file. Front matter follows frontmatter.Split; its
// "form" key lists the fields and its "tags" key the template's tags, taken
// literally. The body is the rest with leading blank lines and trailing
// newlines removed.
func Parse(content string) (Template, error) {
	fm, rest, err := frontmatter.Split(content)
	if err != nil {
		return Template{}, err
	}
	if !fm.Present {
		return Template{Body: body(strings.Split(rest, "\n"))}, nil
	}
	var raw frontMatter
	if err := yaml.Unmarshal([]byte(fm.YAML()), &raw); err != nil {
		return Template{}, fmt.Errorf("front matter: %w", err)
	}
	tags, err := frontmatter.Tags(fm.YAML())
	if err != nil {
		return Template{}, fmt.Errorf("front matter: %w", err)
	}
	fields, err := validate(raw.Form)
	if err != nil {
		return Template{}, err
	}
	return Template{Fields: fields, Tags: tags, Body: body(strings.Split(rest, "\n"))}, nil
}

// body joins lines, dropping leading whitespace-only lines and trailing
// newlines.
func body(lines []string) string {
	i := 0
	for i < len(lines) && strings.TrimSpace(lines[i]) == "" {
		i++
	}
	return strings.TrimRight(strings.Join(lines[i:], "\n"), "\n")
}

func validate(raw []rawField) ([]Field, error) {
	seen := make(map[string]bool, len(raw))
	var fields []Field
	for i, r := range raw {
		where := fmt.Sprintf("form[%d]", i)
		if r.Name != "" {
			where = fmt.Sprintf("form[%d] %q", i, r.Name)
		}
		switch {
		case r.Name == "":
			return nil, fmt.Errorf("%s: name is missing", where)
		case strings.ContainsAny(r.Name, "{}") || strings.IndexFunc(r.Name, unicode.IsSpace) >= 0:
			return nil, fmt.Errorf("%s: name must not contain whitespace, { or }", where)
		case seen[r.Name]:
			return nil, fmt.Errorf("%s: duplicate name", where)
		}
		seen[r.Name] = true
		switch r.Type {
		case TypeText, TypeNumber:
		case TypeOption:
			if len(r.Options) == 0 {
				return nil, fmt.Errorf("%s: option field needs a non-empty options list", where)
			}
		case "":
			return nil, fmt.Errorf("%s: type is missing", where)
		default:
			return nil, fmt.Errorf("%s: unsupported type %q", where, r.Type)
		}
		label := r.Label
		if label == "" {
			label = r.Name
		}
		fields = append(fields, Field{Name: r.Name, Label: label, Type: r.Type, Options: r.Options})
	}
	return fields, nil
}

var placeholder = regexp.MustCompile(`\{\{([^{}\s]+)\}\}`)

// Expand replaces every {{name}} of a declared field with its value from
// values in a single pass; values are inserted literally and placeholders
// naming no field are left as they are.
func (t Template) Expand(values map[string]string) string {
	declared := make(map[string]bool, len(t.Fields))
	for _, f := range t.Fields {
		declared[f.Name] = true
	}
	return placeholder.ReplaceAllStringFunc(t.Body, func(m string) string {
		name := m[2 : len(m)-2]
		if !declared[name] {
			return m
		}
		return values[name]
	})
}
