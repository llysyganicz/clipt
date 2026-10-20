# editor Specification

## Purpose

The writing screen: the current note's (or a template's) raw markdown in a
text area with keyboard selection and template insertion at the cursor, but
no modal editing. Leaving it saves the file when it changed and returns to
preview, or to the finder for a template.

## Requirements

### Requirement: Content area
Edit mode SHALL show the current note's raw content in a text area with a
line number gutter that fills the terminal above the status bar. The
content SHALL be re-read from disk every time edit mode is entered. The
cursor SHALL start at the end of the document and the view SHALL show the
cursor. Lines longer than the width SHALL be soft-wrapped. There SHALL be no
limit on the number of characters or lines in a note. Resizing the terminal
SHALL resize the text area to the new size.

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

### Requirement: Line number gutter
Every row of the text area SHALL start with a gutter holding the line
number, right-aligned, followed by one space. The gutter SHALL be the same
width on every row, and that width SHALL be the number of digits in the
note's line count plus one space on each side. Rows that continue a
soft-wrapped line SHALL show a blank gutter of the same width. The text of
every row therefore SHALL start in the same column. When the line count
gains or loses a digit, the gutter SHALL change width on every row at once.
The number of the cursor's line SHALL be highlighted.

#### Scenario: Text aligns across line 10
- **WHEN** a note has 12 lines each starting with two spaces
- **THEN** the text of line 9 and the text of line 10 start in the same
  column and the gutter is two digits wide on every row

#### Scenario: Text aligns across line 100
- **WHEN** a note has 120 lines
- **THEN** the text of line 99 and the text of line 100 start in the same
  column and the gutter is three digits wide on every row

#### Scenario: Wrapped line
- **WHEN** a line is longer than the width and wraps onto a second row
- **THEN** the second row has a blank gutter and its text starts in the same
  column as the first row's

#### Scenario: Gutter grows with the note
- **WHEN** a note has 9 lines and the user presses Enter at the end of the
  last line
- **THEN** the note has 10 lines and the text of every row has moved one
  column to the right

#### Scenario: Cursor line number
- **WHEN** the cursor is on line 3
- **THEN** the number 3 is highlighted and the other numbers are not

### Requirement: Status bar
The last line SHALL be a status bar showing the available keys on the left,
including that Esc saves and returns and that `ctrl+q` saves and quits, and
the current note's name on the right. While a selection exists, the hints
SHALL also show that `ctrl+shift+c` copies it.

#### Scenario: Status bar content
- **WHEN** `todo` is open in edit mode
- **THEN** the last line shows the Esc and `ctrl+q` hints on the left and
  `todo` on the right

#### Scenario: Copy hint while selecting
- **WHEN** the user presses Shift+Left on `ab` with the cursor at its end
- **THEN** the hints include `ctrl+shift+c copy`

#### Scenario: Copy hint gone after the selection is cleared
- **WHEN** a selection exists and the user presses Right
- **THEN** the hints no longer include `ctrl+shift+c copy`

### Requirement: Editing keys
Every key other than Esc, `ctrl+c`, `ctrl+q`, `ctrl+t`, Tab and Shift+Tab
SHALL be handled by the text area with its default bindings: printable keys
insert at the cursor, Enter inserts a line break, Backspace and Delete
remove characters, the arrow keys, Home and End move the cursor, and the
Shift-modified movement keys select (see the Selection requirement). The
text area's default `ctrl+t` binding (transpose characters) SHALL be
disabled; no other default binding SHALL be changed. There SHALL be no
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

### Requirement: Selection
Shift+Left and Shift+Right SHALL extend a selection by one character,
Shift+Up and Shift+Down by one line, and Ctrl+Shift+Left and
Ctrl+Shift+Right by one word, in each case moving the cursor as the
unshifted key would. `ctrl+g` SHALL select the whole note. A selection
starts at the cursor position where the first selecting key was pressed and
ends at the cursor. Selected text SHALL be drawn with a distinct background.
Any unshifted cursor movement SHALL clear the selection without changing
the content. Typing a printable key, Enter, Tab, pasting or inserting a
template SHALL replace the selection with the inserted text; Backspace and
Delete SHALL remove the selection. Making or clearing a selection is not a change for auto-save.
Leaving edit mode SHALL discard the selection.

#### Scenario: Select backwards by character
- **WHEN** the line reads `hello` with the cursor at its end and the user
  presses Shift+Left twice
- **THEN** `lo` is selected and the cursor is before `l`

#### Scenario: Select by line
- **WHEN** the cursor is at the start of line 2 of a three-line note and the
  user presses Shift+Down
- **THEN** line 2 is selected and the cursor is at the start of line 3

#### Scenario: Select by word
- **WHEN** the cursor is at the start of `one two` and the user presses
  Ctrl+Shift+Right
- **THEN** `one` is selected

#### Scenario: Select all
- **WHEN** the note has three lines and the user presses `ctrl+g`
- **THEN** all three lines are selected

#### Scenario: Movement clears the selection
- **WHEN** `lo` is selected in `hello` and the user presses Left
- **THEN** nothing is selected and the line still reads `hello`

#### Scenario: Typing replaces the selection
- **WHEN** `lo` is selected in `hello` and the user types `p`
- **THEN** the line reads `help` and the cursor is after `p`

#### Scenario: Backspace removes the selection
- **WHEN** `lo` is selected in `hello` and the user presses Backspace
- **THEN** the line reads `hel`

#### Scenario: Selection across lines
- **WHEN** the note reads `ab\ncd`, the cursor is after `a`, and the user
  presses Shift+Down and then Delete
- **THEN** the note reads `ad`

#### Scenario: Template insertion replaces the selection
- **WHEN** `lo` is selected in `hello` and the insert mode hands back
  `# Standup\n- `
- **THEN** the content is `hel# Standup\n- ` and the cursor is right after
  the inserted `- `

#### Scenario: Selection alone is not a change
- **WHEN** the user opens an unchanged note, presses `ctrl+g` and then Esc
- **THEN** the file is not written

#### Scenario: Selection in a template
- **WHEN** the template `meeting` is open in edit mode and the user presses
  Shift+Left
- **THEN** the character before the cursor is selected

### Requirement: Copy selection
`ctrl+shift+c` SHALL copy the selected text to the system clipboard and
leave the selection and the content unchanged. With no selection,
`ctrl+shift+c` SHALL do nothing. When the clipboard cannot be written, the
error SHALL be shown above the status bar the way write errors are shown,
and disappear on the next key press.

#### Scenario: Copy
- **WHEN** `lo` is selected in `hello` and the user presses `ctrl+shift+c`
- **THEN** the clipboard holds `lo`, `lo` is still selected and the line
  still reads `hello`

#### Scenario: Copy without a selection
- **WHEN** nothing is selected and the user presses `ctrl+shift+c`
- **THEN** the content, the cursor and the clipboard are unchanged

#### Scenario: Clipboard unavailable
- **WHEN** no clipboard utility is available and the user presses
  `ctrl+shift+c` on a selection
- **THEN** an error is shown above the status bar, the selection is kept and
  the note is unchanged

### Requirement: Indentation keys
Tab SHALL insert two spaces at the cursor and leave the cursor after them.
Shift+Tab SHALL remove up to two spaces from the start of the cursor's line,
moving the cursor left by the number of spaces removed, but not before the
start of the line. Both count as a change for auto-save.

#### Scenario: Tab at line start
- **WHEN** the cursor is at the start of `item` and the user presses Tab
- **THEN** the line reads `  item` and the cursor is before `item`

#### Scenario: Tab inside a line
- **WHEN** the cursor is between `a` and `b` on the line `ab` and the user
  presses Tab
- **THEN** the line reads `a  b` and the cursor is before `b`

#### Scenario: Shift+Tab removes two spaces
- **WHEN** the line reads `    item`, the cursor is at its end, and the user
  presses Shift+Tab
- **THEN** the line reads `  item` and the cursor is still at its end

#### Scenario: Shift+Tab removes a single space
- **WHEN** the line reads ` item` and the user presses Shift+Tab
- **THEN** the line reads `item`

#### Scenario: Shift+Tab with no indentation
- **WHEN** the line reads `item` and the user presses Shift+Tab
- **THEN** the line is unchanged

#### Scenario: Indentation is saved
- **WHEN** the user presses Tab on an unchanged note and then Esc
- **THEN** the file is written with the two spaces

### Requirement: Leaving edit mode
Esc SHALL return to preview of the current note, or to the finder in
template scope when a template is being edited. Before returning, the file
SHALL be written if and only if its content changed since it was loaded.
`ctrl+q` SHALL write the file under the same rule and then quit clipt; when
the write fails, edit mode SHALL stay open showing the error, as it does on
Esc, and clipt SHALL keep running. `ctrl+c` SHALL quit clipt immediately
without writing.

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

#### Scenario: ctrl+q saves and quits
- **WHEN** the user types a line and presses `ctrl+q`
- **THEN** `todo.md` contains the new line and clipt exits

#### Scenario: ctrl+q on unchanged content
- **WHEN** the user opens `todo` in edit mode and presses `ctrl+q` without
  changing anything
- **THEN** clipt exits and `todo.md` is not written

#### Scenario: ctrl+q with a failing write
- **WHEN** the notes directory became read-only, the user types a line and
  presses `ctrl+q`
- **THEN** edit mode stays open with the typed content and an error, and
  clipt keeps running

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
