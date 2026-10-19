package frontmatter

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestSplit(t *testing.T) {
	cases := []struct {
		name    string
		in      string
		want    Block
		body    string
		wantErr error
	}{
		{
			name: "present",
			in:   "---\ntags: [work]\n---\n# Standup\n",
			want: Block{Present: true, Lines: []string{"tags: [work]"}, Start: 0, End: 2},
			body: "# Standup\n",
		},
		{
			name: "absent",
			in:   "# Standup\n---\n",
			want: Block{End: -1},
			body: "# Standup\n---\n",
		},
		{
			name:    "unclosed",
			in:      "---\ntags: [work]\n# Standup\n",
			want:    Block{Present: true, Lines: []string{"tags: [work]", "# Standup", ""}, End: -1},
			body:    "---\ntags: [work]\n# Standup\n",
			wantErr: ErrUnclosed,
		},
		{
			name: "crlf",
			in:   "---\r\ntitle: x\r\n---\r\nbody\r\n",
			want: Block{Present: true, Lines: []string{"title: x"}, End: 2},
			body: "body\n",
		},
		{
			name: "empty block",
			in:   "---\n---\n",
			want: Block{Present: true, Lines: []string{}, End: 1},
			body: "",
		},
		{
			name: "empty file",
			in:   "",
			want: Block{End: -1},
			body: "",
		},
		{
			name: "fence with trailing space is not a fence",
			in:   "--- \nx\n",
			want: Block{End: -1},
			body: "--- \nx\n",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			fm, body, err := Split(c.in)
			if !errors.Is(err, c.wantErr) {
				t.Fatalf("err = %v, want %v", err, c.wantErr)
			}
			if !reflect.DeepEqual(fm, c.want) {
				t.Errorf("block = %+v, want %+v", fm, c.want)
			}
			if body != c.body {
				t.Errorf("body = %q, want %q", body, c.body)
			}
		})
	}
}

func TestTags(t *testing.T) {
	cases := []struct {
		name    string
		in      string
		want    []string
		wantErr string
	}{
		{"flow list", "tags: [work, meeting]", []string{"work", "meeting"}, ""},
		{"block list", "tags:\n  - work\n  - meeting", []string{"work", "meeting"}, ""},
		{"block list at column 0", "tags:\n- work\n- meeting", []string{"work", "meeting"}, ""},
		{"non-string scalar", "tags: [2026, work]", []string{"2026", "work"}, ""},
		{"duplicates and blanks", `tags: [Work, "", work, home]`, []string{"Work", "home"}, ""},
		{"trimmed", `tags: [" a ", b]`, []string{"a", "b"}, ""},
		{"null values dropped", "tags: [a, ~]", []string{"a"}, ""},
		{"other keys", "title: Standup\ntags: [work]", []string{"work"}, ""},
		{"no key", "title: Standup", nil, ""},
		{"empty list", "tags: []", nil, ""},
		{"null value", "tags:", nil, ""},
		{"empty yaml", "", nil, ""},
		{"plain string", "tags: work", nil, "tags must be a list"},
		{"nested map", "tags: [a, {b: c}]", nil, "tags[1]"},
		{"nested list", "tags: [a, [b]]", nil, "tags[1]"},
		{"invalid yaml", "tags: [work", nil, "]"},
		{"not a mapping", "- a", nil, "mapping"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := Tags(c.in)
			if c.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), c.wantErr) {
					t.Fatalf("err = %v, want containing %q", err, c.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Tags: %v", err)
			}
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("Tags = %#v, want %#v", got, c.want)
			}
		})
	}
}

func TestParse(t *testing.T) {
	if got, err := Parse("---\ntags: [work]\n---\n# Standup\n"); err != nil || !reflect.DeepEqual(got, []string{"work"}) {
		t.Errorf("with front matter = %v, %v", got, err)
	}
	if got, err := Parse("# Standup\n---\n"); err != nil || got != nil {
		t.Errorf("without front matter = %v, %v", got, err)
	}
	if _, err := Parse("---\ntags: [work]\n"); !errors.Is(err, ErrUnclosed) {
		t.Errorf("unclosed err = %v", err)
	}
	if _, err := Parse("---\ntags: work\n---\n"); err == nil {
		t.Error("plain string tags accepted")
	}
}

func TestFlowRoundTrip(t *testing.T) {
	cases := [][]string{
		{"work", "meeting"},
		{"a, b", "x: y", "[z]", "#t"},
		{"2026", "yes", "null", "true", "é", `a"b`, "{x}", "- d", "a'b"},
	}
	for _, tags := range cases {
		line := Line(tags)
		if strings.Contains(line, "\n") {
			t.Errorf("Line(%q) spans lines: %q", tags, line)
		}
		got, err := Tags(line)
		if err != nil || !reflect.DeepEqual(got, tags) {
			t.Errorf("round trip of %q via %q = %q, %v", tags, line, got, err)
		}
	}
	if got := Line([]string{"work", "meeting"}); got != "tags: [work, meeting]" {
		t.Errorf("plain tags line = %q", got)
	}
	if got := Line(nil); got != "tags: []" {
		t.Errorf("empty line = %q", got)
	}
}

func TestDisplay(t *testing.T) {
	if got := Display([]string{"work", "meeting"}); got != "#work #meeting" {
		t.Errorf("Display = %q", got)
	}
	if got := Display(nil); got != "" {
		t.Errorf("Display(nil) = %q", got)
	}
}

func TestMergeTags(t *testing.T) {
	cases := []struct {
		name    string
		in      string
		add     []string
		want    string
		delta   int
		wantErr bool
	}{
		{
			name:  "no front matter",
			in:    "# Standup",
			add:   []string{"meeting"},
			want:  "---\ntags: [meeting]\n---\n# Standup",
			delta: 3,
		},
		{
			name:  "empty note",
			in:    "",
			add:   []string{"meeting"},
			want:  "---\ntags: [meeting]\n---\n",
			delta: 3,
		},
		{
			name:  "front matter without tags",
			in:    "---\ntitle: Standup\n---\nbody",
			add:   []string{"meeting"},
			want:  "---\ntitle: Standup\ntags: [meeting]\n---\nbody",
			delta: 1,
		},
		{
			name:  "empty front matter",
			in:    "---\n---\nbody",
			add:   []string{"meeting"},
			want:  "---\ntags: [meeting]\n---\nbody",
			delta: 1,
		},
		{
			name:  "existing block list",
			in:    "---\ntags:\n  - work\ntitle: x\n---\nbody",
			add:   []string{"meeting"},
			want:  "---\ntags: [work, meeting]\ntitle: x\n---\nbody",
			delta: -1,
		},
		{
			name:  "block list collapses",
			in:    "---\ntags:\n  - work\n  - home\n---\nbody",
			add:   []string{"meeting"},
			want:  "---\ntags: [work, home, meeting]\n---\nbody",
			delta: -2,
		},
		{
			name:  "column 0 block list",
			in:    "---\ntags:\n- work\ntitle: x\n---\nbody",
			add:   []string{"meeting"},
			want:  "---\ntags: [work, meeting]\ntitle: x\n---\nbody",
			delta: -1,
		},
		{
			name:  "existing flow list",
			in:    "---\n# comment\ntags: [work]\nz: 1\n---\nbody",
			add:   []string{"meeting", "Work"},
			want:  "---\n# comment\ntags: [work, meeting]\nz: 1\n---\nbody",
			delta: 0,
		},
		{
			name:  "empty tags entry",
			in:    "---\ntags: []\n---\nbody",
			add:   []string{"a"},
			want:  "---\ntags: [a]\n---\nbody",
			delta: 0,
		},
		{
			name: "already present",
			in:   "---\ntags: [work]\n---\nbody",
			add:  []string{"Work"},
			want: "---\ntags: [work]\n---\nbody",
		},
		{
			name: "nothing to add",
			in:   "# Standup",
			add:  []string{"", "  "},
			want: "# Standup",
		},
		{
			name:  "duplicates in add",
			in:    "x",
			add:   []string{"a", "A", "b"},
			want:  "---\ntags: [a, b]\n---\nx",
			delta: 3,
		},
		{
			name:    "invalid front matter",
			in:      "---\ntags: [work\n---\nbody",
			add:     []string{"meeting"},
			want:    "---\ntags: [work\n---\nbody",
			wantErr: true,
		},
		{
			name:    "unclosed front matter",
			in:      "---\ntags: [work]\nbody",
			add:     []string{"meeting"},
			want:    "---\ntags: [work]\nbody",
			wantErr: true,
		},
		{
			name:    "tags not at column 0 is refused",
			in:      "---\n? tags\n: [work]\n---\nbody",
			add:     []string{"meeting"},
			want:    "---\n? tags\n: [work]\n---\nbody",
			wantErr: true,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, delta, err := MergeTags(c.in, c.add)
			if (err != nil) != c.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, c.wantErr)
			}
			if got != c.want {
				t.Errorf("merged = %q, want %q", got, c.want)
			}
			if delta != c.delta {
				t.Errorf("delta = %d, want %d", delta, c.delta)
			}
		})
	}
}

func TestMergeTagNeedingQuotes(t *testing.T) {
	got, _, err := MergeTags("body", []string{"a, b"})
	if err != nil {
		t.Fatal(err)
	}
	tags, err := Parse(got)
	if err != nil || !reflect.DeepEqual(tags, []string{"a, b"}) {
		t.Errorf("merged %q parses to %q, %v", got, tags, err)
	}
}
