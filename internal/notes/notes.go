// Package notes manages markdown notes in a single flat directory.
package notes

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/llysyganicz/clipt/internal/frontmatter"
)

const ext = ".md"

// Note is a markdown file in the notes directory. Name is the file name
// without the trailing .md; Tags come from its front matter.
type Note struct {
	Name    string
	ModTime time.Time
	Tags    []string
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

// List returns the notes sorted by modification time (newest first), then
// name, each with the tags of its front matter. A note whose front matter
// cannot be read or is invalid is listed without tags.
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
		out = append(out, Note{
			Name:    strings.TrimSuffix(fn, ext),
			ModTime: info.ModTime(),
			Tags:    readTags(filepath.Join(s.Dir, fn)),
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if !out[i].ModTime.Equal(out[j].ModTime) {
			return out[i].ModTime.After(out[j].ModTime)
		}
		return out[i].Name < out[j].Name
	})
	return out, nil
}

// readTags reads only the head of the file at path, up to the closing
// front matter fence, and returns its tags; nil on any error.
func readTags(path string) []string {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 4096), 1<<20)
	var head []string
	for sc.Scan() {
		line := strings.TrimSuffix(sc.Text(), "\r")
		head = append(head, line)
		if len(head) == 1 && line != frontmatter.Fence {
			return nil
		}
		if len(head) > 1 && line == frontmatter.Fence {
			tags, err := frontmatter.Parse(strings.Join(head, "\n"))
			if err != nil {
				return nil
			}
			return tags
		}
	}
	return nil
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

// Read returns the content of the note.
func (s Store) Read(name string) (string, error) {
	b, err := os.ReadFile(filepath.Join(s.Dir, name+ext))
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Normalize returns content in the on-disk form clipt writes: trailing
// newlines trimmed and exactly one appended.
func Normalize(content string) string {
	return strings.TrimRight(content, "\n") + "\n"
}

// Write writes content to <name>.md in place, creating the file when it is
// missing and overwriting it otherwise. It does not normalize the content.
func (s Store) Write(name, content string) error {
	return os.WriteFile(filepath.Join(s.Dir, name+ext), []byte(content), 0o644)
}

// Rename renames a note. It validates the new name, does nothing when the
// name is unchanged and never overwrites an existing note.
func (s Store) Rename(oldName, newName string) error {
	if err := ValidateName(newName); err != nil {
		return err
	}
	if oldName == newName {
		return nil
	}
	if s.Exists(newName) {
		return fmt.Errorf("note %q already exists", newName)
	}
	return os.Rename(filepath.Join(s.Dir, oldName+ext), filepath.Join(s.Dir, newName+ext))
}

// Delete removes the note file.
func (s Store) Delete(name string) error {
	return os.Remove(filepath.Join(s.Dir, name+ext))
}
