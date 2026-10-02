package notes

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestValidateName(t *testing.T) {
	bad := []string{"", ".", "..", "a/b", `a\b`, "/", `\`}
	for _, n := range bad {
		if ValidateName(n) == nil {
			t.Errorf("ValidateName(%q) accepted", n)
		}
	}
	for _, n := range []string{"foo.md", "meeting notes", "..a", ".hidden"} {
		if err := ValidateName(n); err != nil {
			t.Errorf("ValidateName(%q) = %v", n, err)
		}
	}
}

func touch(t *testing.T, dir, name string, mtime time.Time) {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(p, mtime, mtime); err != nil {
		t.Fatal(err)
	}
}

func TestList(t *testing.T) {
	dir := t.TempDir()
	base := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	touch(t, dir, "a.md", base)
	touch(t, dir, "b.md", base.Add(2*time.Hour))
	touch(t, dir, "c.md", base.Add(time.Hour))
	touch(t, dir, "d.md", base) // ties with a
	touch(t, dir, ".hidden.md", base)
	touch(t, dir, "image.png", base)
	touch(t, dir, "notes.txt", base)
	if err := os.Mkdir(filepath.Join(dir, "old.md"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}

	got, err := Store{Dir: dir}.List()
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, n := range got {
		names = append(names, n.Name)
	}
	if want := []string{"b", "c", "a", "d"}; !reflect.DeepEqual(names, want) {
		t.Errorf("List() = %v, want %v", names, want)
	}
}

func TestCreate(t *testing.T) {
	dir := t.TempDir()
	s := Store{Dir: dir}

	if err := s.Create("todo"); err != nil {
		t.Fatal(err)
	}
	if !s.Exists("todo") {
		t.Error("todo should exist")
	}
	if err := s.Create("foo.md"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "foo.md.md")); err != nil {
		t.Errorf("foo.md.md not created: %v", err)
	}

	p := filepath.Join(dir, "todo.md")
	if err := os.WriteFile(p, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := s.Create("todo"); !errors.Is(err, os.ErrExist) {
		t.Errorf("Create existing = %v, want ErrExist", err)
	}
	if b, _ := os.ReadFile(p); string(b) != "keep" {
		t.Errorf("existing note overwritten: %q", b)
	}
	if err := s.Create("a/b"); err == nil {
		t.Error("Create accepted invalid name")
	}
}

func TestCreateReadOnlyDir(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root ignores directory permissions")
	}
	dir := t.TempDir()
	if err := os.Chmod(dir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(dir, 0o755) })
	if err := (Store{Dir: dir}).Create("x"); err == nil {
		t.Error("Create succeeded on read-only dir")
	}
}
