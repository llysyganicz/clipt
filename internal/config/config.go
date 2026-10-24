// Package config reads and writes clipt's config file.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Status describes the result of loading the config.
type Status int

const (
	// OK means notes_dir is set and is an existing directory.
	OK Status = iota
	// MissingFile means the config file does not exist.
	MissingFile
	// EmptyDir means notes_dir is missing or empty.
	EmptyDir
	// BadDir means notes_dir does not exist or is not a directory.
	BadDir
)

// ParseError reports a config file that is not valid TOML.
type ParseError struct {
	Path string
	Err  error
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("invalid config file %s: %v", e.Path, e.Err)
}

func (e *ParseError) Unwrap() error { return e.Err }

// Config is the content of the config file.
type Config struct {
	// NotesDir is the configured notes directory; it may be stale.
	NotesDir string
	// Theme names the color theme; empty when the file sets none.
	Theme string
}

type file struct {
	NotesDir string `toml:"notes_dir"`
	Theme    string `toml:"theme,omitempty"`
}

// Dir returns the clipt config directory, the parent of the config file.
func Dir() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "clipt"), nil
}

// Path returns the location of the config file.
func Path() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.toml"), nil
}

// Load reads the config. It returns the configured values (the notes dir
// possibly stale), a status, and an error only for a parse failure or an
// unexpected I/O failure. For BadDir the error describes why the dir is
// unusable.
func Load() (Config, Status, error) {
	p, err := Path()
	if err != nil {
		return Config{}, MissingFile, err
	}
	var f file
	if _, err := toml.DecodeFile(p, &f); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Config{}, MissingFile, nil
		}
		var perr toml.ParseError
		if errors.As(err, &perr) {
			return Config{}, EmptyDir, &ParseError{Path: p, Err: err}
		}
		return Config{}, EmptyDir, err
	}
	cfg := Config{NotesDir: f.NotesDir, Theme: f.Theme}
	if f.NotesDir == "" {
		return cfg, EmptyDir, nil
	}
	info, err := os.Stat(f.NotesDir)
	switch {
	case err != nil && errors.Is(err, os.ErrNotExist):
		return cfg, BadDir, fmt.Errorf("directory does not exist: %s", f.NotesDir)
	case err != nil:
		return cfg, BadDir, err
	case !info.IsDir():
		return cfg, BadDir, fmt.Errorf("not a directory: %s", f.NotesDir)
	}
	return cfg, OK, nil
}

// Save writes cfg to the config file, creating its directory. The theme key
// is omitted when cfg.Theme is empty.
func Save(cfg Config) error {
	p, err := Path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	out, err := os.OpenFile(p, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	if err := toml.NewEncoder(out).Encode(file{NotesDir: cfg.NotesDir, Theme: cfg.Theme}); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}
