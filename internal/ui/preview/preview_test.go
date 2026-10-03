package preview

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/llysyganicz/clipt/internal/notes"
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
	nm, cmd := m.Update(msg)
	m = nm.(Model)
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
		m, _ = send(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	return m
}

func press(m Model, k tea.KeyType) (Model, tea.Msg) { return send(m, tea.KeyMsg{Type: k}) }

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
	v := m.View()
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
		m2, _ := send(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		if m2.View() != m.View() {
			t.Errorf("%q scrolled", r)
		}
	}
	// d opens the confirm prompt instead of scrolling.
	m2, _ := send(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	if m2.state != stateConfirm {
		t.Error("d did not ask for confirmation")
	}
	before := m.View()
	m, _ = press(m, tea.KeyDown)
	if m.View() == before {
		t.Error("down did not scroll")
	}
	m, _ = press(m, tea.KeyUp)
	if m.View() != before {
		t.Error("up did not scroll back")
	}
	m, _ = press(m, tea.KeyPgDown)
	if m.vp.YOffset == 0 {
		t.Error("pgdown did not scroll")
	}
	m, _ = press(m, tea.KeyPgUp)
	if m.vp.YOffset != 0 {
		t.Error("pgup did not scroll back")
	}
	m, _ = press(m, tea.KeyUp) // stops at the top
	if m.vp.YOffset != 0 {
		t.Error("scrolled above the top")
	}
}

func TestResizeRerendersAndResetsTop(t *testing.T) {
	s := setup(t, map[string]string{"n": longNote() + "word word word word word word word word word word\n"})
	m := resize(open(t, s, "n"), 80, 10)
	m, _ = press(m, tea.KeyPgDown)
	m = resize(m, 20, 10)
	if m.vp.YOffset != 0 {
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
	m = resize(m, 50, 5)
	last = lastLine(m.View())
	if !strings.Contains(last, "esc back") || strings.Contains(last, "a-rather-long-note-name") {
		t.Errorf("narrow bar = %q", last)
	}
	if lipgloss.Width(last) > 50 {
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
	if _, out := press(m, tea.KeyEsc); out != (BackMsg{}) {
		t.Errorf("esc -> %v", out)
	}
	if _, out := send(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}}); out != (EditMsg{Note: notes.Note{Name: "n"}}) {
		t.Errorf("e -> %v", out)
	}
	if _, out := press(m, tea.KeyCtrlC); out != tea.Msg(tea.QuitMsg{}) {
		t.Errorf("ctrl+c -> %v", out)
	}
	m2, out := send(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	if out != nil || m2.View() != m.View() {
		t.Error("x did something")
	}
}

func TestRenameSucceeds(t *testing.T) {
	s := setup(t, map[string]string{"todo": "x"})
	m := open(t, s, "todo")
	m = runes(m, "r")
	if !strings.Contains(m.View(), "Rename: todo") {
		t.Fatalf("prompt not prefilled:\n%s", m.View())
	}
	m = runes(m, "-2026")
	m, out := press(m, tea.KeyEnter)
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
	m, out := press(m, tea.KeyEsc)
	if out != nil || m.state != stateView || !s.Exists("todo") {
		t.Error("esc did not cancel")
	}
	m = runes(m, "r")
	m, out = press(m, tea.KeyEnter)
	if out != nil || m.state != stateView || !s.Exists("todo") {
		t.Error("same name did not cancel")
	}
}

func TestRenameRefused(t *testing.T) {
	s := setup(t, map[string]string{"todo": "T", "ideas": "I"})
	m := runes(open(t, s, "todo"), "r")
	for range "todo" {
		m, _ = press(m, tea.KeyBackspace)
	}
	m = runes(m, "ideas")
	m, out := press(m, tea.KeyEnter)
	if out != nil || m.state != stateRename || !strings.Contains(m.View(), "already exists") {
		t.Errorf("exists not refused: state=%v\n%s", m.state, m.View())
	}
	if c, _ := s.Read("ideas"); c != "I" || !s.Exists("todo") {
		t.Error("files changed")
	}
	// error clears on next key and the prompt stays usable
	for range "ideas" {
		m, _ = press(m, tea.KeyBackspace)
	}
	m = runes(m, "a/b")
	m, out = press(m, tea.KeyEnter)
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
	m, out := send(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if out != nil || m.state != stateView || !s.Exists("todo") {
		t.Error("n did not cancel")
	}
	m = runes(m, "d")
	m, out = press(m, tea.KeyEsc)
	if out != nil || m.state != stateView || !s.Exists("todo") {
		t.Error("esc did not cancel")
	}
	m = runes(m, "d")
	m, out = send(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	if out != (DeletedMsg{}) || s.Exists("todo") || !s.Exists("other") {
		t.Errorf("y: out=%v", out)
	}
}

func TestDeleteFails(t *testing.T) {
	s := setup(t, map[string]string{"todo": "x"})
	m := open(t, s, "todo")
	os.Remove(filepath.Join(s.Dir, "todo.md")) // already gone
	m = runes(m, "d")
	m, out := send(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	if out != nil || m.state != stateView || !strings.Contains(m.View(), "no such file") {
		t.Errorf("delete failure not shown: out=%v\n%s", out, m.View())
	}
}
