package preview

import (
	"strings"
	"testing"

	"github.com/llysyganicz/clipt/internal/ui/uitest"
)

func lines(m Model) []string { return strings.Split(uitest.Strip(m.View()), "\n") }

func TestFrontMatterNotRendered(t *testing.T) {
	s := setup(t, map[string]string{"standup": "---\ntags: [work]\n---\n# Standup\n"})
	v := uitest.Strip(resize(open(t, s, "standup"), 60, 10).View())
	if !strings.Contains(v, "Standup") || strings.Contains(v, "tags:") || strings.Contains(v, "---") {
		t.Errorf("view:\n%s", v)
	}
}

func TestUnclosedFrontMatterRenderedAsBody(t *testing.T) {
	s := setup(t, map[string]string{"draft": "---\ntags: [work]\n# Draft\n"})
	v := uitest.Strip(resize(open(t, s, "draft"), 60, 10).View())
	if !strings.Contains(v, "Draft") || !strings.Contains(v, "tags: [work]") {
		t.Errorf("unclosed front matter is not rendered as body:\n%s", v)
	}
	if !strings.Contains(v, "invalid front matter") {
		t.Errorf("no error header:\n%s", v)
	}
}

func TestTagsHeader(t *testing.T) {
	s := setup(t, map[string]string{"standup": "---\ntags: [work, meeting]\n---\n# Standup\n"})
	m := resize(open(t, s, "standup"), 60, 10)
	l := lines(m)
	if strings.TrimSpace(l[0]) != "#work #meeting" {
		t.Errorf("first line = %q", l[0])
	}
	if len(l) != 10 {
		t.Errorf("view height = %d, want 10", len(l))
	}
	if !strings.Contains(strings.Join(l[1:], "\n"), "Standup") {
		t.Errorf("body not below header:\n%s", strings.Join(l, "\n"))
	}
}

func TestTagsHeaderWraps(t *testing.T) {
	s := setup(t, map[string]string{"n": "---\ntags: [alpha, beta, gamma, delta]\n---\nbody\n"})
	m := resize(open(t, s, "n"), 14, 10)
	l := lines(m)
	if len(l) != 10 {
		t.Fatalf("view height = %d, want 10:\n%s", len(l), strings.Join(l, "\n"))
	}
	if strings.TrimSpace(l[0]) != "#alpha #beta" || strings.TrimSpace(l[1]) != "#gamma #delta" {
		t.Errorf("header = %q / %q", l[0], l[1])
	}
	if m.vp.Height() != 10-1-2 {
		t.Errorf("viewport height = %d", m.vp.Height())
	}
}

func TestNoHeaderWithoutTags(t *testing.T) {
	s := setup(t, map[string]string{
		"ideas": "---\ntitle: Ideas\n---\nfirst\n",
		"plain": "first\n",
	})
	for _, name := range []string{"ideas", "plain"} {
		m := resize(open(t, s, name), 60, 10)
		if m.header != "" || m.vp.Height() != 9 {
			t.Errorf("%s: header %q, viewport height %d", name, m.header, m.vp.Height())
		}
		if !strings.Contains(lines(m)[0], "first") && !strings.Contains(lines(m)[1], "first") {
			t.Errorf("%s: body does not start at the top:\n%s", name, uitest.Strip(m.View()))
		}
	}
}

func TestInvalidFrontMatterHeader(t *testing.T) {
	s := setup(t, map[string]string{"n": "---\ntags: work\n---\nbody\n"})
	m := resize(open(t, s, "n"), 60, 10)
	if l := lines(m); !strings.Contains(l[0], "invalid front matter") {
		t.Errorf("first line = %q", l[0])
	}
}

func TestHeaderStaysWhileScrolling(t *testing.T) {
	s := setup(t, map[string]string{"n": "---\ntags: [work]\n---\n" + longNote()})
	m := resize(open(t, s, "n"), 40, 10)
	m, _ = press(m, "pgdown")
	if m.vp.YOffset() == 0 {
		t.Fatal("pgdown did not scroll")
	}
	if l := lines(m); strings.TrimSpace(l[0]) != "#work" || len(l) != 10 {
		t.Errorf("after pgdown first line = %q, height %d", l[0], len(l))
	}
}
