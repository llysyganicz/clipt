package theme

import (
	"fmt"
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
	if s.Cursor.Color == nil {
		t.Error("cursor has no color")
	}
}

// withLight applies gruvbox_light for the test and restores gruvbox_dark
// after it.
func withLight(t *testing.T) Palette {
	t.Helper()
	light := mustBuiltin("gruvbox_light")
	t.Cleanup(func() { Apply(mustBuiltin("gruvbox_dark")) })
	Apply(light)
	return light
}

func TestBackgroundForegroundFollowPalette(t *testing.T) {
	light := withLight(t)
	if got := Background(); got != lipgloss.Color(light.Background) {
		t.Errorf("Background() = %v, want %s", got, light.Background)
	}
	if got := Foreground(); got != lipgloss.Color(light.Foreground) {
		t.Errorf("Foreground() = %v, want %s", got, light.Foreground)
	}
}

func TestNormalHasNoBackground(t *testing.T) {
	t.Cleanup(func() { Apply(mustBuiltin("gruvbox_dark")) })
	for _, name := range Builtin() {
		p := mustBuiltin(name)
		Apply(p)
		out := Normal.Render("x")
		if !strings.Contains(out, uitestFg(p.Foreground)) {
			t.Errorf("%s: Normal.Render = %q, lacks foreground %s", name, out, p.Foreground)
		}
		if strings.Contains(out, "48;2;") {
			t.Errorf("%s: Normal.Render = %q, sets a background", name, out)
		}
		if Normal.GetBackground() != (lipgloss.NoColor{}) {
			t.Errorf("%s: Normal has background %v", name, Normal.GetBackground())
		}
	}
}

func TestInputTextUsesForeground(t *testing.T) {
	light := withLight(t)
	s := InputStyles()
	for name, st := range map[string]lipgloss.Style{"focused": s.Focused.Text, "blurred": s.Blurred.Text} {
		if fg := st.GetForeground(); fg != lipgloss.Color(light.Foreground) {
			t.Errorf("%s text foreground = %v, want %s", name, fg, light.Foreground)
		}
		if out := st.Render("x"); !strings.Contains(out, uitestFg(light.Foreground)) {
			t.Errorf("%s text renders as %q, want foreground %s", name, out, light.Foreground)
		}
	}
}

// uitestFg returns the SGR parameters that set a truecolor foreground to
// the "#rrggbb" color hex.
func uitestFg(hex string) string {
	var r, g, b int
	fmt.Sscanf(hex, "#%02x%02x%02x", &r, &g, &b)
	return fmt.Sprintf("38;2;%d;%d;%d", r, g, b)
}

func TestNoColorLiteralsOutsideTheme(t *testing.T) {
	hex := regexp.MustCompile(`#[0-9a-fA-F]{6}\b`)
	root := filepath.Join("..", "..", "..")
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == "openspec" || strings.HasPrefix(d.Name(), ".") && p != root {
				return filepath.SkipDir
			}
			return nil
		}
		// The palette lives in TOML files; Go code, this package included,
		// must not hard-code colors. Test files may name colors to check.
		if !strings.HasSuffix(p, ".go") || strings.HasSuffix(p, "_test.go") {
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
