## ADDED Requirements

### Requirement: Template insertion
While editing a note, `ctrl+t` SHALL open the template picker. Text handed
back by the insert mode SHALL be inserted at the cursor position, exactly as
if typed, with the cursor ending after the inserted text and the view
scrolled to it. Inserted text counts as a change for auto-save. Returning
from the insert mode without a result SHALL leave the content and the cursor
as they were. While editing a template, `ctrl+t` SHALL do nothing.

#### Scenario: Insert at the cursor
- **WHEN** the note reads `ab`, the cursor is between `a` and `b`, and the
  insert mode hands back `# Standup\n- `
- **THEN** the content is `a# Standup\n- b` and the cursor is right after
  the inserted `- `

#### Scenario: Inserted text is saved
- **WHEN** a template body is inserted and the user presses Esc without
  typing anything else
- **THEN** the note is written with the inserted text

#### Scenario: Cancelled insertion
- **WHEN** the user presses `ctrl+t` and then Esc in the picker
- **THEN** the content and the cursor position are unchanged and the status
  bar shows the editor's hints again

#### Scenario: ctrl+t on a template
- **WHEN** the user is editing a template and presses `ctrl+t`
- **THEN** nothing happens

### Requirement: Editing a template
When edit mode is opened on a template, the status bar SHALL show the
template's name marked as a template on the right, and the hints SHALL NOT
offer template insertion. Saving SHALL follow the same rules as for a note,
writing `.templates/<name>.md`.

#### Scenario: Status bar for a template
- **WHEN** the template `meeting` is open in edit mode
- **THEN** the right side of the status bar identifies `meeting` as a
  template

#### Scenario: Template is saved
- **WHEN** the user types a line in the template `meeting` and presses Esc
- **THEN** `.templates/meeting.md` contains the new line

## MODIFIED Requirements

### Requirement: Editing keys
Every key other than Esc, `ctrl+c` and `ctrl+t` SHALL be handled by the text
area with standard single-line-editor bindings: printable keys insert at the
cursor, Enter inserts a line break, Backspace and Delete remove characters,
and the arrow keys, Home and End move the cursor. The text area's default
`ctrl+t` binding (transpose characters) SHALL be disabled. There SHALL be no
modal editing; the text area is always in insert mode.

#### Scenario: Typing
- **WHEN** the cursor is at the end of `hello` and the user types ` world`
- **THEN** the line reads `hello world`

#### Scenario: New line
- **WHEN** the user presses Enter at the end of the last line
- **THEN** a new empty line is added below it and the cursor moves there

#### Scenario: Moving up
- **WHEN** the cursor is on line 3 and the user presses Up
- **THEN** the cursor is on line 2 and the content is unchanged

#### Scenario: ctrl+t does not transpose
- **WHEN** the cursor is after `ab` and the user presses `ctrl+t`
- **THEN** the line still reads `ab`

### Requirement: Leaving edit mode
Esc SHALL return to preview of the current note, or to the finder in
template scope when a template is being edited. Before returning, the file
SHALL be written if and only if its content changed since it was loaded.
`ctrl+c` SHALL quit clipt immediately without writing.

#### Scenario: Changed content is saved
- **WHEN** `todo` is open in edit mode, the user types a line and presses Esc
- **THEN** `todo.md` contains the new line and preview opens showing it

#### Scenario: Unchanged content is not written
- **WHEN** `todo.md` was last modified yesterday, the user opens it in edit
  mode and presses Esc without changing anything
- **THEN** preview opens and `todo.md` keeps yesterday's modification time

#### Scenario: Edit then undo by hand
- **WHEN** the user types `x`, deletes it again and presses Esc
- **THEN** the file is not written

#### Scenario: ctrl+c discards
- **WHEN** the user types a line and presses `ctrl+c`
- **THEN** clipt exits and the file is unchanged

#### Scenario: Esc on a template
- **WHEN** the template `meeting` is open in edit mode and the user presses
  Esc
- **THEN** the finder opens in template scope and the current note is
  unchanged
