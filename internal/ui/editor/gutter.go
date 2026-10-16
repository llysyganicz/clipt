package editor

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/llysyganicz/clipt/internal/ui/theme"
)

// The text area's own line numbers are padded to the digit count of its
// MaxHeight, which the editor sets to 0 to lift the line limit, so their
// width would change at line 10 and line 100 and shift the text. The editor
// draws the gutter itself instead, through the text area's per-row prompt
// callback: a line number on the first visual row of each logical line, a
// blank on soft-wrapped continuation rows, all padded to the digit count of
// the note's line count.

// gutterRow describes one visual row of the gutter: the line number on the
// first row of a logical line (0 on a continuation row) and whether the row
// belongs to the cursor's line.
type gutterRow struct {
	num    int
	cursor bool
}

// gutterWidth returns the gutter width for a note with lines lines: the
// digits of the line count plus one space on each side.
func gutterWidth(lines int) int {
	return len(strconv.Itoa(lines)) + 2
}

// gutterRows computes the gutter for every visual row of the text area by
// wrapping each logical line at the text area's width.
func (m Model) gutterRows() []gutterRow {
	lines := strings.Split(m.ta.Value(), "\n")
	width := m.ta.Width()
	cur := m.ta.Line()
	var rows []gutterRow
	for i, line := range lines {
		rows = append(rows, gutterRow{num: i + 1, cursor: i == cur})
		for n := len(wrap([]rune(line), width)); n > 1; n-- {
			rows = append(rows, gutterRow{cursor: i == cur})
		}
	}
	return rows
}

// gutterLabels renders gutterRows into one styled string per visual row,
// each exactly m.gutterW columns wide.
func (m Model) gutterLabels() []string {
	rows := m.gutterRows()
	labels := make([]string, len(rows))
	for i, r := range rows {
		if r.num == 0 {
			labels[i] = strings.Repeat(" ", m.gutterW)
			continue
		}
		style := theme.LineNumber
		if r.cursor {
			style = theme.CursorLineNumber
		}
		labels[i] = style.Render(fmt.Sprintf(" %*d ", m.gutterW-2, r.num))
	}
	return labels
}

// syncGutterWidth re-reserves the gutter columns in the text area when the
// line count gained or lost a digit, re-wrapping the text at the new width.
// It reports whether the width changed.
func (m *Model) syncGutterWidth() bool {
	w := gutterWidth(m.ta.LineCount())
	if w == m.gutterW {
		return false
	}
	m.gutterW = w
	// The prompt callback set here only reserves the width; View installs
	// the one that draws the labels, on its own copy of the text area.
	m.ta.SetPromptFunc(w, func(int) string { return "" })
	if m.width > 0 {
		m.ta.SetWidth(m.width)
	}
	return true
}
