package editor

import (
	"strings"
	"testing"
)

// typedAfter types s at the cursor and returns the text area's value.
func typedAfter(m Model, s string) string { return runes(m, s).ta.Value() }

func TestInsertTemplateIntoNoteWithoutFrontMatter(t *testing.T) {
	m := open(t, setup(t, map[string]string{"n": "# Standup\n"}), "n")
	m = m.InsertTemplate("- agenda", []string{"meeting"})
	if got := m.ta.Value(); got != "---\ntags: [meeting]\n---\n# Standup- agenda" {
		t.Errorf("value = %q", got)
	}
	if m.ta.Line() != 3 {
		t.Errorf("cursor row = %d, want 3", m.ta.Line())
	}
	if got := typedAfter(m, "X"); !strings.HasSuffix(got, "- agendaX") {
		t.Errorf("cursor not after the body: %q", got)
	}
}

func TestInsertTemplateMultiLineBody(t *testing.T) {
	m := open(t, setup(t, map[string]string{"n": "ab\n"}), "n")
	m, _ = press(m, "left")
	m = m.InsertTemplate("1\n2", []string{"t"})
	if got := typedAfter(m, "X"); got != "---\ntags: [t]\n---\na1\n2Xb" {
		t.Errorf("value = %q", got)
	}
}

func TestInsertTemplateAppendsTags(t *testing.T) {
	m := open(t, setup(t, map[string]string{"n": "---\ntags: [work]\n---\nbody\n"}), "n")
	m = m.InsertTemplate("!", []string{"meeting", "Work"})
	if got := typedAfter(m, "X"); got != "---\ntags: [work, meeting]\n---\nbody!X" {
		t.Errorf("value = %q", got)
	}
}

func TestInsertTemplateBlockListCollapse(t *testing.T) {
	m := open(t, setup(t, map[string]string{"n": "---\ntags:\n  - a\n  - b\n---\nline1\nline2\n"}), "n")
	m, _ = press(m, "up")
	m = m.InsertTemplate("Z", []string{"c"})
	if m.ta.Line() != 3 {
		t.Errorf("cursor row = %d, want 3", m.ta.Line())
	}
	if got := typedAfter(m, "Y"); got != "---\ntags: [a, b, c]\n---\nline1ZY\nline2" {
		t.Errorf("value = %q", got)
	}
}

func TestInsertTemplateWithoutTags(t *testing.T) {
	m := open(t, setup(t, map[string]string{"n": "x\n"}), "n")
	m = m.InsertTemplate("-y", nil)
	if got := m.ta.Value(); got != "x-y" {
		t.Errorf("value = %q", got)
	}
}

func TestInsertTemplateBrokenFrontMatter(t *testing.T) {
	m := open(t, setup(t, map[string]string{"n": "---\ntags: [work\n---\nbody\n"}), "n")
	m = m.InsertTemplate("!", []string{"meeting"})
	if got := m.ta.Value(); got != "---\ntags: [work\n---\nbody!" {
		t.Errorf("value = %q", got)
	}
	if v := view(m); !strings.Contains(v, "tags not merged") {
		t.Errorf("no error shown:\n%s", v)
	}
}
