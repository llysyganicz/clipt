# finder Specification

## Purpose

The home screen and note switcher: fuzzy-find a note by name to open it, or
create a new one by typing a name that does not exist yet.

## Requirements

### Requirement: Note listing
The finder SHALL list the notes in the configured directory. A note is a
regular file whose name ends in `.md` and does not start with `.`.
Directories, dotfiles and files with other extensions SHALL be ignored. The
displayed name of a note is its file name without the trailing `.md`. The
listing SHALL be re-read from disk every time the finder is entered.

#### Scenario: Only markdown notes are listed
- **WHEN** the directory contains `todo.md`, `ideas.md`, `.hidden.md`,
  `image.png` and a subdirectory `old/`
- **THEN** the finder lists `todo` and `ideas` only

#### Scenario: Listing refreshes on entry
- **WHEN** a note is created, renamed or deleted and the user returns to the
  finder
- **THEN** the list reflects the change without restarting clipt

### Requirement: Default order
With an empty query the finder SHALL list notes by modification time, newest
first. Ties SHALL be broken by name.

#### Scenario: Newest first
- **WHEN** `a.md` was modified at 10:00, `b.md` at 12:00 and `c.md` at 11:00
  and the query is empty
- **THEN** the list order is `b`, `c`, `a`

### Requirement: Current note is hidden
The finder SHALL hide the current note (the note last opened in preview or
created, if any) from the list, so the top row is the next most recent note.

#### Scenario: Returning from a note
- **WHEN** `b` is the current note and the user is in the finder with an
  empty query
- **THEN** `b` does not appear in the list and the top row is the most
  recently modified other note

#### Scenario: Switching between two notes
- **WHEN** `b` is current, `c` is the next most recently modified note, and
  the user presses Enter on the top row, then Esc from preview, then Enter
  on the top row again
- **THEN** the user is back in preview of `b`

### Requirement: Fuzzy matching
With a non-empty query the finder SHALL show the notes whose displayed name
contains every character of the query in order (a subsequence match),
compared case-insensitively. Matching notes SHALL keep the default order
(modification time, newest first); there is no relevance scoring. Apart
from case folding, the query is matched literally against the displayed
name; no extension handling or other normalization is applied. Changing the
query SHALL reset the selection to the top row.

#### Scenario: Subsequence match
- **WHEN** notes `todos`, `todo-archive` and `ideas` exist and the query is
  `tdo`
- **THEN** the list shows `todos` and `todo-archive` and not `ideas`

#### Scenario: Matches keep modification order
- **WHEN** `todos` was modified after `todo-archive` and the query is `tdo`
- **THEN** the list order is `todos`, `todo-archive`

#### Scenario: Case-insensitive
- **WHEN** a note `Meeting Notes` exists and the query is `mn`
- **THEN** `Meeting Notes` is listed

#### Scenario: Query is literal
- **WHEN** a note `foo` exists and the query is `foo.md`
- **THEN** `foo` is not listed, because its displayed name `foo` does not
  contain the characters `.md`

#### Scenario: Selection resets on query change
- **WHEN** the user has moved the selection to the third row and then types
  another character
- **THEN** the first row of the new result list is selected

### Requirement: Create row
When the trimmed query is non-empty and no note's displayed name equals it
exactly, the finder SHALL append a final row offering to create a note named
after the query. The create row SHALL always be last, and SHALL be the only
row, and selected, when nothing matches.

#### Scenario: Create row below matches
- **WHEN** a note `todos` exists and the query is `todo`
- **THEN** the list shows `todos` selected, followed by a last row
  `Create "todo"`

#### Scenario: Create row alone
- **WHEN** no note matches the query `meeting-notes`
- **THEN** the only row is `Create "meeting-notes"` and it is selected

#### Scenario: No create row for an exact name
- **WHEN** a note `todo` exists and the query is `todo`
- **THEN** no create row is shown

#### Scenario: No create row for the hidden current note
- **WHEN** `todo` is the current note and the query is `todo`
- **THEN** the list is empty: `todo` is hidden and no create row is shown

#### Scenario: No create row for an empty query
- **WHEN** the query is empty or only whitespace
- **THEN** no create row is shown

### Requirement: Navigation
Up/Down and `ctrl+p`/`ctrl+n` SHALL move the selection one row. The selection
SHALL stop at the first and last row without wrapping. Any other printable
key SHALL edit the query.

#### Scenario: Moving down to the create row
- **WHEN** the list has two matches and a create row, and the user presses
  Down twice from the top
- **THEN** the create row is selected

#### Scenario: No wrapping
- **WHEN** the last row is selected and the user presses Down
- **THEN** the last row stays selected

### Requirement: Opening a note
Enter on a note row SHALL make that note the current note and open it in
preview. The query SHALL be cleared when the finder is next entered.

#### Scenario: Open match
- **WHEN** `todos` is selected and the user presses Enter
- **THEN** preview opens for `todos`, and on returning to the finder the
  query is empty and `todos` is hidden

#### Scenario: Enter on an empty list
- **WHEN** the list has no rows and the user presses Enter
- **THEN** nothing happens

### Requirement: Creating a note
Enter on the create row SHALL create an empty file `<query>.md` in the notes
directory, make it the current note and open it in edit mode. The name is
the trimmed query, used literally. A name SHALL be rejected, with an error
shown in the finder and no file created, when it is empty, contains `/` or
`\`, or is `.` or `..`. Any other failure to create the file SHALL be shown
as an error in the finder.

#### Scenario: Create as typed
- **WHEN** the query is `foo.md` and the user presses Enter on the create row
- **THEN** a file `foo.md.md` is created and edit mode opens for it

#### Scenario: Name with a separator is rejected
- **WHEN** the query is `a/b` and the user presses Enter on the create row
- **THEN** no file is created and the finder shows an error about the
  invalid name

#### Scenario: Creation fails on disk
- **WHEN** the notes directory is not writable and the user presses Enter on
  the create row
- **THEN** no file is created, the finder shows the OS error and stays open

### Requirement: Leaving the finder
Esc SHALL return to preview of the current note when there is one, and
SHALL quit clipt when there is none. Esc SHALL NOT clear the query.
`ctrl+c` SHALL always quit. `ctrl+o` SHALL open the notes directory picker.

#### Scenario: Esc with a current note
- **WHEN** `todo` is the current note, the query is `ide`, and the user
  presses Esc
- **THEN** preview opens for `todo`

#### Scenario: Esc without a current note
- **WHEN** no note is current and the user presses Esc
- **THEN** clipt exits

#### Scenario: Current note disappeared
- **WHEN** `todo` is the current note, its file has been removed from disk,
  and the user presses Esc
- **THEN** the finder stays open, shows an error that `todo` no longer
  exists, and clears the current note

### Requirement: Empty directory
When the directory contains no notes and the query is empty, the finder
SHALL show a hint that typing a name and pressing Enter creates the first
note.

#### Scenario: Fresh directory
- **WHEN** the finder opens on a directory with no notes
- **THEN** the list is empty and the hint is shown
