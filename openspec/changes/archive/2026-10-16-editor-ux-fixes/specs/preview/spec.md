## MODIFIED Requirements

### Requirement: Status bar
The last line SHALL be a status bar showing the available keys on the left,
including `ctrl+q` to quit, and the current note's name on the right. While
a rename prompt or delete confirmation is active, it SHALL replace the
status bar.

#### Scenario: Normal state
- **WHEN** `todo` is open and no prompt is active
- **THEN** the last line shows the key hints for edit, rename, delete, back
  and quit on the left and `todo` on the right

### Requirement: Leaving preview
Esc SHALL return to the finder with an empty query. `e` SHALL open the
current note in edit mode. `ctrl+q` SHALL quit clipt from the normal state
and while a rename prompt or delete confirmation is active. `ctrl+c` SHALL
quit clipt. Other keys SHALL do nothing.

#### Scenario: Back to the finder
- **WHEN** the user typed `ide` in the finder, pressed Esc to return to
  preview, and presses Esc again
- **THEN** the finder opens with an empty query and the current note hidden

#### Scenario: Enter edit
- **WHEN** `todo` is open and the user presses `e`
- **THEN** edit mode opens for `todo`

#### Scenario: ctrl+q
- **WHEN** `todo` is open and the user presses `ctrl+q`
- **THEN** clipt exits

#### Scenario: ctrl+q during rename
- **WHEN** the rename prompt is open and the user presses `ctrl+q`
- **THEN** clipt exits and the note keeps its name

#### Scenario: Unbound key
- **WHEN** the user presses `x`
- **THEN** nothing changes
