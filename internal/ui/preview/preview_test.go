package preview

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/llysyganicz/clipt/internal/notes"
	"github.com/llysyganicz/clipt/internal/ui/uitest"
)

func setup(t *testing.T, files map[string]string) notes.Store {
	t.Helper()
	dir := t.TempDir()
	for n, c := range files {
		if err := os.WriteFile(filepath.Join(dir, n+".md"), []byte(c), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return notes.Store{Dir: dir}
}

func open(t *testing.T, s notes.Store, name string) Model {
	t.Helper()
	m, err := New(s, notes.Note{Name: name})
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func send(m Model, msg tea.Msg) (Model, tea.Msg) {
	var cmd tea.Cmd
	m, cmd = m.Update(msg)
	if cmd == nil {
		return m, nil
	}
	// Blink and tick commands sleep; only commands that answer at once
	// carry the messages the tests care about.
	ch := make(chan tea.Msg, 1)
	go func() { ch <- cmd() }()
	select {
	case out := <-ch:
		return m, out
	case <-time.After(20 * time.Millisecond):
		return m, nil
	}
}

func runes(m Model, s string) Model {
	for _, r := range s {
		m, _ = send(m, uitest.Key(string(r)))
	}
	return m
}

func press(m Model, k string) (Model, tea.Msg) { return send(m, uitest.Key(k)) }

func resize(m Model, w, h int) Model {
	m, _ = send(m, tea.WindowSizeMsg{Width: w, Height: h})
	return m
}

func longNote() string {
	var b strings.Builder
	for i := 0; i < 100; i++ {
		b.WriteString("line\n\n")
	}
	return b.String()
}

func TestNewMissing(t *testing.T) {
	s := setup(t, nil)
	if _, err := New(s, notes.Note{Name: "gone"}); err == nil {
		t.Fatal("expected error for missing note")
	}
}

func TestNewShowsNote(t *testing.T) {
	s := setup(t, map[string]string{"todo": "# Todo\n\n- a\n- b\n"})
	m := open(t, s, "todo")
	v := uitest.Strip(m.View())
	if !strings.Contains(v, "todo") || !strings.Contains(v, "Todo") {
		t.Errorf("view missing name or heading:\n%s", v)
	}
}

func TestEmptyNote(t *testing.T) {
	s := setup(t, map[string]string{"e": ""})
	m := open(t, s, "e")
	if !strings.Contains(m.View(), "esc back") {
		t.Error("status bar missing for empty note")
	}
}

func TestOnlyArrowsAndPagesScroll(t *testing.T) {
	s := setup(t, map[string]string{"n": longNote()})
	m := resize(open(t, s, "n"), 40, 10)
	for _, r := range []rune{'u', 'f', 'b', 'j', 'k', ' '} {
		m2, _ := send(m, uitest.Key(string(r)))
		if m2.View() != m.View() {
			t.Errorf("%q scrolled", r)
		}
	}
	// d opens the confirm prompt instead of scrolling.
	m2, _ := send(m, uitest.Key("d"))
	if m2.state != stateConfirm {
		t.Error("d did not ask for confirmation")
	}
	before := m.View()
	m, _ = press(m, "down")
	if m.View() == before {
		t.Error("down did not scroll")
	}
	m, _ = press(m, "up")
	if m.View() != before {
		t.Error("up did not scroll back")
	}
	m, _ = press(m, "pgdown")
	if m.vp.YOffset() == 0 {
		t.Error("pgdown did not scroll")
	}
	m, _ = press(m, "pgup")
	if m.vp.YOffset() != 0 {
		t.Error("pgup did not scroll back")
	}
	m, _ = press(m, "up") // stops at the top
	if m.vp.YOffset() != 0 {
		t.Error("scrolled above the top")
	}
}

func TestResizeRerendersAndResetsTop(t *testing.T) {
	s := setup(t, map[string]string{"n": longNote() + "word word word word word word word word word word\n"})
	m := resize(open(t, s, "n"), 80, 10)
	m, _ = press(m, "pgdown")
	m = resize(m, 20, 10)
	if m.vp.YOffset() != 0 {
		t.Error("resize did not reset to top")
	}
	lines := strings.Split(m.View(), "\n")
	for _, l := range lines[:len(lines)-1] { // the status bar clips instead
		if w := lipgloss.Width(l); w > 20 {
			t.Errorf("line wider than 20 (%d): %q", w, l)
		}
	}
}

func TestStatusBar(t *testing.T) {
	s := setup(t, map[string]string{"a-rather-long-note-name": "x"})
	m := resize(open(t, s, "a-rather-long-note-name"), 120, 5)
	last := lastLine(m.View())
	if !strings.Contains(last, "a-rather-long-note-name") || !strings.Contains(last, "esc back") {
		t.Errorf("wide bar = %q", last)
	}
	m = resize(m, 70, 5)
	last = lastLine(m.View())
	if !strings.Contains(last, "esc back") || !strings.Contains(last, "ctrl+q quit") || strings.Contains(last, "a-rather-long-note-name") {
		t.Errorf("narrow bar = %q", last)
	}
	if lipgloss.Width(last) > 70 {
		t.Errorf("narrow bar too wide: %d", lipgloss.Width(last))
	}
}

func lastLine(v string) string {
	l := strings.Split(v, "\n")
	return l[len(l)-1]
}

func TestLeavingKeys(t *testing.T) {
	s := setup(t, map[string]string{"n": "x"})
	m := open(t, s, "n")
	if _, out := press(m, "esc"); out != (BackMsg{}) {
		t.Errorf("esc -> %v", out)
	}
	if _, out := send(m, uitest.Key("e")); out != (EditMsg{Note: notes.Note{Name: "n"}}) {
		t.Errorf("e -> %v", out)
	}
	if _, out := press(m, "ctrl+c"); out != tea.Msg(tea.QuitMsg{}) {
		t.Errorf("ctrl+c -> %v", out)
	}
	m2, out := send(m, uitest.Key("x"))
	if out != nil || m2.View() != m.View() {
		t.Error("x did something")
	}
}

func TestRenameSucceeds(t *testing.T) {
	s := setup(t, map[string]string{"todo": "x"})
	m := open(t, s, "todo")
	m = runes(m, "r")
	if !strings.Contains(uitest.Strip(m.View()), "Rename: todo") {
		t.Fatalf("prompt not prefilled:\n%s", m.View())
	}
	m = runes(m, "-2026")
	m, out := press(m, "enter")
	if r, ok := out.(RenamedMsg); !ok || r.Note.Name != "todo-2026" {
		t.Fatalf("out = %v", out)
	}
	if s.Exists("todo") || !s.Exists("todo-2026") {
		t.Error("file not renamed")
	}
	if m.state != stateView || !strings.Contains(lastLine(m.View()), "todo-2026") {
		t.Errorf("bar = %q", lastLine(m.View()))
	}
}

func TestRenameCancelAndSame(t *testing.T) {
	s := setup(t, map[string]string{"todo": "x"})
	m := runes(open(t, s, "todo"), "r")
	m = runes(m, "zz")
	m, out := press(m, "esc")
	if out != nil || m.state != stateView || !s.Exists("todo") {
		t.Error("esc did not cancel")
	}
	m = runes(m, "r")
	m, out = press(m, "enter")
	if out != nil || m.state != stateView || !s.Exists("todo") {
		t.Error("same name did not cancel")
	}
}

func TestRenameRefused(t *testing.T) {
	s := setup(t, map[string]string{"todo": "T", "ideas": "I"})
	m := runes(open(t, s, "todo"), "r")
	for range "todo" {
		m, _ = press(m, "backspace")
	}
	m = runes(m, "ideas")
	m, out := press(m, "enter")
	if out != nil || m.state != stateRename || !strings.Contains(m.View(), "already exists") {
		t.Errorf("exists not refused: state=%v\n%s", m.state, m.View())
	}
	if c, _ := s.Read("ideas"); c != "I" || !s.Exists("todo") {
		t.Error("files changed")
	}
	// error clears on next key and the prompt stays usable
	for range "ideas" {
		m, _ = press(m, "backspace")
	}
	m = runes(m, "a/b")
	m, out = press(m, "enter")
	if out != nil || m.state != stateRename || !strings.Contains(m.View(), "invalid name") {
		t.Errorf("invalid not refused:\n%s", m.View())
	}
	if !s.Exists("todo") {
		t.Error("todo lost")
	}
}

func TestDeleteFlow(t *testing.T) {
	s := setup(t, map[string]string{"todo": "x", "other": "y"})
	m := runes(open(t, s, "todo"), "d")
	if !strings.Contains(lastLine(m.View()), "(y/n)") {
		t.Error("no confirmation")
	}
	m, out := send(m, uitest.Key("n"))
	if out != nil || m.state != stateView || !s.Exists("todo") {
		t.Error("n did not cancel")
	}
	m = runes(m, "d")
	m, out = press(m, "esc")
	if out != nil || m.state != stateView || !s.Exists("todo") {
		t.Error("esc did not cancel")
	}
	m = runes(m, "d")
	m, out = send(m, uitest.Key("y"))
	if out != (DeletedMsg{}) || s.Exists("todo") || !s.Exists("other") {
		t.Errorf("y: out=%v", out)
	}
}

func TestDeleteFails(t *testing.T) {
	s := setup(t, map[string]string{"todo": "x"})
	m := open(t, s, "todo")
	os.Remove(filepath.Join(s.Dir, "todo.md")) // already gone
	m = runes(m, "d")
	m, out := send(m, uitest.Key("y"))
	if out != nil || m.state != stateView || !strings.Contains(m.View(), "no such file") {
		t.Errorf("delete failure not shown: out=%v\n%s", out, m.View())
	}
}

func TestCtrlQQuitsFromEveryState(t *testing.T) {
	s := setup(t, map[string]string{"n": "x"})
	m := open(t, s, "n")
	if _, out := press(m, "ctrl+q"); out != tea.Msg(tea.QuitMsg{}) {
		t.Errorf("ctrl+q -> %v", out)
	}
	rn := runes(m, "r")
	rn = runes(rn, "2")
	if rn.state != stateRename {
		t.Fatalf("state = %v", rn.state)
	}
	if _, out := press(rn, "ctrl+q"); out != tea.Msg(tea.QuitMsg{}) {
		t.Errorf("ctrl+q during rename -> %v", out)
	}
	if _, err := os.Stat(filepath.Join(s.Dir, "n.md")); err != nil {
		t.Error("note renamed by ctrl+q")
	}
	dl := runes(m, "d")
	if dl.state != stateConfirm {
		t.Fatalf("state = %v", dl.state)
	}
	if _, out := press(dl, "ctrl+q"); out != tea.Msg(tea.QuitMsg{}) {
		t.Errorf("ctrl+q during delete confirmation -> %v", out)
	}
	if _, err := os.Stat(filepath.Join(s.Dir, "n.md")); err != nil {
		t.Error("note deleted by ctrl+q")
	}
}
