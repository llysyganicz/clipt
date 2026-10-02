package finder

import (
	"reflect"
	"testing"

	"github.com/llysyganicz/clipt/internal/notes"
)

func TestMatch(t *testing.T) {
	tests := []struct {
		q, name string
		want    bool
	}{
		{"tdo", "todo-archive", true},
		{"tdo", "todos", true},
		{"tdo", "ideas", false},
		{"mn", "Meeting Notes", true},
		{"MN", "meeting notes", true},
		{"", "anything", true},
		{"", "", true},
		{".md", "foo", false},
		{"foo.md", "foo", false},
		{"éü", "Écüme", true},
		{"日本", "日本語", true},
		{"本日", "日本語", false},
		{"abc", "ab", false},
	}
	for _, tc := range tests {
		if got := Match(tc.q, tc.name); got != tc.want {
			t.Errorf("Match(%q, %q) = %v, want %v", tc.q, tc.name, got, tc.want)
		}
	}
}

func names(rows []Row) []string {
	out := []string{}
	for _, r := range rows {
		if r.IsCreate() {
			out = append(out, "+"+r.Create)
		} else {
			out = append(out, r.Note.Name)
		}
	}
	return out
}

func TestBuildRows(t *testing.T) {
	all := []notes.Note{{Name: "todos"}, {Name: "todo-archive"}, {Name: "ideas"}, {Name: "b"}}
	cur := &notes.Note{Name: "b"}
	tests := []struct {
		name    string
		current *notes.Note
		query   string
		want    []string
	}{
		{"empty query order", nil, "", []string{"todos", "todo-archive", "ideas", "b"}},
		{"current hidden", cur, "", []string{"todos", "todo-archive", "ideas"}},
		{"filtered keeps order", nil, "tdo", []string{"todos", "todo-archive", "+tdo"}},
		{"create row last", nil, "todo", []string{"todos", "todo-archive", "+todo"}},
		{"create row alone", nil, "meeting-notes", []string{"+meeting-notes"}},
		{"no create on exact name", nil, "ideas", []string{"ideas"}},
		{"no create on hidden current exact", cur, "b", []string{}},
		{"no create on blank query", nil, "   ", []string{}},
		{"trimmed create name", nil, " zzz ", []string{"+zzz"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := names(BuildRows(all, tc.current, tc.query))
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("rows = %v, want %v", got, tc.want)
			}
		})
	}
}
