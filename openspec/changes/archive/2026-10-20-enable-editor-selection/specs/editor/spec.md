## MODIFIED Requirements

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

## ADDED Requirements

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
