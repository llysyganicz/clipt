## MODIFIED Requirements

### Requirement: Leaving the picker
Esc in the path input SHALL quit clipt when the picker was opened at
startup, and SHALL return to the finder with the directory unchanged when
the picker was opened via `ctrl+o`. `ctrl+q` SHALL quit clipt from any
picker state without writing the config, and the picker's key hints SHALL
list it. `ctrl+c` SHALL quit from any picker state.

#### Scenario: Esc on first run
- **WHEN** the picker was opened at startup and the user presses Esc in the
  path input
- **THEN** clipt exits without writing the config

#### Scenario: Esc from ctrl+o
- **WHEN** the picker was opened via `ctrl+o` and the user presses Esc in the
  path input
- **THEN** the finder reopens with the same notes directory, the same current
  note and the same query as before `ctrl+o`

#### Scenario: ctrl+q from ctrl+o
- **WHEN** the picker was opened via `ctrl+o` and the user presses `ctrl+q`
  in the path input
- **THEN** clipt exits and the config is unchanged

#### Scenario: ctrl+q during the create confirmation
- **WHEN** the picker asks whether to create a missing directory and the
  user presses `ctrl+q`
- **THEN** clipt exits and no directory is created
