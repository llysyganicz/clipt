// Package notes manages markdown notes in a single flat directory.
package notes

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const ext = ".md"

// Note is a markdown file in the notes directory. Name is the file name
// without the trailing .md.
type Note struct {
	Name    string
	ModTime time.Time
}

// FileName returns the note's file name on disk.
func (n Note) FileName() string { return n.Name + ext }

// Store gives access to the notes in Dir.
type Store struct {
	Dir string
}

// ValidateName rejects names that cannot be used as a note name.
func ValidateName(name string) error {
	switch {
	case name == "":
		return errors.New("name is empty")
	case name == "." || name == "..":
		return fmt.Errorf("invalid name %q", name)
	case strings.ContainsAny(name, `/\`):
		return fmt.Errorf("invalid name %q: must not contain / or \\", name)
	}
	return nil
}

// List returns the notes sorted by modification time (newest first), then name.
func (s Store) List() ([]Note, error) {
	entries, err := os.ReadDir(s.Dir)
	if err != nil {
		return nil, err
	}
	var out []Note
	for _, e := range entries {
		fn := e.Name()
		if strings.HasPrefix(fn, ".") || !strings.HasSuffix(fn, ext) {
			continue
		}
		info, err := e.Info()
		if err != nil || !info.Mode().IsRegular() {
			continue
		}
		out = append(out, Note{Name: strings.TrimSuffix(fn, ext), ModTime: info.ModTime()})
	}
	sort.Slice(out, func(i, j int) bool {
		if !out[i].ModTime.Equal(out[j].ModTime) {
			return out[i].ModTime.After(out[j].ModTime)
		}
		return out[i].Name < out[j].Name
	})
	return out, nil
}

// Exists reports whether a note with the given name exists as a file.
func (s Store) Exists(name string) bool {
	info, err := os.Stat(filepath.Join(s.Dir, name+ext))
	return err == nil && info.Mode().IsRegular()
}

// Create writes an empty <name>.md. It never overwrites an existing file.
func (s Store) Create(name string) error {
	if err := ValidateName(name); err != nil {
		return err
	}
	f, err := os.OpenFile(filepath.Join(s.Dir, name+ext), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	return f.Close()
}
