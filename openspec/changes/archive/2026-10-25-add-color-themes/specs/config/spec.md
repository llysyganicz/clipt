## MODIFIED Requirements

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
