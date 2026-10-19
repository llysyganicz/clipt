## MODIFIED Requirements

### Requirement: Fuzzy matching
With a non-empty name query (see the tags capability's tag query) the finder
SHALL show the notes whose displayed name contains every character of the
name query in order (a subsequence match), compared case-insensitively.
Matching notes SHALL keep the default order (modification time, newest
first); there is no relevance scoring. Apart from case folding, the name
query is matched literally against the displayed name; no extension handling
or other normalization is applied. Changing the query SHALL reset the
selection to the top row.

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
When the trimmed name query is non-empty and no note's displayed name equals
it exactly, the finder SHALL append a final row offering to create a note
named after the name query. The create row SHALL show the query's tag
filters after the name, in the tag display form. The create row SHALL always
be last, and SHALL be the only row, and selected, when nothing matches.

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

#### Scenario: Create row with tags
- **WHEN** no note is named `standup` and the query is `#work standup`
- **THEN** the last row is `Create "standup" #work`

#### Scenario: No create row for tags alone
- **WHEN** the query is `#work`
- **THEN** no create row is shown

#### Scenario: Exact name with a tag filter
- **WHEN** a note `standup` without tags exists and the query is
  `#work standup`
- **THEN** the list is empty and no create row is shown

### Requirement: Creating a note
Enter on the create row SHALL create a file `<name>.md` in the notes
directory, where the name is the trimmed name query used literally, make it
the current note and open it in edit mode. When the query has no tag filter
the file SHALL be empty. When it has tag filters the file SHALL hold only a
front matter block whose `tags` lists the filters' text in query order,
without duplicates compared case-insensitively, and edit mode SHALL open
with the cursor on an empty line after the front matter. A name SHALL be
rejected, with an error shown in the finder and no file created, when it is
empty, contains `/` or `\`, or is `.` or `..`. Any other failure to create
the file SHALL be shown as an error in the finder.

#### Scenario: Create as typed
- **WHEN** the query is `foo.md` and the user presses Enter on the create row
- **THEN** a file `foo.md.md` is created and edit mode opens for it

#### Scenario: Create with tags
- **WHEN** the query is `#work #meeting standup` and the user presses Enter
  on the create row
- **THEN** `standup.md` reads `---\ntags: [work, meeting]\n---\n` and edit
  mode opens with the cursor on the line after the closing `---`

#### Scenario: Tag filters are written as typed
- **WHEN** a note tagged `work` exists, the query is `#wo standup` and the
  user presses Enter on the create row
- **THEN** `standup.md` has the tag `wo`

#### Scenario: Name with a separator is rejected
- **WHEN** the query is `a/b` and the user presses Enter on the create row
- **THEN** no file is created and the finder shows an error about the
  invalid name

#### Scenario: Creation fails on disk
- **WHEN** the notes directory is not writable and the user presses Enter on
  the create row
- **THEN** no file is created, the finder shows the OS error and stays open

### Requirement: Creating a template
In template scope the create row SHALL create `.templates/<name>.md`,
creating the `.templates` directory first when needed, write the template
skeleton into it and open it in edit mode. The skeleton's `tags` SHALL list
the query's tag filters in query order, without duplicates compared
case-insensitively, and SHALL be empty when the query has none. Name
validation and error display SHALL be the same as for notes.

#### Scenario: Create a template
- **WHEN** no `.templates` directory exists, the query is `meeting` and the
  user presses Enter on the create row
- **THEN** `.templates/meeting.md` exists with the skeleton content and edit
  mode opens on it

#### Scenario: Create a template with tags
- **WHEN** the query in template scope is `#meeting standup` and the user
  presses Enter on the create row
- **THEN** `.templates/standup.md` holds the skeleton with the tags
  `meeting`

#### Scenario: Invalid template name
- **WHEN** the query is `a/b` in template scope and the user presses Enter on
  the create row
- **THEN** nothing is created and the finder shows the invalid name error

## ADDED Requirements

### Requirement: Tag filter
The finder SHALL apply the query's tag filters (see the tags capability's
tag query) in both scopes: in notes scope against each note's tags, in
template scope against each template's `tags`. A note or template with
invalid front matter has no tags; it SHALL still be listed when the query
has no tag filter and no error SHALL be shown for it. Tags SHALL be re-read
from disk every time the finder is entered, together with the listing.

#### Scenario: Filter notes by tag
- **WHEN** `standup` has the tag `work`, `groceries` has the tag `home`, and
  the query is `#work`
- **THEN** the list shows `standup` only

#### Scenario: Filter templates by tag
- **WHEN** template `meeting` has the tag `meeting`, template `daily` has
  none, and the query in template scope is `#meet`
- **THEN** the list shows `meeting` only

#### Scenario: Broken front matter
- **WHEN** `draft` has unclosed front matter and the query is empty
- **THEN** `draft` is listed without tags and no error is shown

#### Scenario: Tags edited on disk
- **WHEN** the user adds the tag `work` to `ideas` in edit mode and returns
  to the finder with the query `#work`
- **THEN** `ideas` is listed

### Requirement: Tags on rows
Each note or template row SHALL show the item's tags after its name, in the
tag display form, in a dimmed style. When the row does not fit the terminal
width the tags SHALL be cut from the end and marked with `…`; the name
SHALL NOT be shortened to make room for tags, and the tags SHALL be omitted
when there is no room for them. Items without tags show the name only.

#### Scenario: Tags shown
- **WHEN** `standup` has the tags `work` and `meeting`
- **THEN** its row shows `standup` followed by `#work #meeting`

#### Scenario: Narrow terminal
- **WHEN** the terminal is too narrow for a row's name and all its tags
- **THEN** the name is shown in full and the tags are cut with `…`

#### Scenario: Untagged note
- **WHEN** `ideas` has no tags
- **THEN** its row shows `ideas` only
