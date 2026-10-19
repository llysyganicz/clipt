package finder

import (
	"reflect"
	"testing"

	"github.com/llysyganicz/clipt/internal/frontmatter"
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
			out = append(out, "+"+r.Create+tagSuffix(r.CreateTags))
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
			got := names(BuildRows(all, tc.current, ParseQuery(tc.query)))
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("rows = %v, want %v", got, tc.want)
			}
		})
	}
}

func tagSuffix(tags []string) string {
	if len(tags) == 0 {
		return ""
	}
	return " " + frontmatter.Display(tags)
}

func TestParseQuery(t *testing.T) {
	tests := []struct {
		in   string
		want Query
	}{
		{"", Query{}},
		{"todo", Query{Name: "todo"}},
		{" a  b ", Query{Name: " a  b "}},
		{"a#b", Query{Name: "a#b"}},
		{"#work", Query{Tags: []string{"work"}}},
		{"#work #meet sta", Query{Name: "sta", Tags: []string{"work", "meet"}}},
		{"sta  #work  up", Query{Name: "sta up", Tags: []string{"work"}}},
		{"# sta", Query{Name: "sta"}},
		{"##x", Query{Tags: []string{"#x"}}},
	}
	for _, tc := range tests {
		if got := ParseQuery(tc.in); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("ParseQuery(%q) = %#v, want %#v", tc.in, got, tc.want)
		}
	}
}

func TestTagQuery(t *testing.T) {
	all := []notes.Note{
		{Name: "standup", Tags: []string{"Work", "meeting"}},
		{Name: "review", Tags: []string{"work"}},
		{Name: "ideas", Tags: []string{"home"}},
		{Name: "draft"},
	}
	tests := []struct {
		name, query string
		want        []string
	}{
		{"tag filter", "#work", []string{"standup", "review"}},
		{"prefix match", "#wo", []string{"standup", "review"}},
		{"case-insensitive", "#WORK", []string{"standup", "review"}},
		{"several filters", "#work #meet", []string{"standup"}},
		{"tags and name", "#work sta", []string{"standup", "+sta #work"}},
		{"filter position", "sta #work", []string{"standup", "+sta #work"}},
		{"bare hash ignored", "# sta", []string{"standup", "+sta"}},
		{"untagged and a filter", "#home", []string{"ideas"}},
		{"no match", "#nothing", []string{}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := names(BuildRows(all, nil, ParseQuery(tc.query)))
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("rows = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestCreateRowTags(t *testing.T) {
	all := []notes.Note{{Name: "standup"}, {Name: "review", Tags: []string{"work"}}}
	tests := []struct {
		name, query string
		want        []string
	}{
		{"create row with tags", "#work stand-up", []string{"+stand-up #work"}},
		{"no create row for tags alone", "#work", []string{"review"}},
		{"exact name with a tag filter", "#work standup", []string{}},
		{"deduplicated tags", "#work #Work #meeting x", []string{"+x #work #meeting"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := names(BuildRows(all, nil, ParseQuery(tc.query)))
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("rows = %v, want %v", got, tc.want)
			}
		})
	}
}
