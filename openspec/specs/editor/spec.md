# editor Specification

## Purpose

The writing screen: the current note's (or a template's) raw markdown in an
insert-only text area, with template insertion at the cursor. Leaving it
saves the file when it changed and returns to preview, or to the finder for
a template.

## Requirements

### Requirement: Content area
Edit mode SHALL show the current note's raw content in a text area with line
numbers that fills the terminal above the status bar. The content SHALL be
re-read from disk every time edit mode is entered. The cursor SHALL start at
the end of the document and the view SHALL show the cursor. Lines longer
than the width SHALL be soft-wrapped. There SHALL be no limit on the number
of characters or lines in a note. Resizing the terminal SHALL resize the text
area to the new size.

#### Scenario: Open an existing note
- **WHEN** `todo` contains three lines and the user presses `e` in preview
- **THEN** edit mode shows the three numbered lines and the cursor is at the
  end of the last line

#### Scenario: Open a new note
- **WHEN** the user creates `ideas` from the finder
- **THEN** edit mode opens with an empty text area, line number 1 and the
  cursor on it

#### Scenario: Cursor is visible on entry
- **WHEN** the note has more lines than the terminal has rows
- **THEN** edit mode opens scrolled so that the last line and the cursor are
  visible

#### Scenario: No length limit
- **WHEN** a note has 5,000 lines and the user appends another line
- **THEN** the line is inserted and the note keeps all 5,001 lines

#### Scenario: Resize
- **WHEN** the terminal is made taller while edit mode is open
- **THEN** the text area grows to fill the space above the status bar

### Requirement: Status bar
The last line SHALL be a status bar showing the available keys on the left,
including that Esc saves and returns, and the current note's name on the
right.

#### Scenario: Status bar content
- **WHEN** `todo` is open in edit mode
- **THEN** the last line shows the Esc hint on the left and `todo` on the
  right

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

### Requirement: Saved form
A written note SHALL end with exactly one newline, regardless of how many
trailing newlines the text area holds. Content SHALL count as unchanged when
it equals the loaded content apart from trailing newlines. On load, Windows
line endings (`\r\n`) SHALL be treated as `\n` and tab characters SHALL be
shown as four spaces; a note is written in that form only when its content
changed.

#### Scenario: Trailing newline added
- **WHEN** the user types `done` on the last line and presses Esc
- **THEN** the file ends with `done` followed by a single newline

#### Scenario: Extra trailing blank lines are dropped
- **WHEN** the text area ends with three empty lines after `done` and the
  content changed elsewhere
- **THEN** the written file ends with `done` followed by a single newline

#### Scenario: Only the trailing newline differs
- **WHEN** `todo.md` contains `a` with no trailing newline and the user
  presses Esc without typing
- **THEN** the file is not written

#### Scenario: Windows line endings
- **WHEN** `todo.md` contains `a\r\nb\r\n`
- **THEN** edit mode shows two lines, `a` and `b`, with no blank line between
  them

#### Scenario: Tabs
- **WHEN** `todo.md` contains a line starting with a tab and the user edits
  another line and presses Esc
- **THEN** the written file has four spaces where the tab was

### Requirement: Write failure
When the note cannot be written on Esc, edit mode SHALL stay open with the
content preserved and show the error above the status bar. The error SHALL
disappear on the next key press, and Esc SHALL try again. When the note's
file was removed while editing, writing SHALL create it again.

#### Scenario: Directory not writable
- **WHEN** the notes directory became read-only, the user types a line and
  presses Esc
- **THEN** edit mode stays open, the typed content is still there and an
  error with the OS message is shown

#### Scenario: Retry succeeds
- **WHEN** a write failed, the directory is made writable again and the user
  presses Esc
- **THEN** the note is written and preview opens

#### Scenario: File removed while editing
- **WHEN** `todo.md` is removed externally while the user is editing it, and
  the user types a line and presses Esc
- **THEN** `todo.md` is created with the edited content and preview opens

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
