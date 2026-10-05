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

func write(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestRead(t *testing.T) {
	dir := t.TempDir()
	s := Store{Dir: dir}
	write(t, dir, "a.md", "# hi\n")
	write(t, dir, "empty.md", "")
	if got, err := s.Read("a"); err != nil || got != "# hi\n" {
		t.Errorf("Read(a) = %q, %v", got, err)
	}
	if got, err := s.Read("empty"); err != nil || got != "" {
		t.Errorf("Read(empty) = %q, %v", got, err)
	}
	if _, err := s.Read("missing"); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("Read(missing) err = %v", err)
	}
}

func TestNormalize(t *testing.T) {
	cases := map[string]string{
		"":          "\n",
		"a":         "a\n",
		"a\n":       "a\n",
		"a\n\n\n":   "a\n",
		"\n\n":      "\n",
		"a\n\nb\n":  "a\n\nb\n",
		"a\n \n":    "a\n \n",
		"a\r\n\r\n": "a\r\n\r\n",
	}
	for in, want := range cases {
		if got := Normalize(in); got != want {
			t.Errorf("Normalize(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestWrite(t *testing.T) {
	dir := t.TempDir()
	s := Store{Dir: dir}
	p := filepath.Join(dir, "todo.md")

	if err := s.Write("todo", "one\n"); err != nil {
		t.Fatalf("create: %v", err)
	}
	if b, _ := os.ReadFile(p); string(b) != "one\n" {
		t.Errorf("created content = %q", b)
	}

	if err := s.Write("todo", "two\n"); err != nil {
		t.Fatalf("overwrite: %v", err)
	}
	if b, _ := os.ReadFile(p); string(b) != "two\n" {
		t.Errorf("overwritten content = %q", b)
	}

	if err := os.Remove(p); err != nil {
		t.Fatal(err)
	}
	if err := s.Write("todo", "three\n"); err != nil {
		t.Fatalf("recreate: %v", err)
	}
	if b, _ := os.ReadFile(p); string(b) != "three\n" {
		t.Errorf("recreated content = %q", b)
	}
}

func TestWriteReadOnlyDir(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root ignores directory permissions")
	}
	dir := t.TempDir()
	if err := os.Chmod(dir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(dir, 0o755) })
	if err := (Store{Dir: dir}).Write("x", "x\n"); err == nil {
		t.Error("Write succeeded on read-only dir")
	}
	if _, err := os.Stat(filepath.Join(dir, "x.md")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("x.md stat = %v, want not exist", err)
	}
}

func TestRename(t *testing.T) {
	dir := t.TempDir()
	s := Store{Dir: dir}
	write(t, dir, "a.md", "A")
	write(t, dir, "b.md", "B")

	if err := s.Rename("a", "c"); err != nil {
		t.Fatalf("Rename: %v", err)
	}
	if s.Exists("a") || !s.Exists("c") {
		t.Error("rename did not move the file")
	}
	if got, _ := s.Read("c"); got != "A" {
		t.Errorf("content = %q", got)
	}

	if err := s.Rename("c", "c"); err != nil || !s.Exists("c") {
		t.Errorf("same name: err=%v exists=%v", err, s.Exists("c"))
	}

	if err := s.Rename("c", "b"); err == nil {
		t.Error("rename onto existing note accepted")
	}
	if got, _ := s.Read("b"); got != "B" {
		t.Errorf("b overwritten: %q", got)
	}
	if !s.Exists("c") {
		t.Error("c lost after refused rename")
	}

	if err := s.Rename("c", "x/y"); err == nil {
		t.Error("invalid name accepted")
	}
	if err := s.Rename("missing", "z"); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("missing source err = %v", err)
	}
}

func TestDelete(t *testing.T) {
	dir := t.TempDir()
	s := Store{Dir: dir}
	write(t, dir, "a.md", "A")
	write(t, dir, "b.md", "B")
	if err := s.Delete("a"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if s.Exists("a") || !s.Exists("b") {
		t.Error("delete touched the wrong files")
	}
	if err := s.Delete("a"); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("Delete(missing) err = %v", err)
	}
}
