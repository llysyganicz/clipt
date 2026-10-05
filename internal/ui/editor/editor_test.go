package editor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/llysyganicz/clipt/internal/notes"
)

var oldTime = time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)

// setup writes the given notes into a temp dir with an old mtime so a test
// can tell whether a note was rewritten.
func setup(t *testing.T, files map[string]string) notes.Store {
	t.Helper()
	dir := t.TempDir()
	for n, c := range files {
		p := filepath.Join(dir, n+".md")
		if err := os.WriteFile(p, []byte(c), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(p, oldTime, oldTime); err != nil {
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

// send delivers msg and returns the message its command produces, if it
// answers at once. Batches (cursor blink) and sleeping commands yield nil.
func send(m Model, msg tea.Msg) (Model, tea.Msg) {
	nm, cmd := m.Update(msg)
	m = nm.(Model)
	if cmd == nil {
		return m, nil
	}
	ch := make(chan tea.Msg, 1)
	go func() { ch <- cmd() }()
	select {
	case out := <-ch:
		if _, ok := out.(tea.BatchMsg); ok {
			return m, nil
		}
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

func lastLine(v string) string {
	l := strings.Split(v, "\n")
	return l[len(l)-1]
}

// file returns the note's bytes and whether its mtime is still the old one.
func file(t *testing.T, s notes.Store, name string) (string, bool) {
	t.Helper()
	p := filepath.Join(s.Dir, name+".md")
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(p)
	if err != nil {
		t.Fatal(err)
	}
	return string(b), info.ModTime().Equal(oldTime)
}

func TestNewMissing(t *testing.T) {
	s := setup(t, nil)
	if _, err := New(s, notes.Note{Name: "gone"}); err == nil {
		t.Fatal("expected error for missing note")
	}
}

func TestCRLFShowsTwoLines(t *testing.T) {
	s := setup(t, map[string]string{"todo": "a\r\nb\r\n"})
	m := open(t, s, "todo")
	if got := m.ta.Value(); got != "a\nb" {
		t.Errorf("value = %q, want %q", got, "a\nb")
	}
	lines := strings.Split(m.View(), "\n")
	if !strings.Contains(lines[0], "a") || !strings.Contains(lines[1], "b") {
		t.Errorf("lines 1-2 = %q, %q", lines[0], lines[1])
	}
	if strings.Contains(lines[0], "b") {
		t.Errorf("b on the first line: %q", lines[0])
	}
	// Still unchanged: the CRLF file is not rewritten without an edit.
	if _, out := press(m, tea.KeyEsc); out != (DoneMsg{}) {
		t.Errorf("esc -> %v", out)
	}
	if b, old := file(t, s, "todo"); b != "a\r\nb\r\n" || !old {
		t.Errorf("file rewritten: %q old=%v", b, old)
	}
}

func TestTypeThenEscWrites(t *testing.T) {
	s := setup(t, map[string]string{"todo": "hello"})
	m := open(t, s, "todo")
	m = runes(m, " world")
	m, out := press(m, tea.KeyEsc)
	if out != (DoneMsg{}) {
		t.Fatalf("esc -> %v", out)
	}
	if b, old := file(t, s, "todo"); b != "hello world\n" || old {
		t.Errorf("file = %q old=%v", b, old)
	}
	if m.err != "" {
		t.Errorf("err = %q", m.err)
	}
}

func TestEscWithoutEditsDoesNotWrite(t *testing.T) {
	s := setup(t, map[string]string{"todo": "a\nb\n"})
	m := open(t, s, "todo")
	m, _ = press(m, tea.KeyUp)
	m, _ = press(m, tea.KeyDown)
	if _, out := press(m, tea.KeyEsc); out != (DoneMsg{}) {
		t.Errorf("esc -> %v", out)
	}
	if b, old := file(t, s, "todo"); b != "a\nb\n" || !old {
		t.Errorf("file touched: %q old=%v", b, old)
	}
}

func TestTypeThenDeleteDoesNotWrite(t *testing.T) {
	s := setup(t, map[string]string{"todo": "a\n"})
	m := runes(open(t, s, "todo"), "x")
	m, _ = press(m, tea.KeyBackspace)
	if _, out := press(m, tea.KeyEsc); out != (DoneMsg{}) {
		t.Errorf("esc -> %v", out)
	}
	if b, old := file(t, s, "todo"); b != "a\n" || !old {
		t.Errorf("file touched: %q old=%v", b, old)
	}
}

func TestMissingTrailingNewlineNotRewritten(t *testing.T) {
	s := setup(t, map[string]string{"todo": "a"})
	m := open(t, s, "todo")
	if _, out := press(m, tea.KeyEsc); out != (DoneMsg{}) {
		t.Errorf("esc -> %v", out)
	}
	if b, old := file(t, s, "todo"); b != "a" || !old {
		t.Errorf("file touched: %q old=%v", b, old)
	}
}

func TestTabs(t *testing.T) {
	s := setup(t, map[string]string{"todo": "\tcode\nplain\n"})
	m := open(t, s, "todo")
	if _, out := press(m, tea.KeyEsc); out != (DoneMsg{}) {
		t.Errorf("esc -> %v", out)
	}
	if b, old := file(t, s, "todo"); b != "\tcode\nplain\n" || !old {
		t.Errorf("file touched without edit: %q old=%v", b, old)
	}
	m = runes(m, "!") // cursor is at the end of "plain"
	if _, out := press(m, tea.KeyEsc); out != (DoneMsg{}) {
		t.Errorf("esc -> %v", out)
	}
	if b, _ := file(t, s, "todo"); b != "    code\nplain!\n" {
		t.Errorf("file = %q", b)
	}
}

func TestTrailingBlankLinesCollapse(t *testing.T) {
	s := setup(t, map[string]string{"todo": "x\n"})
	m := runes(open(t, s, "todo"), "done")
	for i := 0; i < 3; i++ {
		m, _ = press(m, tea.KeyEnter)
	}
	if _, out := press(m, tea.KeyEsc); out != (DoneMsg{}) {
		t.Errorf("esc -> %v", out)
	}
	if b, _ := file(t, s, "todo"); b != "xdone\n" {
		t.Errorf("file = %q", b)
	}
}

func TestWriteFailureKeepsContentAndRetries(t *testing.T) {
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

	m, out := press(m, tea.KeyEsc)
	if out != nil {
		t.Fatalf("esc on failure -> %v", out)
	}
	if m.err == "" || !strings.Contains(m.View(), "permission denied") {
		t.Errorf("error not shown: err=%q\n%s", m.err, m.View())
	}
	if got := m.ta.Value(); got != "ab" {
		t.Errorf("content lost: %q", got)
	}
	if b, old := file(t, s, "todo"); b != "a\n" || !old {
		t.Errorf("file touched: %q old=%v", b, old)
	}

	// The error disappears on the next key press.
	m = runes(m, "c")
	if m.err != "" || strings.Contains(m.View(), "permission denied") {
		t.Error("error not cleared by a key press")
	}

	if err := os.Chmod(s.Dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(p, 0o644); err != nil {
		t.Fatal(err)
	}
	m, out = press(m, tea.KeyEsc)
	if out != (DoneMsg{}) || m.err != "" {
		t.Fatalf("retry -> %v err=%q", out, m.err)
	}
	if b, _ := file(t, s, "todo"); b != "abc\n" {
		t.Errorf("file = %q", b)
	}
}

func TestRemovedFileIsRecreated(t *testing.T) {
	s := setup(t, map[string]string{"todo": "a\n"})
	m := open(t, s, "todo")
	if err := os.Remove(filepath.Join(s.Dir, "todo.md")); err != nil {
		t.Fatal(err)
	}
	m = runes(m, "b")
	if _, out := press(m, tea.KeyEsc); out != (DoneMsg{}) {
		t.Errorf("esc -> %v", out)
	}
	if b, _ := file(t, s, "todo"); b != "ab\n" {
		t.Errorf("file = %q", b)
	}
}

func TestCtrlCQuitsWithoutWriting(t *testing.T) {
	s := setup(t, map[string]string{"todo": "a\n"})
	m := runes(open(t, s, "todo"), "zzz")
	if _, out := press(m, tea.KeyCtrlC); out != tea.Msg(tea.QuitMsg{}) {
		t.Errorf("ctrl+c -> %v", out)
	}
	if b, old := file(t, s, "todo"); b != "a\n" || !old {
		t.Errorf("file touched: %q old=%v", b, old)
	}
}

func TestEditingKeys(t *testing.T) {
	s := setup(t, map[string]string{"todo": "one\ntwo\nthree"})
	m := open(t, s, "todo")
	m, _ = press(m, tea.KeyEnter)
	m = runes(m, "four")
	if got := m.ta.Value(); got != "one\ntwo\nthree\nfour" {
		t.Errorf("after enter: %q", got)
	}
	before := m.ta.Value()
	m, _ = press(m, tea.KeyUp)
	if m.ta.Line() != 2 || m.ta.Value() != before {
		t.Errorf("up: line=%d value=%q", m.ta.Line(), m.ta.Value())
	}
	m, _ = press(m, tea.KeyHome)
	m, _ = press(m, tea.KeyDelete)
	if got := m.ta.Value(); got != "one\ntwo\nhree\nfour" {
		t.Errorf("delete: %q", got)
	}
	m, _ = press(m, tea.KeyEnd)
	m, _ = press(m, tea.KeyBackspace)
	if got := m.ta.Value(); got != "one\ntwo\nhre\nfour" {
		t.Errorf("backspace: %q", got)
	}
}

func TestNewNoteStartsOnLineOne(t *testing.T) {
	s := setup(t, map[string]string{"ideas": ""})
	m := open(t, s, "ideas")
	if got := strings.Split(m.View(), "\n")[0]; !strings.Contains(got, "1") {
		t.Errorf("first line lacks line number 1: %q", got)
	}
	m = runes(m, "x")
	if _, out := press(m, tea.KeyEsc); out != (DoneMsg{}) {
		t.Error("esc")
	}
	if b, _ := file(t, s, "ideas"); b != "x\n" {
		t.Errorf("file = %q", b)
	}
}

func TestViewStatusBarAndError(t *testing.T) {
	s := setup(t, map[string]string{"a-rather-long-note-name": "x"})
	m := resize(open(t, s, "a-rather-long-note-name"), 120, 6)
	last := lastLine(m.View())
	if !strings.Contains(last, "a-rather-long-note-name") || !strings.Contains(last, "esc save & back") {
		t.Errorf("wide bar = %q", last)
	}
	if got := strings.Count(m.View(), "\n") + 1; got != 6 {
		t.Errorf("view height = %d, want 6", got)
	}
	m = resize(m, 36, 6)
	last = lastLine(m.View())
	if !strings.Contains(last, "esc save & back") || strings.Contains(last, "a-rather-long-note-name") || !strings.Contains(last, "…") {
		t.Errorf("narrow bar = %q", last)
	}
	if lipgloss.Width(last) > 36 {
		t.Errorf("narrow bar too wide: %d", lipgloss.Width(last))
	}

	m.err = "boom"
	lines := strings.Split(m.View(), "\n")
	if got := lines[len(lines)-2]; !strings.Contains(got, "boom") {
		t.Errorf("error line = %q", got)
	}
	if len(lines) != 6 {
		t.Errorf("view height with error = %d, want 6", len(lines))
	}
}

func TestLongNoteShowsLastLineOnEntry(t *testing.T) {
	var b strings.Builder
	for i := 1; i <= 100; i++ {
		fmt.Fprintf(&b, "line%d\n", i)
	}
	s := setup(t, map[string]string{"long": b.String()})
	m := open(t, s, "long") // 80x24 default
	if v := m.View(); !strings.Contains(v, "line100") || strings.Contains(v, "line1 ") {
		t.Errorf("view not scrolled to the end:\n%s", v)
	}
	m = resize(m, 80, 200)
	if v := m.View(); !strings.Contains(v, "line1 ") || !strings.Contains(v, "line100") {
		t.Errorf("tall view lacks first or last line:\n%s", v)
	}
	if got := strings.Count(m.View(), "\n") + 1; got != 200 {
		t.Errorf("view height = %d, want 200", got)
	}
	m = resize(m, 80, 10)
	if v := m.View(); !strings.Contains(v, "line100") || strings.Contains(v, "line90 ") {
		t.Errorf("short view not scrolled to the cursor:\n%s", v)
	}
}

func TestResizeKeepsCursorAndFillsView(t *testing.T) {
	var b strings.Builder
	for i := 1; i <= 100; i++ {
		fmt.Fprintf(&b, "line%d\n", i)
	}
	s := setup(t, map[string]string{"long": b.String()})
	m := resize(open(t, s, "long"), 80, 24) // scrolled to the end
	for i := 0; i < 50; i++ {
		m, _ = press(m, tea.KeyUp)
	}
	m, _ = press(m, tea.KeyLeft) // line 50, before its last character
	m = resize(m, 80, 60)
	if m.ta.Line() != 49 {
		t.Errorf("cursor line after resize = %d, want 49", m.ta.Line())
	}
	m = runes(m, "X")
	if got := strings.Split(m.ta.Value(), "\n")[49]; got != "line5X0" {
		t.Errorf("line 50 after typing = %q", got)
	}
	lines := strings.Split(m.View(), "\n")
	if len(lines) != 60 {
		t.Fatalf("view height = %d, want 60", len(lines))
	}
	for i, l := range lines[:59] {
		if strings.TrimSpace(l) == "" {
			t.Errorf("row %d is blank after growing the terminal", i+1)
		}
	}
	if !strings.Contains(m.View(), "line5X0") {
		t.Error("cursor line not visible after resize")
	}
}

func TestManyLines(t *testing.T) {
	var b strings.Builder
	for i := 1; i <= 5000; i++ {
		fmt.Fprintf(&b, "l%d\n", i)
	}
	s := setup(t, map[string]string{"big": b.String()})
	m, _ := press(open(t, s, "big"), tea.KeyEnter)
	m = runes(m, "l5001")
	if n := m.ta.LineCount(); n != 5001 {
		t.Errorf("lines = %d, want 5001", n)
	}
	if _, out := press(m, tea.KeyEsc); out != (DoneMsg{}) {
		t.Error("esc")
	}
	if got, _ := file(t, s, "big"); !strings.HasSuffix(got, "l5000\nl5001\n") {
		t.Errorf("file tail = %q", got[len(got)-20:])
	}
}
