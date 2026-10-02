package finder

import (
	"strings"
	"unicode"

	"github.com/llysyganicz/clipt/internal/notes"
)

// Row is one line in the finder list: either a note or the create row.
type Row struct {
	Note   *notes.Note // nil for the create row
	Create string      // name to create; set only for the create row
}

// IsCreate reports whether the row is the create row.
func (r Row) IsCreate() bool { return r.Note == nil }

// Match reports whether query is a case-insensitive subsequence of name.
// An empty query matches everything.
func Match(query, name string) bool {
	q := []rune(query)
	if len(q) == 0 {
		return true
	}
	i := 0
	for _, r := range name {
		if unicode.ToLower(r) == unicode.ToLower(q[i]) {
			i++
			if i == len(q) {
				return true
			}
		}
	}
	return false
}

// BuildRows builds the finder rows from all notes (already in mtime order).
// The current note is hidden, matches keep the input order, and a trailing
// create row is added when the trimmed query is non-empty and no note has
// exactly that name (the hidden current note included).
func BuildRows(all []notes.Note, current *notes.Note, query string) []Row {
	var rows []Row
	for i := range all {
		n := &all[i]
		if current != nil && n.Name == current.Name {
			continue
		}
		if Match(query, n.Name) {
			rows = append(rows, Row{Note: n})
		}
	}
	name := strings.TrimSpace(query)
	if name == "" {
		return rows
	}
	for _, n := range all {
		if n.Name == name {
			return rows
		}
	}
	return append(rows, Row{Create: name})
}
