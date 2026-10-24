package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setup(t *testing.T) string {
	t.Helper()
	xdg := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdg)
	return xdg
}

func writeConfig(t *testing.T, xdg, content string) string {
	t.Helper()
	p := filepath.Join(xdg, "clipt", "config.toml")
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestPath(t *testing.T) {
	xdg := setup(t)
	got, err := Path()
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(xdg, "clipt", "config.toml"); got != want {
		t.Errorf("Path() = %q, want %q", got, want)
	}
}

func TestLoad(t *testing.T) {
	notes := t.TempDir()
	file := filepath.Join(t.TempDir(), "afile")
	if err := os.WriteFile(file, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	gone := filepath.Join(t.TempDir(), "gone")

	tests := []struct {
		name    string
		content *string
		dir     string
		theme   string
		status  Status
		errSub  string
	}{
		{name: "no file", status: MissingFile},
		{name: "empty notes_dir", content: ptr(`notes_dir = ""`), status: EmptyDir},
		{name: "absent notes_dir", content: ptr(`editor = "vim"`), status: EmptyDir},
		{name: "unknown keys ignored", content: ptr("notes_dir = \"" + notes + "\"\neditor = \"vim\"\n"), dir: notes, status: OK},
		{name: "no theme", content: ptr("notes_dir = \"" + notes + "\"\n"), dir: notes, status: OK},
		{name: "theme", content: ptr("notes_dir = \"" + notes + "\"\ntheme = \"gruvbox_light\"\n"), dir: notes, theme: "gruvbox_light", status: OK},
		{name: "theme with bad dir", content: ptr("notes_dir = \"" + gone + "\"\ntheme = \"gruvbox_light\"\n"), dir: gone, theme: "gruvbox_light", status: BadDir, errSub: "does not exist"},
		{name: "path is a file", content: ptr("notes_dir = \"" + file + "\""), dir: file, status: BadDir, errSub: "not a directory"},
		{name: "path missing", content: ptr("notes_dir = \"" + gone + "\""), dir: gone, status: BadDir, errSub: "does not exist"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			xdg := setup(t)
			if tc.content != nil {
				writeConfig(t, xdg, *tc.content)
			}
			cfg, status, err := Load()
			if cfg.NotesDir != tc.dir || cfg.Theme != tc.theme || status != tc.status {
				t.Errorf("Load() = (%+v, %v), want (%q, %q, %v)", cfg, status, tc.dir, tc.theme, tc.status)
			}
			if tc.errSub == "" && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tc.errSub != "" && (err == nil || !strings.Contains(err.Error(), tc.errSub)) {
				t.Errorf("error = %v, want containing %q", err, tc.errSub)
			}
		})
	}
}

func ptr(s string) *string { return &s }

func TestSaveRoundTrip(t *testing.T) {
	xdg := setup(t)
	notes := t.TempDir()
	if err := Save(Config{NotesDir: notes}); err != nil {
		t.Fatal(err)
	}
	cfg, status, err := Load()
	if err != nil || status != OK || cfg != (Config{NotesDir: notes}) {
		t.Errorf("Load() = (%+v, %v, %v), want (%q, OK, nil)", cfg, status, err, notes)
	}
	b, err := os.ReadFile(filepath.Join(xdg, "clipt", "config.toml"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), "theme") {
		t.Errorf("config without a theme has a theme line:\n%s", b)
	}
}

func TestSaveKeepsTheme(t *testing.T) {
	xdg := setup(t)
	notes := t.TempDir()
	writeConfig(t, xdg, "notes_dir = \"/old\"\ntheme = \"gruvbox_light\"\n")
	cfg, _, _ := Load()
	cfg.NotesDir = notes
	if err := Save(cfg); err != nil {
		t.Fatal(err)
	}
	got, status, err := Load()
	if err != nil || status != OK || got != (Config{NotesDir: notes, Theme: "gruvbox_light"}) {
		t.Errorf("Load() = (%+v, %v, %v)", got, status, err)
	}
	b, _ := os.ReadFile(filepath.Join(xdg, "clipt", "config.toml"))
	if !strings.Contains(string(b), `theme = "gruvbox_light"`) {
		t.Errorf("rewritten config lost the theme:\n%s", b)
	}
}

func TestDir(t *testing.T) {
	xdg := setup(t)
	got, err := Dir()
	if err != nil {
		t.Fatal(err)
	}
	p, _ := Path()
	if want := filepath.Join(xdg, "clipt"); got != want || filepath.Dir(p) != got {
		t.Errorf("Dir() = %q, want %q (Path() = %q)", got, want, p)
	}
}

func TestSaveReadOnlyConfigDir(t *testing.T) {
	xdg := setup(t)
	cfgDir := filepath.Join(xdg, "clipt")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(cfgDir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(cfgDir, 0o755) })
	if os.Geteuid() == 0 {
		t.Skip("root ignores directory permissions")
	}
	if err := Save(Config{NotesDir: t.TempDir()}); err == nil {
		t.Error("Save() succeeded on read-only config dir")
	}
}

func TestLoadParseError(t *testing.T) {
	xdg := setup(t)
	const content = "notes_dir = "
	p := writeConfig(t, xdg, content)
	_, _, err := Load()
	var perr *ParseError
	if !errors.As(err, &perr) {
		t.Fatalf("error = %v, want *ParseError", err)
	}
	if !strings.Contains(err.Error(), p) {
		t.Errorf("error %q does not name %q", err, p)
	}
	got, _ := os.ReadFile(p)
	if string(got) != content {
		t.Errorf("config modified: %q", got)
	}
}
