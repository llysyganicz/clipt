package templates

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestDirStoreEnsureDir(t *testing.T) {
	dir := t.TempDir()
	want := filepath.Join(dir, ".templates")
	if got := Dir(dir); got != want {
		t.Errorf("Dir = %q, want %q", got, want)
	}
	s := Store(dir)
	if s.Dir != want {
		t.Errorf("Store.Dir = %q, want %q", s.Dir, want)
	}
	if _, err := s.List(); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("List on missing dir err = %v, want ErrNotExist", err)
	}
	if err := EnsureDir(dir); err != nil {
		t.Fatalf("EnsureDir: %v", err)
	}
	info, err := os.Stat(want)
	if err != nil || !info.IsDir() {
		t.Fatalf("stat %q: %v, isDir=%v", want, err, info != nil && info.IsDir())
	}
	if err := EnsureDir(dir); err != nil {
		t.Errorf("second EnsureDir: %v", err)
	}
	if got, err := s.List(); err != nil || len(got) != 0 {
		t.Errorf("List on empty dir = %v, %v", got, err)
	}
}

func TestParse(t *testing.T) {
	text := func(name string) Field { return Field{Name: name, Label: name, Type: TypeText} }
	cases := []struct {
		name    string
		in      string
		want    Template
		wantErr string // substring; empty means success
	}{
		{
			name: "no front matter",
			in:   "# Daily\n\n- [ ] \n",
			want: Template{Body: "# Daily\n\n- [ ] "},
		},
		{
			name: "front matter with fields",
			in:   "---\nform:\n  - name: topic\n    type: text\n---\n# {{topic}}\n",
			want: Template{Fields: []Field{text("topic")}, Body: "# {{topic}}"},
		},
		{
			name: "crlf",
			in:   "---\r\nform:\r\n  - name: topic\r\n    type: text\r\n---\r\n# {{topic}}\r\n",
			want: Template{Fields: []Field{text("topic")}, Body: "# {{topic}}"},
		},
		{
			name:    "unterminated fence",
			in:      "---\nform: []\n# Title\n",
			wantErr: "not closed",
		},
		{
			name:    "invalid yaml",
			in:      "---\nform: [\n---\n# Title\n",
			wantErr: "front matter",
		},
		{
			name:    "missing name",
			in:      "---\nform:\n  - type: text\n---\n",
			wantErr: "form[0]",
		},
		{
			name:    "name with brace",
			in:      "---\nform:\n  - name: to}pic\n    type: text\n---\n",
			wantErr: "to}pic",
		},
		{
			name:    "name with whitespace",
			in:      "---\nform:\n  - name: to pic\n    type: text\n---\n",
			wantErr: "to pic",
		},
		{
			name:    "duplicate name",
			in:      "---\nform:\n  - name: topic\n    type: text\n  - name: topic\n    type: number\n---\n",
			wantErr: "topic",
		},
		{
			name:    "unknown type",
			in:      "---\nform:\n  - name: when\n    type: date\n---\n",
			wantErr: "date",
		},
		{
			name:    "option without options",
			in:      "---\nform:\n  - name: prio\n    type: option\n---\n",
			wantErr: "prio",
		},
		{
			name: "numeric and boolean options",
			in:   "---\nform:\n  - name: n\n    type: option\n    options: [1, 2.5, yes, \"no\"]\n---\nx\n",
			want: Template{Fields: []Field{{Name: "n", Label: "n", Type: TypeOption, Options: []string{"1", "2.5", "yes", "no"}}}, Body: "x"},
		},
		{
			name: "extra keys ignored",
			in:   "---\ntitle: Meeting\nform:\n  - name: topic\n    type: text\n---\n# {{topic}}\n",
			want: Template{Fields: []Field{text("topic")}, Body: "# {{topic}}"},
		},
		{
			name: "template with tags",
			in:   "---\ntags: [meeting]\nform:\n  - name: topic\n    type: text\n---\n# {{topic}}\n",
			want: Template{Fields: []Field{text("topic")}, Tags: []string{"meeting"}, Body: "# {{topic}}"},
		},
		{
			name: "tags follow the tags key rules",
			in:   "---\ntags:\n  - Work\n  - work\n  - 2026\n---\nx\n",
			want: Template{Tags: []string{"Work", "2026"}, Body: "x"},
		},
		{
			name: "placeholder in tags is literal",
			in:   "---\ntags: [\"{{topic}}\"]\nform:\n  - name: topic\n    type: text\n---\n{{topic}}\n",
			want: Template{Fields: []Field{text("topic")}, Tags: []string{"{{topic}}"}, Body: "{{topic}}"},
		},
		{
			name:    "invalid tags",
			in:      "---\ntags: meeting\n---\nx\n",
			wantErr: "tags",
		},
		{
			name: "label given",
			in:   "---\nform:\n  - name: topic\n    label: Topic\n    type: text\n---\n",
			want: Template{Fields: []Field{{Name: "topic", Label: "Topic", Type: TypeText}}},
		},
		{
			name: "no form key",
			in:   "---\ntitle: x\n---\nbody\n",
			want: Template{Body: "body"},
		},
		{
			name: "empty form",
			in:   "---\nform: []\n---\n\n# Title\n",
			want: Template{Body: "# Title"},
		},
		{
			name: "several blank lines after fence",
			in:   "---\nform: []\n---\n\n\n\n# Title\n",
			want: Template{Body: "# Title"},
		},
		{
			name: "whitespace-only lines after fence",
			in:   "---\nform: []\n---\n  \n\t\n# Title\n",
			want: Template{Body: "# Title"},
		},
		{
			name: "blank lines inside body kept",
			in:   "---\nform: []\n---\n# Title\n\ntext\n",
			want: Template{Body: "# Title\n\ntext"},
		},
		{
			name: "trailing newlines trimmed",
			in:   "# Title\n\n\n",
			want: Template{Body: "# Title"},
		},
		{
			name: "fence alone",
			in:   "---\n---\n",
			want: Template{},
		},
		{
			name: "dashes not on first line are body",
			in:   "x\n---\ny\n",
			want: Template{Body: "x\n---\ny"},
		},
		{
			name: "empty file",
			in:   "",
			want: Template{},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := Parse(c.in)
			if c.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), c.wantErr) {
					t.Fatalf("err = %v, want containing %q", err, c.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("Parse = %+v, want %+v", got, c.want)
			}
		})
	}
}

func TestFieldDefault(t *testing.T) {
	cases := []struct {
		f    Field
		want string
	}{
		{Field{Type: TypeText}, ""},
		{Field{Type: TypeNumber}, "0"},
		{Field{Type: TypeOption, Options: []string{"low", "high"}}, "low"},
	}
	for _, c := range cases {
		if got := c.f.Default(); got != c.want {
			t.Errorf("%s default = %q, want %q", c.f.Type, got, c.want)
		}
	}
}

func TestExpand(t *testing.T) {
	fields := func(names ...string) []Field {
		var out []Field
		for _, n := range names {
			out = append(out, Field{Name: n, Label: n, Type: TypeText})
		}
		return out
	}
	cases := []struct {
		name   string
		tpl    Template
		values map[string]string
		want   string
	}{
		{
			name:   "replace",
			tpl:    Template{Fields: fields("topic", "prio"), Body: "# {{topic}} ({{prio}})"},
			values: map[string]string{"topic": "Standup", "prio": "high"},
			want:   "# Standup (high)",
		},
		{
			name:   "repeated",
			tpl:    Template{Fields: fields("a"), Body: "{{a}} and {{a}}"},
			values: map[string]string{"a": "x"},
			want:   "x and x",
		},
		{
			name:   "unknown left intact",
			tpl:    Template{Fields: fields("topic"), Body: "{{topic}} {{date}}"},
			values: map[string]string{"topic": "x"},
			want:   "x {{date}}",
		},
		{
			name:   "value with placeholder not re-expanded",
			tpl:    Template{Fields: fields("a", "b"), Body: "{{a}} {{b}}"},
			values: map[string]string{"a": "{{b}}", "b": "y"},
			want:   "{{b}} y",
		},
		{
			name:   "spaces are not a placeholder",
			tpl:    Template{Fields: fields("name"), Body: "{{ name }} {{name}}"},
			values: map[string]string{"name": "n"},
			want:   "{{ name }} n",
		},
		{
			name:   "missing value is empty",
			tpl:    Template{Fields: fields("a"), Body: "[{{a}}]"},
			values: nil,
			want:   "[]",
		},
	}
	for _, c := range cases {
		if got := c.tpl.Expand(c.values); got != c.want {
			t.Errorf("%s: Expand = %q, want %q", c.name, got, c.want)
		}
	}
}

func TestSkeletonParses(t *testing.T) {
	if !strings.Contains(Skeleton, "\ntags: []\n") {
		t.Errorf("skeleton lacks tags: []: %q", Skeleton)
	}
	tpl, err := Parse(Skeleton)
	if err != nil {
		t.Fatalf("Parse(Skeleton): %v", err)
	}
	if len(tpl.Tags) != 0 {
		t.Errorf("tags = %q", tpl.Tags)
	}
	if len(tpl.Fields) != 1 || tpl.Fields[0].Name != "title" || tpl.Fields[0].Type != TypeText {
		t.Errorf("fields = %+v", tpl.Fields)
	}
	if tpl.Body != "# {{title}}" {
		t.Errorf("body = %q", tpl.Body)
	}
	if got := tpl.Expand(map[string]string{"title": "Hi"}); got != "# Hi" {
		t.Errorf("expanded = %q", got)
	}
}

func TestSkeletonWithTags(t *testing.T) {
	tpl, err := Parse(SkeletonWithTags([]string{"meeting", "Meeting", "a, b"}))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if !reflect.DeepEqual(tpl.Tags, []string{"meeting", "a, b"}) {
		t.Errorf("tags = %q", tpl.Tags)
	}
	if len(tpl.Fields) != 1 || tpl.Fields[0].Type != TypeText {
		t.Errorf("fields = %+v", tpl.Fields)
	}
	if SkeletonWithTags(nil) != Skeleton {
		t.Error("SkeletonWithTags(nil) differs from Skeleton")
	}
}

func TestStoreListsTemplateTags(t *testing.T) {
	dir := t.TempDir()
	if err := EnsureDir(dir); err != nil {
		t.Fatal(err)
	}
	s := Store(dir)
	if err := s.Write("meeting", SkeletonWithTags([]string{"meeting", "work"})); err != nil {
		t.Fatal(err)
	}
	if err := s.Write("daily", "# Daily\n"); err != nil {
		t.Fatal(err)
	}
	if err := s.Write("broken", "---\ntags: [x\n"); err != nil {
		t.Fatal(err)
	}
	got, err := s.List()
	if err != nil {
		t.Fatal(err)
	}
	tags := map[string][]string{}
	for _, n := range got {
		tags[n.Name] = n.Tags
	}
	want := map[string][]string{"meeting": {"meeting", "work"}, "daily": nil, "broken": nil}
	if !reflect.DeepEqual(tags, want) {
		t.Errorf("tags = %q, want %q", tags, want)
	}
}
