## Purpose

Lets the user choose the notes directory from inside clipt, both on first run
and later via `ctrl+o`, without leaving the keyboard-driven TUI.

## ADDED Requirements

### Requirement: Picker entry points
The picker SHALL open at startup whenever the notes directory is not
configured (see `config`), and from the finder on `ctrl+o`. The picker SHALL
remember which entry point opened it, because Esc behaves differently for
each.

#### Scenario: Opened at startup
- **WHEN** clipt starts with no configured notes directory
- **THEN** the picker is the first screen shown

#### Scenario: Opened from the finder
- **WHEN** the user presses `ctrl+o` in the finder
- **THEN** the picker opens prefilled with the currently configured directory

### Requirement: Path input
The picker SHALL present a single-line text input for the path. On first run
the input SHALL be prefilled with `~/notes`. A leading `~` SHALL be expanded
to the user's home directory, and relative paths SHALL be resolved against
the current working directory. The picker SHALL NOT show live feedback while
typing; errors appear only after Enter.

#### Scenario: First-run default
- **WHEN** the picker opens at startup with no prior path
- **THEN** the input contains `~/notes` and Enter alone accepts it

#### Scenario: Tilde expansion
- **WHEN** the user enters `~/docs/notes` and the home directory is
  `/home/user`
- **THEN** the accepted path is `/home/user/docs/notes`

#### Scenario: Whitespace is trimmed
- **WHEN** the user enters `  ~/notes  `
- **THEN** the path is treated as `~/notes`

#### Scenario: Empty input
- **WHEN** the user presses Enter with an empty input
- **THEN** the picker shows an error saying a path is required and stays open

### Requirement: Accepting an existing directory
When the entered path is an existing directory, the picker SHALL write it to
the config and enter the finder listing that directory.

#### Scenario: Existing directory
- **WHEN** the user enters a path to an existing directory and presses Enter
- **THEN** the config is written with that path and the finder opens listing
  the notes in it

#### Scenario: Path is a file
- **WHEN** the user enters a path to an existing regular file and presses
  Enter
- **THEN** the picker shows an error saying the path is not a directory and
  stays open with the input unchanged

### Requirement: Creating a missing directory
When the entered path does not exist, the picker SHALL ask for confirmation
before creating it. On `y` it SHALL create the directory and any missing
parents, write the config and enter the finder. On `n` or Esc it SHALL
return to the path input with the path unchanged.

#### Scenario: Confirm creation
- **WHEN** the user enters `/home/user/new/notes`, which does not exist, and
  presses Enter
- **THEN** the picker shows a prompt asking whether to create
  `/home/user/new/notes` (y/n)

#### Scenario: Creation accepted
- **WHEN** the create prompt is shown and the user presses `y`
- **THEN** the directory and its missing parents are created, the config is
  written and the finder opens with an empty list

#### Scenario: Creation declined
- **WHEN** the create prompt is shown and the user presses `n` or Esc
- **THEN** the picker returns to the path input with the same path and no
  directory is created

#### Scenario: Creation fails
- **WHEN** the user confirms creation and the directory cannot be created
  (for example the parent is read-only)
- **THEN** the picker shows the creation error, returns to the path input and
  does not write the config

### Requirement: Leaving the picker
Esc in the path input SHALL quit clipt when the picker was opened at
startup, and SHALL return to the finder with the directory unchanged when
the picker was opened via `ctrl+o`. `ctrl+c` SHALL quit from any picker
state.

#### Scenario: Esc on first run
- **WHEN** the picker was opened at startup and the user presses Esc in the
  path input
- **THEN** clipt exits without writing the config

#### Scenario: Esc from ctrl+o
- **WHEN** the picker was opened via `ctrl+o` and the user presses Esc in the
  path input
- **THEN** the finder reopens with the same notes directory, the same current
  note and the same query as before `ctrl+o`

### Requirement: Changing the directory clears the current note
When the picker was opened via `ctrl+o` and a directory is accepted, the
finder SHALL reopen with an empty query listing the new directory. If the
accepted directory differs from the previous one, the current note SHALL be
cleared. Accepting the same directory SHALL keep the current note.

#### Scenario: Different directory accepted
- **WHEN** the user opens the picker via `ctrl+o` while `todo` is the current
  note and accepts a different directory
- **THEN** the finder lists the new directory with no note hidden and Esc in
  the finder quits clipt

#### Scenario: Same directory accepted
- **WHEN** the user opens the picker via `ctrl+o` while `todo` is the current
  note and accepts the directory already configured
- **THEN** the finder reopens with `todo` still current and hidden
