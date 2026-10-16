# preview Specification

## Purpose

The reading screen: shows the current note as rendered markdown and is the
place to edit, rename or delete it, or go back to the finder.

## Requirements

### Requirement: Rendered content
Preview SHALL show the current note's content rendered as markdown,
word-wrapped to the terminal width, in a scrollable area that fills the
terminal above the status bar. The note SHALL be re-read from disk every time
preview is entered, and the content SHALL start scrolled to the top. The
rendered markdown SHALL use the application's theme palette. Resizing the
terminal SHALL re-render the content at the new width.

#### Scenario: Open a note
- **WHEN** `todo` contains `# Todo` and a list, and the user opens it from
  the finder
- **THEN** preview shows the heading and list rendered, starting at the top

#### Scenario: Content changed on disk
- **WHEN** the file behind the current note is changed externally while the
  user is in the finder, and the user presses Esc
- **THEN** preview shows the new content

#### Scenario: Empty note
- **WHEN** the current note is an empty file
- **THEN** the content area is empty and the status bar is shown

#### Scenario: Resize
- **WHEN** the terminal is made narrower while preview is open
- **THEN** the content is re-wrapped to the new width

### Requirement: Scrolling
Up/Down SHALL scroll the content by one line and PgUp/PgDn by one page.
Scrolling SHALL stop at the first and last line without wrapping. The status
bar SHALL stay visible while scrolling.

#### Scenario: Long note
- **WHEN** the rendered note is taller than the content area and the user
  presses Down
- **THEN** the content moves up by one line and the status bar stays in place

#### Scenario: End of content
- **WHEN** the last line is already visible and the user presses PgDn
- **THEN** the view does not change

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

### Requirement: Rename
`r` SHALL open a rename prompt in place of the status bar, prefilled with the
current name with the cursor at its end. Enter SHALL rename the note file to
the trimmed input and make the renamed note the current note, keeping the
content view as it is. Esc, or Enter with the name unchanged, SHALL cancel
without touching the file. The rename SHALL be refused, with an error shown
and no file changed, when the name is invalid (empty, contains `/` or `\`,
or is `.` or `..`) or when a note with that name already exists. Any other
failure SHALL be shown as an error and leave the prompt open.

#### Scenario: Rename succeeds
- **WHEN** `todo` is open, the user presses `r`, changes the name to
  `todo-2026` and presses Enter
- **THEN** `todo.md` is renamed to `todo-2026.md`, the status bar shows
  `todo-2026`, and on returning to the finder `todo-2026` is the hidden
  current note

#### Scenario: Rename cancelled
- **WHEN** the user presses `r`, edits the name and presses Esc
- **THEN** no file changes and the status bar shows the original name

#### Scenario: Same name
- **WHEN** the user presses `r` and presses Enter without changing the name
- **THEN** no file changes and the prompt closes

#### Scenario: Target exists
- **WHEN** notes `todo` and `ideas` exist, `todo` is open, and the user
  renames it to `ideas`
- **THEN** both files are unchanged and an error says `ideas` already exists

#### Scenario: Invalid name
- **WHEN** the user renames the note to `a/b`
- **THEN** no file changes and an error about the invalid name is shown

### Requirement: Delete
`d` SHALL show a `y/n` confirmation in place of the status bar. `y` SHALL
remove the note file, clear the current note and return to the finder with
an empty query. `n` or Esc SHALL cancel and restore the status bar. A
failure to remove the file SHALL be shown as an error and leave the user in
preview.

#### Scenario: Delete confirmed
- **WHEN** `todo` is open, the user presses `d` and then `y`
- **THEN** `todo.md` is removed, the finder opens with an empty query, and
  `todo` is not listed and not hidden as current

#### Scenario: Delete cancelled
- **WHEN** the user presses `d` and then `n`
- **THEN** the file is unchanged and the status bar is shown again

#### Scenario: Delete fails
- **WHEN** the user confirms deletion and the file cannot be removed
- **THEN** preview stays open and shows the OS error

### Requirement: Note cannot be read
When the current note cannot be read on entering preview, clipt SHALL return
to the finder, show an error naming the note, and clear the current note.

#### Scenario: File removed before opening
- **WHEN** `todo` is listed in the finder, its file is removed externally,
  and the user presses Enter on it
- **THEN** the finder stays open with an error that `todo` cannot be read,
  and no note is current
