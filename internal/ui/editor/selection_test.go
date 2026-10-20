package editor

import (
	"errors"
	"strings"
	"testing"

	"github.com/llysyganicz/clipt/internal/ui/uitest"
)

// selectLo opens "hello" in an 80x6 terminal and selects its last two
// characters backwards.
func selectLo(t *testing.T) Model {
	t.Helper()
	s := setup(t, map[string]string{"todo": "hello"})
	m := resize(open(t, s, "todo"), 80, 6) // cursor after "hello"
	m, _ = press(m, "shift+left")
	m, _ = press(m, "shift+left")
	return m
}

func TestShiftLeftSelectsBackwards(t *testing.T) {
	m := selectLo(t)
	if got := m.ta.SelectedText(); got != "lo" {
		t.Errorf("selected = %q, want lo", got)
	}
	if c := cursorCol(m); c != 3 {
		t.Errorf("cursor column = %d, want 3", c)
	}
	if got := m.ta.Value(); got != "hello" {
		t.Errorf("value = %q", got)
	}
}

func TestShiftDownSelectsLine(t *testing.T) {
	s := setup(t, map[string]string{"todo": "one\ntwo\nthree"})
	m := open(t, s, "todo")
	m, _ = press(m, "up")
	m, _ = press(m, "home") // start of line 2
	m, _ = press(m, "shift+down")
	if got := m.ta.SelectedText(); got != "two\n" {
		t.Errorf("selected = %q, want %q", got, "two\n")
	}
	if m.ta.Line() != 2 || cursorCol(m) != 0 {
		t.Errorf("cursor at line %d col %d, want line 2 col 0", m.ta.Line(), cursorCol(m))
	}
}

func TestCtrlShiftRightSelectsWord(t *testing.T) {
	s := setup(t, map[string]string{"todo": "one two"})
	m := open(t, s, "todo")
	m, _ = press(m, "home")
	m, _ = press(m, "ctrl+shift+right")
	if got := m.ta.SelectedText(); got != "one" {
		t.Errorf("selected = %q, want one", got)
	}
}

func TestCtrlGSelectsAll(t *testing.T) {
	s := setup(t, map[string]string{"todo": "one\ntwo\nthree"})
	m := open(t, s, "todo")
	m, _ = press(m, "ctrl+g")
	if got := m.ta.SelectedText(); got != "one\ntwo\nthree" {
		t.Errorf("selected = %q", got)
	}
}

func TestMovementClearsSelection(t *testing.T) {
	m := selectLo(t)
	m, _ = press(m, "left")
	if m.ta.HasSelection() {
		t.Errorf("selection kept after left: %q", m.ta.SelectedText())
	}
	if got := m.ta.Value(); got != "hello" {
		t.Errorf("value = %q", got)
	}
}

func TestTypingReplacesSelection(t *testing.T) {
	m := runes(selectLo(t), "p")
	if got := m.ta.Value(); got != "help" {
		t.Errorf("value = %q, want help", got)
	}
	if c := cursorCol(m); c != 4 {
		t.Errorf("cursor column = %d, want 4", c)
	}
	if m.ta.HasSelection() {
		t.Error("selection kept after typing")
	}
}

func TestBackspaceRemovesSelection(t *testing.T) {
	m := selectLo(t)
	m, _ = press(m, "backspace")
	if got := m.ta.Value(); got != "hel" {
		t.Errorf("value = %q, want hel", got)
	}
}

func TestDeleteRemovesSelectionAcrossLines(t *testing.T) {
	s := setup(t, map[string]string{"todo": "ab\ncd"})
	m := open(t, s, "todo")
	m, _ = press(m, "up")
	m, _ = press(m, "home")
	m, _ = press(m, "right") // after "a"
	m, _ = press(m, "shift+down")
	if got := m.ta.SelectedText(); got != "b\nc" {
		t.Fatalf("selected = %q, want %q", got, "b\nc")
	}
	m, _ = press(m, "delete")
	if got := m.ta.Value(); got != "ad" {
		t.Errorf("value = %q, want ad", got)
	}
}

func TestTabReplacesSelection(t *testing.T) {
	m := selectLo(t)
	m, _ = press(m, "tab")
	if got := m.ta.Value(); got != "hel  " {
		t.Errorf("value = %q, want %q", got, "hel  ")
	}
	if c := cursorCol(m); c != 5 {
		t.Errorf("cursor column = %d, want 5", c)
	}
	if m.ta.HasSelection() {
		t.Error("selection kept after tab")
	}
}

func TestInsertReplacesSelection(t *testing.T) {
	m := selectLo(t).Insert("# Standup\n- ")
	if got := m.ta.Value(); got != "hel# Standup\n- " {
		t.Errorf("value = %q", got)
	}
	if m.ta.Line() != 1 || cursorCol(m) != 2 || m.ta.HasSelection() {
		t.Errorf("cursor at line %d col %d selection=%v", m.ta.Line(), cursorCol(m), m.ta.HasSelection())
	}

	m = selectLo(t).InsertTemplate("# Standup\n", []string{"work"})
	if got := m.ta.Value(); got != "---\ntags: [work]\n---\nhel# Standup\n" {
		t.Errorf("value with tags = %q", got)
	}
	if m.ta.HasSelection() {
		t.Error("selection kept after template insertion")
	}
}

func TestSelectAllThenEscDoesNotWrite(t *testing.T) {
	s := setup(t, map[string]string{"todo": "a\nb\n"})
	m := open(t, s, "todo")
	m, _ = press(m, "ctrl+g")
	if _, out := press(m, "esc"); out != (DoneMsg{}) {
		t.Errorf("esc -> %v", out)
	}
	if b, old := file(t, s, "todo"); b != "a\nb\n" || !old {
		t.Errorf("file touched: %q old=%v", b, old)
	}
}

func TestSelectionInTemplate(t *testing.T) {
	ts := setupTemplate(t, map[string]string{"meeting": "ab"})
	m := openTemplate(t, ts, "meeting")
	m, _ = press(m, "shift+left")
	if got := m.ta.SelectedText(); got != "b" {
		t.Errorf("selected = %q, want b", got)
	}
}

func TestSelectionIsHighlighted(t *testing.T) {
	m := selectLo(t)
	plain := uitest.Strip(m.ta.View())
	styled := m.ta.View()
	if !strings.Contains(plain, "hello") {
		t.Fatalf("view lacks the text:\n%s", plain)
	}
	// With a selection the text is split into differently styled runs, so
	// "hello" no longer appears as one unstyled run.
	if strings.Contains(styled, "hello") {
		t.Error("selection drawn without a style change")
	}
}

// copyKey presses ctrl+shift+c without running the command it returns, so
// the test never touches the system clipboard.
func copyKey(m Model) Model {
	m, _ = m.Update(uitest.Key("ctrl+shift+c"))
	return m
}

func TestCopyKeepsSelectionAndContent(t *testing.T) {
	m := selectLo(t)
	m = copyKey(m)
	if got := m.ta.SelectedText(); got != "lo" {
		t.Errorf("selected after copy = %q, want lo", got)
	}
	if got := m.ta.Value(); got != "hello" || cursorCol(m) != 3 {
		t.Errorf("value = %q col %d", got, cursorCol(m))
	}
	if m.err != "" {
		t.Errorf("err = %q", m.err)
	}
}

func TestCopyWithoutSelectionDoesNothing(t *testing.T) {
	s := setup(t, map[string]string{"todo": "hello"})
	m := open(t, s, "todo")
	m, _ = press(m, "left")
	m = copyKey(m)
	if got := m.ta.Value(); got != "hello" || cursorCol(m) != 4 || m.ta.HasSelection() {
		t.Errorf("value = %q col %d selection=%v", got, cursorCol(m), m.ta.HasSelection())
	}
}

func TestTextAreaErrorIsShownAboveStatusBar(t *testing.T) {
	m := selectLo(t)
	m.ta.Err = errors.New("no clipboard utilities found")
	m, _ = m.Update(struct{}{}) // any message the text area receives
	if m.ta.Err != nil {
		t.Error("text area error not cleared")
	}
	lines := strings.Split(view(m), "\n")
	if got := lines[len(lines)-2]; !strings.Contains(got, "no clipboard utilities found") {
		t.Errorf("error line = %q", got)
	}
	if len(lines) != 6 {
		t.Errorf("view height = %d, want 6", len(lines))
	}
	if got := m.ta.SelectedText(); got != "lo" || m.ta.Value() != "hello" {
		t.Errorf("selection or content changed: %q %q", got, m.ta.Value())
	}
	m, _ = press(m, "shift+left")
	if m.err != "" || strings.Contains(view(m), "no clipboard") {
		t.Error("error not cleared by the next key press")
	}
}

func TestCopyHintFollowsSelection(t *testing.T) {
	s := setup(t, map[string]string{"todo": "ab"})
	m := resize(open(t, s, "todo"), 80, 6)
	if strings.Contains(lastLine(view(m)), "ctrl+shift+c copy") {
		t.Errorf("copy hint without a selection: %q", lastLine(view(m)))
	}
	m, _ = press(m, "shift+left")
	if last := lastLine(view(m)); !strings.Contains(last, "ctrl+shift+c copy") || !strings.Contains(last, "ctrl+t template") {
		t.Errorf("bar while selecting = %q", last)
	}
	m, _ = press(m, "right")
	if last := lastLine(view(m)); strings.Contains(last, "ctrl+shift+c copy") {
		t.Errorf("copy hint after clearing = %q", last)
	}

	ts := setupTemplate(t, map[string]string{"meeting": "ab"})
	tm := resize(openTemplate(t, ts, "meeting"), 80, 6)
	tm, _ = press(tm, "shift+left")
	if last := lastLine(view(tm)); !strings.Contains(last, "ctrl+shift+c copy") || !strings.Contains(last, "template: meeting") {
		t.Errorf("template bar while selecting = %q", last)
	}
}
