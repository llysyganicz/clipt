# config Specification

## Purpose

Persists clipt's user configuration, the notes directory and the optional
color theme name, in a per-user config file on every supported platform.

## Requirements

### Requirement: Config file location
The config file SHALL be `clipt/config.toml` under the user's platform config
directory: `$XDG_CONFIG_HOME` (default `~/.config`) on Linux, `%AppData%` on
Windows and `~/Library/Application Support` on macOS.

#### Scenario: Linux with XDG_CONFIG_HOME set
- **WHEN** clipt starts on Linux with `XDG_CONFIG_HOME=/tmp/xdg`
- **THEN** it reads and writes `/tmp/xdg/clipt/config.toml`

#### Scenario: Linux without XDG_CONFIG_HOME
- **WHEN** clipt starts on Linux with `XDG_CONFIG_HOME` unset
- **THEN** it reads and writes `~/.config/clipt/config.toml`

### Requirement: Config file shape
The config file SHALL be TOML with a string key `notes_dir` holding an
absolute path to the notes directory and an optional string key `theme`
naming the color theme (see the `theme` capability). Unknown keys SHALL be
ignored.

#### Scenario: Reading a valid config
- **WHEN** the config file contains `notes_dir = "/home/user/notes"`
- **THEN** the configured notes directory is `/home/user/notes` and the
  configured theme name is empty

#### Scenario: Reading a theme
- **WHEN** the config file contains `notes_dir = "/home/user/notes"` and
  `theme = "gruvbox_light"`
- **THEN** the configured theme name is `gruvbox_light`

#### Scenario: Unknown keys are ignored
- **WHEN** the config file contains `notes_dir = "/home/user/notes"` and
  `editor = "vim"`
- **THEN** the configured notes directory is `/home/user/notes` and no error
  is reported

### Requirement: Not-configured detection
clipt SHALL treat the notes directory as not configured when the config file
does not exist, when `notes_dir` is missing or empty, or when `notes_dir`
points to a path that is not an existing directory. In each case clipt SHALL
open the notes directory picker instead of the finder.

#### Scenario: First run
- **WHEN** clipt starts and no config file exists
- **THEN** the notes directory picker opens with no prior path and no error

#### Scenario: Empty notes_dir
- **WHEN** the config file exists and `notes_dir` is `""` or absent
- **THEN** the notes directory picker opens with no prior path and no error

#### Scenario: Configured directory no longer exists
- **WHEN** the config file contains `notes_dir = "/mnt/gone/notes"` and that
  path does not exist
- **THEN** the notes directory picker opens prefilled with `/mnt/gone/notes`
  and shows an error saying the directory does not exist

#### Scenario: Configured path is a file
- **WHEN** `notes_dir` points to an existing regular file
- **THEN** the notes directory picker opens prefilled with that path and shows
  an error saying it is not a directory

### Requirement: Writing the config
When a notes directory is accepted, clipt SHALL write the config file with the
absolute, `~`-expanded path and the theme name that was read at startup,
omitting `theme` when none was configured, creating the config directory if
it does not exist. The config file is the only file clipt writes outside the
notes directory.

#### Scenario: Config directory is created
- **WHEN** the user accepts a notes directory and the config directory does
  not exist
- **THEN** the config directory and `config.toml` are created and the file
  contains `notes_dir` with the accepted absolute path

#### Scenario: Theme survives a notes directory change
- **WHEN** the config file contains `theme = "gruvbox_light"` and the user
  changes the notes directory in-app
- **THEN** the rewritten config file still contains `theme = "gruvbox_light"`
  and the new `notes_dir`

#### Scenario: No theme key is not invented
- **WHEN** the config file has no `theme` key and the user changes the notes
  directory in-app
- **THEN** the rewritten config file has no `theme` key

#### Scenario: Config write fails
- **WHEN** the config file cannot be written (for example the config
  directory is read-only)
- **THEN** the picker stays open and shows the write error, and the finder is
  not entered

### Requirement: Malformed config
A config file that cannot be parsed as TOML SHALL be reported as a startup
error with the file path, and clipt SHALL exit without modifying the file.

#### Scenario: Invalid TOML
- **WHEN** the config file contains `notes_dir = ` (invalid TOML)
- **THEN** clipt prints an error naming the config file path and exits with a
  non-zero status, leaving the file unchanged
