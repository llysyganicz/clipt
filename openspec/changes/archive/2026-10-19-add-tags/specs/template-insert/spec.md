## MODIFIED Requirements

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

## ADDED Requirements

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
