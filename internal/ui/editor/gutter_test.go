package editor

import (
	"strings"
	"testing"

	"charm.land/bubbles/v2/textarea"
)

// TestWrapRowsMatchTextArea guards the copied wrap function: for lines that
// wrap, the row count must equal the number of rows the text area itself
// produces, which it reports for the cursor's line through LineInfo.
func TestWrapRowsMatchTextArea(t *testing.T) {
	cases := []string{
		"",
		"short",
		"  - name: title",
		strings.Repeat("word ", 30),
		strings.Repeat("x", 95),
		"  " + strings.Repeat("ab ", 40) + "end",
		"日本語 " + strings.Repeat("漢字", 30),
		"trailing spaces    " + strings.Repeat("a ", 25),
	}
	for _, width := range []int{20, 37, 80} {
		for _, line := range cases {
			ta := textarea.New()
			ta.ShowLineNumbers = false
			ta.Prompt = ""
			ta.MaxHeight = 0
			ta.Focus()
			ta.SetValue(line) // cursor ends on the last row of the line
			ta.SetWidth(width)
			ta.SetHeight(50)
			got := len(wrap([]rune(line), width))
			want := ta.LineInfo().RowOffset + 1
			if got != want {
				t.Errorf("width %d, line %q: wrap rows = %d, text area rows = %d", width, line, got, want)
			}
		}
	}
}

func numberedNote(n int, prefix string) string {
	var b strings.Builder
	for i := 1; i <= n; i++ {
		b.WriteString(prefix + "- item\n")
	}
	return b.String()
}

func TestGutterRows(t *testing.T) {
	s := setup(t, map[string]string{"n12": numberedNote(12, "  "), "n120": numberedNote(120, "")})
	m := open(t, s, "n12")
	if m.gutterW != 4 {
		t.Errorf("12 lines: gutter width = %d, want 4", m.gutterW)
	}
	rows := m.gutterRows()
	if len(rows) != 12 {
		t.Fatalf("12 lines: %d rows", len(rows))
	}
	for i, r := range rows {
		if r.num != i+1 {
			t.Errorf("row %d: num = %d", i, r.num)
		}
		if r.cursor != (i == 11) {
			t.Errorf("row %d: cursor = %v (cursor is on the last line)", i, r.cursor)
		}
	}
	m, _ = press(m, "up")
	if rows := m.gutterRows(); !rows[10].cursor || rows[11].cursor {
		t.Errorf("after up: cursor flags = %v %v", rows[10].cursor, rows[11].cursor)
	}

	m = open(t, s, "n120")
	if m.gutterW != 5 {
		t.Errorf("120 lines: gutter width = %d, want 5", m.gutterW)
	}
	if rows := m.gutterRows(); len(rows) != 120 || rows[99].num != 100 {
		t.Errorf("120 lines: %d rows, row 100 = %+v", len(rows), rows[99])
	}

	// A wrapped line takes several rows: the number on the first, blanks
	// after it, all flagged as the cursor's line.
	s = setup(t, map[string]string{"w": "top\n" + strings.Repeat("word ", 30) + "\nend"})
	m = resize(open(t, s, "w"), 40, 10)
	m, _ = press(m, "up")
	rows = m.gutterRows()
	want := len(wrap([]rune(strings.Repeat("word ", 30)), m.ta.Width()))
	if want < 2 {
		t.Fatalf("test line does not wrap at width %d", m.ta.Width())
	}
	if len(rows) != want+2 {
		t.Fatalf("rows = %d, want %d", len(rows), want+2)
	}
	if rows[1].num != 2 || !rows[1].cursor {
		t.Errorf("first row of the wrapped line = %+v", rows[1])
	}
	for i := 2; i < 1+want; i++ {
		if rows[i].num != 0 || !rows[i].cursor {
			t.Errorf("continuation row %d = %+v", i, rows[i])
		}
	}
	if rows[want+1].num != 3 || rows[want+1].cursor {
		t.Errorf("row after the wrapped line = %+v", rows[want+1])
	}
}

// textCol returns the column at which the given text starts on a rendered
// row, or -1.
func textCol(row, text string) int { return strings.Index(row, text) }

func TestGutterAlignsText(t *testing.T) {
	s := setup(t, map[string]string{"n12": numberedNote(12, "  "), "n120": numberedNote(120, "  ")})
	rows := strings.Split(view(resize(open(t, s, "n12"), 80, 20)), "\n")
	if a, b := textCol(rows[8], "- item"), textCol(rows[9], "- item"); a < 0 || a != b {
		t.Errorf("line 9 text at column %d, line 10 at %d:\n%s\n%s", a, b, rows[8], rows[9])
	}
	if !strings.HasPrefix(rows[8], "  9 ") || !strings.HasPrefix(rows[9], " 10 ") {
		t.Errorf("gutters: %q / %q", rows[8][:4], rows[9][:4])
	}

	rows = strings.Split(view(resize(open(t, s, "n120"), 80, 200)), "\n")
	if a, b := textCol(rows[98], "- item"), textCol(rows[99], "- item"); a < 0 || a != b {
		t.Errorf("line 99 text at column %d, line 100 at %d:\n%s\n%s", a, b, rows[98], rows[99])
	}
	if !strings.HasPrefix(rows[98], "  99 ") || !strings.HasPrefix(rows[99], " 100 ") {
		t.Errorf("gutters: %q / %q", rows[98][:5], rows[99][:5])
	}

	// Both rows of a wrapped line start their text in the same column.
	s = setup(t, map[string]string{"w": strings.Repeat("word ", 12) + "\nend"})
	m := resize(open(t, s, "w"), 30, 10)
	rows = strings.Split(view(m), "\n")
	if textCol(rows[0], "word") != m.gutterW || textCol(rows[1], "word") != m.gutterW {
		t.Errorf("wrapped rows:\n%q\n%q", rows[0], rows[1])
	}
	n := len(wrap([]rune(strings.Repeat("word ", 12)), m.ta.Width()))
	if n < 2 {
		t.Fatalf("test line does not wrap at width %d", m.ta.Width())
	}
	if !strings.HasPrefix(rows[0], " 1 ") || !strings.HasPrefix(rows[1], "   ") || !strings.HasPrefix(rows[n], " 2 ") {
		t.Errorf("wrapped gutters:\n%q\n%q\n%q", rows[0], rows[1], rows[n])
	}
}

func TestGutterGrowsWithTheNote(t *testing.T) {
	s := setup(t, map[string]string{"n9": strings.TrimSuffix(numberedNote(9, ""), "\n")})
	m := resize(open(t, s, "n9"), 80, 20)
	rows := strings.Split(view(m), "\n")
	if m.gutterW != 3 || textCol(rows[0], "- item") != 3 {
		t.Fatalf("9 lines: width %d, row 1 = %q", m.gutterW, rows[0])
	}
	m, _ = press(m, "enter")
	rows = strings.Split(view(m), "\n")
	if m.gutterW != 4 {
		t.Errorf("10 lines: width %d", m.gutterW)
	}
	for i := 0; i < 9; i++ {
		if textCol(rows[i], "- item") != 4 {
			t.Errorf("row %d after growing: %q", i+1, rows[i])
		}
	}
	if !strings.HasPrefix(rows[9], " 10 ") {
		t.Errorf("row 10 = %q", rows[9])
	}
	m, _ = press(m, "backspace")
	if m.gutterW != 3 {
		t.Errorf("back to 9 lines: width %d", m.gutterW)
	}
}
