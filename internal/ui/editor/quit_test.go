package editor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestCtrlQSavesAndQuits(t *testing.T) {
	s := setup(t, map[string]string{"todo": "a\n"})
	m := runes(open(t, s, "todo"), "b")
	m, out := press(m, "ctrl+q")
	if out != tea.Msg(tea.QuitMsg{}) {
		t.Errorf("ctrl+q -> %v", out)
	}
	if b, old := file(t, s, "todo"); b != "ab\n" || old {
		t.Errorf("file = %q old=%v", b, old)
	}
	if m.err != "" {
		t.Errorf("err = %q", m.err)
	}

	ts := setupTemplate(t, map[string]string{"meeting": "a\n"})
	tm := runes(openTemplate(t, ts, "meeting"), "b")
	if _, out := press(tm, "ctrl+q"); out != tea.Msg(tea.QuitMsg{}) {
		t.Errorf("template ctrl+q -> %v", out)
	}
	if b, err := os.ReadFile(filepath.Join(ts.Dir, "meeting.md")); err != nil || string(b) != "ab\n" {
		t.Errorf("template = %q, %v", b, err)
	}
}

func TestCtrlQUnchangedDoesNotWrite(t *testing.T) {
	s := setup(t, map[string]string{"todo": "a\n"})
	m := open(t, s, "todo")
	m, _ = press(m, "home")
	if _, out := press(m, "ctrl+q"); out != tea.Msg(tea.QuitMsg{}) {
		t.Errorf("ctrl+q -> %v", out)
	}
	if b, old := file(t, s, "todo"); b != "a\n" || !old {
		t.Errorf("file touched: %q old=%v", b, old)
	}
}

func TestCtrlQWriteFailureStaysOpen(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root ignores permissions")
	}
	s := setup(t, map[string]string{"todo": "a\n"})
	m := runes(open(t, s, "todo"), "b")
	p := filepath.Join(s.Dir, "todo.md")
	if err := os.Chmod(p, 0o444); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(s.Dir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(s.Dir, 0o755); os.Chmod(p, 0o644) })

	m, out := press(m, "ctrl+q")
	if out != nil {
		t.Fatalf("ctrl+q on failure -> %v", out)
	}
	if m.err == "" || !strings.Contains(view(m), "permission denied") {
		t.Errorf("error not shown: err=%q", m.err)
	}
	if got := m.ta.Value(); got != "ab" {
		t.Errorf("content lost: %q", got)
	}
	if b, old := file(t, s, "todo"); b != "a\n" || !old {
		t.Errorf("file touched: %q old=%v", b, old)
	}
}
