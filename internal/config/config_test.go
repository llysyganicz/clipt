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
		status  Status
		errSub  string
	}{
		{name: "no file", status: MissingFile},
		{name: "empty notes_dir", content: ptr(`notes_dir = ""`), status: EmptyDir},
		{name: "absent notes_dir", content: ptr(`theme = "dark"`), status: EmptyDir},
		{name: "unknown keys ignored", content: ptr("notes_dir = \"" + notes + "\"\ntheme = \"dark\"\n"), dir: notes, status: OK},
		{name: "path is a file", content: ptr("notes_dir = \"" + file + "\""), dir: file, status: BadDir, errSub: "not a directory"},
		{name: "path missing", content: ptr("notes_dir = \"" + gone + "\""), dir: gone, status: BadDir, errSub: "does not exist"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			xdg := setup(t)
			if tc.content != nil {
				writeConfig(t, xdg, *tc.content)
			}
			dir, status, err := Load()
			if dir != tc.dir || status != tc.status {
				t.Errorf("Load() = (%q, %v), want (%q, %v)", dir, status, tc.dir, tc.status)
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
	setup(t)
	notes := t.TempDir()
	if err := Save(notes); err != nil {
		t.Fatal(err)
	}
	dir, status, err := Load()
	if err != nil || status != OK || dir != notes {
		t.Errorf("Load() = (%q, %v, %v), want (%q, OK, nil)", dir, status, err, notes)
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
	if err := Save(t.TempDir()); err == nil {
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
