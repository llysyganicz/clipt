package theme

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

func TestStylesRender(t *testing.T) {
	styles := map[string]lipgloss.Style{
		"Normal": Normal, "Selected": Selected, "Create": Create,
		"Hint": Hint, "Error": Error, "Prompt": Prompt,
	}
	for name, s := range styles {
		if s.Render("x") == "" {
			t.Errorf("%s rendered an empty string", name)
		}
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
