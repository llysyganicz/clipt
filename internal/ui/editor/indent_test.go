package editor

import (
	"fmt"
	"strings"
	"testing"
)

func cursorCol(m Model) int {
	info := m.ta.LineInfo()
	return info.StartColumn + info.ColumnOffset
}

func TestTabInsertsTwoSpaces(t *testing.T) {
	s := setup(t, map[string]string{"todo": "item"})
	m := open(t, s, "todo")
	m, _ = press(m, "home")
	m, _ = press(m, "tab")
	if got := m.ta.Value(); got != "  item" {
		t.Errorf("tab at line start: %q", got)
	}
	if c := cursorCol(m); c != 2 {
		t.Errorf("cursor column after tab = %d, want 2", c)
	}
	m = runes(m, "X")
	if got := m.ta.Value(); got != "  Xitem" {
		t.Errorf("typing after tab: %q", got)
	}

	s = setup(t, map[string]string{"todo": "ab"})
	m = open(t, s, "todo")
	m, _ = press(m, "left") // between a and b
	m, _ = press(m, "tab")
	if got := m.ta.Value(); got != "a  b" {
		t.Errorf("tab inside a line: %q", got)
	}
	if c := cursorCol(m); c != 3 {
		t.Errorf("cursor column = %d, want 3", c)
	}
}

func TestTabIsSaved(t *testing.T) {
	s := setup(t, map[string]string{"todo": "item\n"})
	m := open(t, s, "todo")
	m, _ = press(m, "home")
	m, _ = press(m, "tab")
	if _, out := press(m, "esc"); out != (DoneMsg{}) {
		t.Fatalf("esc -> %v", out)
	}
	if b, old := file(t, s, "todo"); b != "  item\n" || old {
		t.Errorf("file = %q old=%v", b, old)
	}
}

func TestShiftTabDedents(t *testing.T) {
	cases := []struct {
		line      string
		want      string
		cursorEnd int // expected cursor column with the cursor at the line end
	}{
		{"    item", "  item", 6},
		{"  item", "item", 4},
		{" item", "item", 4},
		{"item", "item", 4},
		{"", "", 0},
	}
	for _, c := range cases {
		s := setup(t, map[string]string{"todo": "first\n" + c.line + "\n"})
		m := open(t, s, "todo") // cursor at the end of the second line
		m, _ = press(m, "shift+tab")
		if got := m.ta.Value(); got != "first\n"+c.want {
			t.Errorf("%q: value = %q, want %q", c.line, got, "first\n"+c.want)
		}
		if m.ta.Line() != 1 || cursorCol(m) != c.cursorEnd {
			t.Errorf("%q: cursor at line %d col %d, want line 1 col %d", c.line, m.ta.Line(), cursorCol(m), c.cursorEnd)
		}
	}

	// A cursor inside the removed spaces lands on the line start.
	s := setup(t, map[string]string{"todo": "    item"})
	m := open(t, s, "todo")
	m, _ = press(m, "home")
	m, _ = press(m, "right")
	m, _ = press(m, "shift+tab")
	if got := m.ta.Value(); got != "  item" {
		t.Errorf("value = %q", got)
	}
	if c := cursorCol(m); c != 0 {
		t.Errorf("cursor column = %d, want 0", c)
	}
}

func TestShiftTabIsSavedAndUnchangedIsNot(t *testing.T) {
	s := setup(t, map[string]string{"todo": "  item\n"})
	m := open(t, s, "todo")
	m, _ = press(m, "shift+tab")
	if _, out := press(m, "esc"); out != (DoneMsg{}) {
		t.Fatalf("esc -> %v", out)
	}
	if b, old := file(t, s, "todo"); b != "item\n" || old {
		t.Errorf("file = %q old=%v", b, old)
	}

	s = setup(t, map[string]string{"todo": "item\n"})
	m = open(t, s, "todo")
	m, _ = press(m, "shift+tab")
	if _, out := press(m, "esc"); out != (DoneMsg{}) {
		t.Fatalf("esc -> %v", out)
	}
	if b, old := file(t, s, "todo"); b != "item\n" || !old {
		t.Errorf("file touched: %q old=%v", b, old)
	}
}

func TestShiftTabBelowFirstScreenKeepsView(t *testing.T) {
	var b strings.Builder
	for i := 1; i <= 100; i++ {
		fmt.Fprintf(&b, "  line%d\n", i)
	}
	s := setup(t, map[string]string{"long": b.String()})
	m := resize(open(t, s, "long"), 80, 10) // cursor at the end of line 100
	m, _ = press(m, "shift+tab")
	if got := strings.Split(m.ta.Value(), "\n")[99]; got != "line100" {
		t.Errorf("line 100 = %q", got)
	}
	if m.ta.Line() != 99 {
		t.Errorf("cursor line = %d, want 99", m.ta.Line())
	}
	if v := view(m); !strings.Contains(v, "line100") || strings.Contains(v, "line1 ") {
		t.Errorf("view not on the cursor:\n%s", v)
	}
	lines := strings.Split(view(m), "\n")
	for i, l := range lines[:9] {
		if strings.TrimSpace(l) == "" {
			t.Errorf("row %d blank after dedent", i+1)
		}
	}
}
