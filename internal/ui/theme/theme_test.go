package theme

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"charm.land/glamour/v2"
	"charm.land/lipgloss/v2"
)

func TestStylesRender(t *testing.T) {
	styles := map[string]lipgloss.Style{
		"Normal": Normal, "Selected": Selected, "Create": Create,
		"Hint": Hint, "Error": Error, "Prompt": Prompt,
		"Tag": Tag, "RowTag": RowTag,
	}
	for name, s := range styles {
		if s.Render("x") == "" {
			t.Errorf("%s rendered an empty string", name)
		}
	}
}

func TestTagStyles(t *testing.T) {
	if Tag.GetForeground() == (lipgloss.NoColor{}) {
		t.Error("Tag has no color")
	}
	if RowTag.GetForeground() == (lipgloss.NoColor{}) || !RowTag.GetFaint() {
		t.Error("RowTag is not dimmed")
	}
}

func TestEditorStylesRender(t *testing.T) {
	s := EditorStyles()
	for name, st := range map[string]lipgloss.Style{
		"focused.LineNumber": s.Focused.LineNumber, "focused.CursorLineNumber": s.Focused.CursorLineNumber,
		"focused.Text": s.Focused.Text, "blurred.LineNumber": s.Blurred.LineNumber,
		"focused.Selection": s.Focused.Selection, "blurred.Selection": s.Blurred.Selection,
	} {
		if st.Render("x") == "" {
			t.Errorf("%s rendered an empty string", name)
		}
	}
	if s.Focused.CursorLine.GetBackground() != (lipgloss.NoColor{}) {
		t.Error("cursor line has a background")
	}
	if s.Cursor.Color == nil {
		t.Error("cursor has no color")
	}
	for name, st := range map[string]lipgloss.Style{"focused": s.Focused.Selection, "blurred": s.Blurred.Selection} {
		if st.GetBackground() == (lipgloss.NoColor{}) {
			t.Errorf("%s selection has no background", name)
		}
		if st.GetBackground() == s.Focused.Text.GetBackground() {
			t.Errorf("%s selection background equals the text background", name)
		}
	}
}

func TestInputStyles(t *testing.T) {
	s := InputStyles()
	for name, st := range map[string]lipgloss.Style{
		"focused.Prompt": s.Focused.Prompt, "blurred.Prompt": s.Blurred.Prompt,
	} {
		if st.Render("x") == "" {
			t.Errorf("%s rendered an empty string", name)
		}
	}
	if s.Blurred.Text.GetForeground() != (lipgloss.NoColor{}) {
		t.Error("blurred text has a foreground")
	}
	if s.Focused.Text.GetForeground() != (lipgloss.NoColor{}) {
		t.Error("focused text has a foreground")
	}
	if s.Cursor.Color == nil {
		t.Error("cursor has no color")
	}
}

func TestNoColorLiteralsOutsideTheme(t *testing.T) {
	hex := regexp.MustCompile(`#[0-9a-fA-F]{6}\b`)
	root := filepath.Join("..", "..", "..")
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == "theme" || d.Name() == "openspec" || strings.HasPrefix(d.Name(), ".") && p != root {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(p, ".go") {
			return nil
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		if hex.Match(b) {
			t.Errorf("color literal found in %s", p)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestMarkdownRenders(t *testing.T) {
	r, err := glamour.NewTermRenderer(glamour.WithStyles(Markdown), glamour.WithWordWrap(40))
	if err != nil {
		t.Fatal(err)
	}
	out, err := r.Render("# h\n\n```go\nfmt.Println(1)\n```\n")
	if err != nil || strings.TrimSpace(out) == "" {
		t.Fatalf("render = %q, %v", out, err)
	}
}
