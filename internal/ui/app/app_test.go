package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/llysyganicz/clipt/internal/notes"
	"github.com/llysyganicz/clipt/internal/ui/finder"
	"github.com/llysyganicz/clipt/internal/ui/picker"
)

func mkdir(t *testing.T, names ...string) string {
	t.Helper()
	dir := t.TempDir()
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for i, n := range names { // later names are newer
		p := filepath.Join(dir, n+".md")
		if err := os.WriteFile(p, nil, 0o644); err != nil {
			t.Fatal(err)
		}
		mt := base.Add(time.Duration(i) * time.Hour)
		os.Chtimes(p, mt, mt)
	}
	return dir
}

// feed delivers msg and then any message its command produces (one level).
func feed(m Model, msg tea.Msg) (Model, tea.Msg) {
	nm, cmd := m.Update(msg)
	m = nm.(Model)
	return m, run(cmd)
}

// press sends a key and re-delivers resulting app-level messages until quiet.
func press(m Model, keys ...tea.KeyType) (Model, bool) {
	quit := false
	for _, k := range keys {
		msg := tea.Msg(tea.KeyMsg{Type: k})
		for msg != nil {
			var out tea.Msg
			m, out = feed(m, msg)
			if _, ok := out.(tea.QuitMsg); ok {
				quit = true
			}
			msg = out
			if _, ok := msg.(tea.QuitMsg); ok {
				msg = nil
			}
		}
	}
	return m, quit
}

func TestSwitcherToggle(t *testing.T) {
	dir := mkdir(t, "c", "b", "a") // a newest
	m := New(dir)
	m, _ = press(m, tea.KeyEnter) // open a
	if m.mode != modeNote || m.current.Name != "a" {
		t.Fatalf("mode=%v current=%v", m.mode, m.current)
	}
	m, _ = press(m, tea.KeyEsc) // back to finder
	if m.mode != modeFinder {
		t.Fatalf("mode = %v", m.mode)
	}
	rows := m.finder.Rows()
	if rows[0].Note.Name != "b" {
		t.Errorf("top row = %q, want b", rows[0].Note.Name)
	}
	m, _ = press(m, tea.KeyEnter) // open b; a now next
	if m.current.Name != "b" {
		t.Errorf("current = %q", m.current.Name)
	}
}

func TestEscQuitsWithoutCurrent(t *testing.T) {
	m := New(mkdir(t, "a"))
	_, quit := press(m, tea.KeyEsc)
	if !quit {
		t.Error("esc without current note did not quit")
	}
}

func TestEscWithCurrentShowsPreview(t *testing.T) {
	m := New(mkdir(t, "a", "b"))
	m, _ = press(m, tea.KeyEnter)
	m, _ = press(m, tea.KeyEsc)     // note -> finder
	m, quit := press(m, tea.KeyEsc) // finder -> preview of current
	if quit || m.mode != modeNote || m.current == nil {
		t.Errorf("mode=%v quit=%v", m.mode, quit)
	}
}

func TestCurrentDisappeared(t *testing.T) {
	dir := mkdir(t, "a", "b")
	m := New(dir)
	m, _ = press(m, tea.KeyEnter) // open b
	m, _ = press(m, tea.KeyEsc)
	os.Remove(filepath.Join(dir, "b.md"))
	m, quit := press(m, tea.KeyEsc)
	if quit || m.mode != modeFinder || m.current != nil {
		t.Fatalf("mode=%v current=%v quit=%v", m.mode, m.current, quit)
	}
	if v := m.View(); !strings.Contains(v, "no longer exists") {
		t.Errorf("view lacks error:\n%s", v)
	}
	if _, quit := press(m, tea.KeyEsc); !quit {
		t.Error("esc after clearing current should quit")
	}
}

func TestCreateOpensNote(t *testing.T) {
	dir := mkdir(t)
	m := New(dir)
	nm, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("new")})
	m = nm.(Model)
	m, _ = press(m, tea.KeyEnter)
	if m.mode != modeNote || m.current == nil || m.current.Name != "new" {
		t.Errorf("mode=%v current=%v", m.mode, m.current)
	}
	if _, err := os.Stat(filepath.Join(dir, "new.md")); err != nil {
		t.Error(err)
	}
}

func TestPickerStartupCancelQuits(t *testing.T) {
	m := NewPicker("", "")
	_, quit := press(m, tea.KeyEsc)
	if !quit {
		t.Error("esc at startup picker did not quit")
	}
}

func TestPickerStartupAccept(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	dir := mkdir(t, "a")
	m, _ := feed(NewPicker("", ""), picker.DirAcceptedMsg{Dir: dir})
	if m.mode != modeFinder || m.dir != dir || len(m.finder.Rows()) != 1 {
		t.Errorf("mode=%v dir=%q rows=%d", m.mode, m.dir, len(m.finder.Rows()))
	}
}

func openPickerFromFinder(t *testing.T, m Model) Model {
	t.Helper()
	m, _ = press(m, tea.KeyCtrlO)
	if m.mode != modePicker || m.pickerFrom != picker.Finder {
		t.Fatalf("mode=%v from=%v", m.mode, m.pickerFrom)
	}
	return m
}

func TestPickerEscRestoresFinder(t *testing.T) {
	m := New(mkdir(t, "a", "b"))
	m, _ = press(m, tea.KeyEnter)
	m, _ = press(m, tea.KeyEsc)
	nm, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("zz")})
	m = nm.(Model)
	cur := m.current
	m = openPickerFromFinder(t, m)
	m, quit := press(m, tea.KeyEsc)
	if quit || m.mode != modeFinder || m.current != cur || m.finder.Query() != "zz" {
		t.Errorf("mode=%v quit=%v query=%q", m.mode, quit, m.finder.Query())
	}
}

func TestDirChangeClearsCurrent(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	m := New(mkdir(t, "a", "b"))
	m, _ = press(m, tea.KeyEnter)
	m, _ = press(m, tea.KeyEsc)
	m = openPickerFromFinder(t, m)
	other := mkdir(t, "x")
	m, _ = feed(m, picker.DirAcceptedMsg{Dir: other})
	if m.current != nil || m.dir != other || m.finder.Query() != "" {
		t.Errorf("current=%v dir=%q query=%q", m.current, m.dir, m.finder.Query())
	}
	if _, quit := press(m, tea.KeyEsc); !quit {
		t.Error("esc should quit after dir change")
	}
}

func TestSameDirKeepsCurrent(t *testing.T) {
	dir := mkdir(t, "a", "b")
	m := New(dir)
	m, _ = press(m, tea.KeyEnter)
	m, _ = press(m, tea.KeyEsc)
	m = openPickerFromFinder(t, m)
	m, _ = feed(m, picker.DirAcceptedMsg{Dir: dir})
	if m.current == nil || m.current.Name != "b" {
		t.Fatalf("current = %v", m.current)
	}
	for _, r := range m.finder.Rows() {
		if !r.IsCreate() && r.Note.Name == "b" {
			t.Error("current note should be hidden")
		}
	}
}

func TestOpenAndCreatedSetCurrent(t *testing.T) {
	m := New(mkdir(t))
	m, _ = feed(m, finder.OpenNoteMsg{Note: notes.Note{Name: "o"}})
	if m.current.Name != "o" {
		t.Error("OpenNoteMsg")
	}
	m, _ = feed(m, finder.NoteCreatedMsg{Note: notes.Note{Name: "c"}})
	if m.current.Name != "c" {
		t.Error("NoteCreatedMsg")
	}
}

func TestNewPickerPrefill(t *testing.T) {
	if v := NewPicker("", "").View(); !strings.Contains(v, "~/notes") {
		t.Errorf("first run view lacks ~/notes:\n%s", v)
	}
	v := NewPicker("/mnt/gone", "directory does not exist: /mnt/gone").View()
	if !strings.Contains(v, "/mnt/gone") || !strings.Contains(v, "does not exist") {
		t.Errorf("stale view:\n%s", v)
	}
}

// run executes a command, giving up quickly on blocking ones (cursor blink).
func run(cmd tea.Cmd) tea.Msg {
	if cmd == nil {
		return nil
	}
	ch := make(chan tea.Msg, 1)
	go func() { ch <- cmd() }()
	select {
	case m := <-ch:
		if _, ok := m.(tea.BatchMsg); ok {
			return nil
		}
		return m
	case <-time.After(20 * time.Millisecond):
		return nil
	}
}
