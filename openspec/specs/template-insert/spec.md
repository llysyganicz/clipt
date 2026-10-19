# template-insert Specification

## Purpose

The insert mode entered from the editor: pick a template by fuzzy name,
fill in its form, and have the expanded body inserted at the cursor, or back
out with Esc leaving the note untouched.

## Requirements

### Requirement: Template picker
Pressing `ctrl+t` in edit mode on a note SHALL open the template picker. The
picker SHALL list the templates by modification time, newest first, with ties
broken by name, and SHALL filter them with the same matching the finder
uses: the case-insensitive subsequence match on the name query and the tag
filters of the tags capability's tag query, against each template's tags.
Each row SHALL show the template's tags as the finder does. There is no
create row. Up/Down and `ctrl+p`/`ctrl+n` SHALL move the selection without
wrapping, any other printable key edits the query, and changing the query
resets the selection to the top. When there are no templates the picker
SHALL show a hint that templates are created from the finder with `ctrl+t`.

#### Scenario: Open the picker
- **WHEN** templates `meeting` and `daily` exist and the user presses
  `ctrl+t` in edit mode
- **THEN** the picker lists both, newest first, with the top row selected

#### Scenario: Filter
- **WHEN** the picker is open and the user types `mt`
- **THEN** `meeting` is listed and `daily` is not

#### Scenario: Filter by tag
- **WHEN** `meeting` has the tag `work`, `daily` has none, and the user
  types `#work`
- **THEN** `meeting` is listed and `daily` is not

#### Scenario: Tags on rows
- **WHEN** `meeting` has the tag `work`
- **THEN** its row shows `meeting` followed by `#work`

#### Scenario: No templates
- **WHEN** there are no templates and the user presses `ctrl+t`
- **THEN** the picker opens with an empty list and the hint about creating
  templates from the finder

### Requirement: Picking a template
Enter on a template SHALL read and parse it. When it has no fields its body
SHALL be inserted immediately. When it has fields the form SHALL open. When
the file cannot be read or parsed the picker SHALL stay open and show the
error. Enter on an empty list SHALL do nothing.

#### Scenario: Template without fields
- **WHEN** the selected template has no front matter and the user presses
  Enter
- **THEN** its body is inserted into the note and edit mode is back

#### Scenario: Template with fields
- **WHEN** the selected template declares two fields and the user presses
  Enter
- **THEN** the form opens with two inputs

#### Scenario: Invalid template
- **WHEN** the selected template has an unknown field type and the user
  presses Enter
- **THEN** the picker stays open and shows the parse error

### Requirement: Form
The form SHALL show one row per field in declaration order, each with the
field's label and an input, and the first row focused. A `text` input starts
empty, a `number` input starts as `0`, and an `option` input shows the first
option. Down, Tab and Enter SHALL move focus to the next row; Up and
Shift+Tab SHALL move to the previous row. Left and Right on an `option` row
SHALL cycle through its options without wrapping past the ends. Printable
keys on a `text` or `number` row edit its value.

#### Scenario: Initial values
- **WHEN** the form opens for fields `topic` (text), `count` (number) and
  `prio` (option: low, high)
- **THEN** the rows show an empty input, `0`, and `low`, and the first row is
  focused

#### Scenario: Cycle an option
- **WHEN** the `prio` row is focused and the user presses Right
- **THEN** it shows `high`; pressing Right again keeps `high`

#### Scenario: Move between rows
- **WHEN** the first of three rows is focused and the user presses Enter
- **THEN** the second row is focused and no insertion happens

### Requirement: Submit
Enter on the last row SHALL submit the form. On submit every `number` value
SHALL be a decimal number; otherwise the form SHALL stay open with an error
naming the field and focus on it. When valid, the body SHALL be expanded with
the entered values and inserted into the note at the cursor, and edit mode
SHALL resume with the cursor after the inserted text.

#### Scenario: Submit valid values
- **WHEN** the body is `# {{topic}}`, the user types `Standup` in the
  `topic` row and presses Enter on the last row
- **THEN** `# Standup` is inserted at the cursor and edit mode is back

#### Scenario: Invalid number
- **WHEN** the `count` row holds `abc` and the user submits
- **THEN** the form stays open, shows an error naming `count`, and focuses
  that row

#### Scenario: Decimal number
- **WHEN** the `count` row holds `1.50` and the user submits
- **THEN** `{{count}}` expands to `1.50`

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

### Requirement: Tag merge on insert
Whenever a template's body is inserted into the note (a template without
fields, or a submitted form), the template's tags SHALL be merged into the
note following the tags capability's merge rules, applied to the note's
text after the body has been inserted. The inserted body SHALL stay where
the cursor was, and the cursor SHALL end up after the inserted text, also
when the merge adds or removes lines above it. When the template has no
tags, insertion SHALL behave as without this requirement. When the merge is
refused because the note's front matter is invalid, the body SHALL still be
inserted, the front matter left unchanged and edit mode SHALL show the
error.

#### Scenario: Note without front matter
- **WHEN** the note reads `# Standup` with the cursor at the end, and the
  template `meeting` with the tag `meeting` and body `- agenda` is inserted
- **THEN** the note reads `---\ntags: [meeting]\n---\n# Standup- agenda` and
  the cursor is after `- agenda`

#### Scenario: Tags appended
- **WHEN** the note has `tags: [work]` and the inserted template has the
  tags `meeting` and `Work`
- **THEN** the note's tags line is `tags: [work, meeting]`

#### Scenario: Template without tags
- **WHEN** a template without tags is inserted into a note without front
  matter
- **THEN** the note gains the body only and no front matter

#### Scenario: Broken note front matter
- **WHEN** the note's front matter is `tags: [work` and a template with the
  tag `meeting` is inserted
- **THEN** the body is inserted, the front matter is unchanged, and edit
  mode shows an error that the tags could not be merged
