## MODIFIED Requirements

### Requirement: Cancel
Esc in the picker or in the form SHALL return to edit mode with the note
unchanged and the cursor where it was. `ctrl+q` in the picker or in the form
SHALL abandon the insertion, write the note if its content changed since it
was loaded, and quit clipt; when the write fails, edit mode SHALL reopen
with the content preserved and the error shown, and clipt SHALL keep
running. `ctrl+c` SHALL quit clipt without writing.

#### Scenario: Esc from the form
- **WHEN** the form is open with values typed and the user presses Esc
- **THEN** edit mode is back, nothing was inserted and the cursor has not
  moved

#### Scenario: Esc from the picker
- **WHEN** the picker is open and the user presses Esc
- **THEN** edit mode is back and nothing was inserted

#### Scenario: ctrl+q from the form
- **WHEN** the user typed a line in `todo`, opened the form and presses
  `ctrl+q`
- **THEN** `todo.md` contains the typed line, nothing from the template, and
  clipt exits

#### Scenario: ctrl+q from the picker with a failing write
- **WHEN** the notes directory became read-only, the user typed a line,
  opened the picker and presses `ctrl+q`
- **THEN** edit mode is back with the typed line and an error, and clipt
  keeps running

### Requirement: Layout
The picker and the form SHALL fill the terminal and end with a status bar
listing the available keys on the left, including `ctrl+q`. Resizing the
terminal SHALL resize the list.

#### Scenario: Status bar
- **WHEN** the form is open
- **THEN** the last line shows hints for moving between rows, submitting,
  Esc and `ctrl+q`
