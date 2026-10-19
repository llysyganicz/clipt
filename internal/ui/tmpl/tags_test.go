package tmpl

import (
	"reflect"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/llysyganicz/clipt/internal/ui/uitest"
)

func rowNames(m Model) []string {
	out := []string{}
	for _, n := range m.Rows() {
		out = append(out, n.Name)
	}
	return out
}

func TestPickerFilterByTag(t *testing.T) {
	m := New(setup(t, []string{"daily", "meeting"}, map[string]string{
		"meeting": "---\ntags: [work]\n---\n- agenda\n",
	}))
	m, _ = press(m, "#work")
	if got := rowNames(m); !reflect.DeepEqual(got, []string{"meeting"}) {
		t.Errorf("rows = %v", got)
	}
	m, _ = press(m, " d")
	if got := rowNames(m); len(got) != 0 {
		t.Errorf("rows for #work d = %v", got)
	}
}

func TestPickerRowsShowTags(t *testing.T) {
	m := New(setup(t, []string{"daily", "meeting"}, map[string]string{
		"meeting": "---\ntags: [work]\n---\n- agenda\n",
	}))
	m, _ = m.Update(tea.WindowSizeMsg{Width: 60, Height: 10})
	l := strings.Split(uitest.Strip(m.View()), "\n")
	if l[1] != "> meeting #work" || l[2] != "  daily" {
		t.Errorf("rows = %q / %q", l[1], l[2])
	}
	m, _ = m.Update(tea.WindowSizeMsg{Width: 13, Height: 10})
	if l := strings.Split(uitest.Strip(m.View()), "\n"); l[1] != "> meeting #w…" {
		t.Errorf("narrow row = %q", l[1])
	}
}

func TestInsertCarriesTags(t *testing.T) {
	store := setup(t, []string{"plain", "form", "daily"}, map[string]string{
		"daily": "---\ntags: [meeting, Meeting, work]\n---\n- agenda\n",
		"form":  "---\ntags: [\"{{topic}}\"]\nform:\n  - name: topic\n    type: text\n---\n# {{topic}}\n",
	})
	_, msg := press(New(store), "enter") // daily: no fields
	if !reflect.DeepEqual(msg, InsertMsg{Text: "- agenda", Tags: []string{"meeting", "work"}}) {
		t.Errorf("no-fields insert = %#v", msg)
	}
	m, _ := press(New(store), "down", "enter") // form
	if !m.InForm() {
		t.Fatal("form not open")
	}
	_, msg = press(m, "Hi", "enter")
	if !reflect.DeepEqual(msg, InsertMsg{Text: "# Hi", Tags: []string{"{{topic}}"}}) {
		t.Errorf("submit insert = %#v", msg)
	}
	_, msg = press(New(store), "down", "down", "enter") // plain
	if got, ok := msg.(InsertMsg); !ok || got.Tags != nil {
		t.Errorf("untagged insert = %#v", msg)
	}
}
