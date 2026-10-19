package finder

import (
	"strings"
	"unicode"

	"github.com/llysyganicz/clipt/internal/frontmatter"
	"github.com/llysyganicz/clipt/internal/notes"
)

// Row is one line in the finder list: either a note or the create row.
type Row struct {
	Note       *notes.Note // nil for the create row
	Create     string      // name to create; set only for the create row
	CreateTags []string    // tags to create with; create row only
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

// Query is a parsed search query: the name query and the tag filters.
type Query struct {
	Name string
	Tags []string
}

// ParseQuery splits q on whitespace. A token starting with "#" and having
// more characters is a tag filter for the text after the "#"; a bare "#" is
// dropped. The other tokens, joined by single spaces, are the name query.
// Without any "#" token the whole query is the name query, unchanged.
func ParseQuery(q string) Query {
	tokens := strings.Fields(q)
	hasHash := false
	for _, t := range tokens {
		if strings.HasPrefix(t, "#") {
			hasHash = true
			break
		}
	}
	if !hasHash {
		return Query{Name: q}
	}
	var out Query
	var name []string
	for _, t := range tokens {
		switch {
		case t == "#":
		case strings.HasPrefix(t, "#"):
			out.Tags = append(out.Tags, t[1:])
		default:
			name = append(name, t)
		}
	}
	out.Name = strings.Join(name, " ")
	return out
}

// MatchItem reports whether a note or template matches q: the name query
// fuzzy-matches its name and every tag filter is a case-insensitive prefix
// of at least one of its tags.
func MatchItem(q Query, n notes.Note) bool {
	if !Match(q.Name, n.Name) {
		return false
	}
	for _, f := range q.Tags {
		f = strings.ToLower(f)
		found := false
		for _, t := range n.Tags {
			if strings.HasPrefix(strings.ToLower(t), f) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// BuildRows builds the finder rows from all notes (already in mtime order).
// The current note is hidden, matches keep the input order, and a trailing
// create row is added when the trimmed name query is non-empty and no note
// has exactly that name (the hidden current note included). The create row
// carries the query's tag filters, deduplicated.
func BuildRows(all []notes.Note, current *notes.Note, q Query) []Row {
	var rows []Row
	for i := range all {
		n := &all[i]
		if current != nil && n.Name == current.Name {
			continue
		}
		if MatchItem(q, *n) {
			rows = append(rows, Row{Note: n})
		}
	}
	name := strings.TrimSpace(q.Name)
	if name == "" {
		return rows
	}
	for _, n := range all {
		if n.Name == name {
			return rows
		}
	}
	return append(rows, Row{Create: name, CreateTags: frontmatter.Clean(q.Tags)})
}
