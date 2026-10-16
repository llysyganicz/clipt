## ADDED Requirements

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

## MODIFIED Requirements

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

### Requirement: Status bar
The last line SHALL be a status bar showing the available keys on the left,
including that Esc saves and returns and that `ctrl+q` saves and quits, and
the current note's name on the right.

#### Scenario: Status bar content
- **WHEN** `todo` is open in edit mode
- **THEN** the last line shows the Esc and `ctrl+q` hints on the left and
  `todo` on the right

### Requirement: Editing keys
Every key other than Esc, `ctrl+c`, `ctrl+q`, `ctrl+t`, Tab and Shift+Tab
SHALL be handled by the text area with standard single-line-editor bindings:
printable keys insert at the cursor, Enter inserts a line break, Backspace
and Delete remove characters, and the arrow keys, Home and End move the
cursor. The text area's default `ctrl+t` binding (transpose characters)
SHALL be disabled. There SHALL be no modal editing; the text area is always
in insert mode.

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
