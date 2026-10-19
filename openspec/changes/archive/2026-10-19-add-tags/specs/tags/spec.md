## Purpose

Tags group notes and templates without folders: they are declared in a
file's YAML front matter, matched by `#tag` tokens in a search query, and
merged into a note when a tagged template is inserted.

## ADDED Requirements

### Requirement: Note front matter
A note SHALL start with front matter when its first line is exactly `---`;
the front matter ends at the next line that is exactly `---`. Front matter
is YAML. Windows line endings SHALL be treated as `\n` before parsing. This
is the same rule templates use. The body of a note is everything after the
closing `---` line, or the whole file when there is no front matter. clipt
SHALL NOT rewrite front matter on its own: it stays on disk and in the
editor exactly as written, except where a requirement below changes it.

#### Scenario: Note with front matter
- **WHEN** a note reads `---\ntags: [work]\n---\n# Standup\n`
- **THEN** it has the tag `work` and the body `# Standup\n`

#### Scenario: Note without front matter
- **WHEN** a note reads `# Standup\n---\n`
- **THEN** it has no front matter and no tags

#### Scenario: Front matter kept as written
- **WHEN** a note with front matter is opened in edit mode and saved without
  changes
- **THEN** the file's front matter is unchanged

### Requirement: Tags key
The front matter's `tags` key SHALL be a YAML sequence (flow `[a, b]` or
block `- a` form) of scalar values. Each value's tag is its text with
surrounding whitespace removed; empty values SHALL be ignored, and a value
equal to an earlier one compared case-insensitively SHALL be ignored. Tags
keep their order and their case as written. Front matter without a `tags`
key, or with an empty list, has no tags. Any other front matter key SHALL be
ignored. The front matter is invalid when it is not closed, is not valid
YAML, or its `tags` value is not a sequence of scalars; a note with invalid
front matter has no tags.

#### Scenario: Flow list
- **WHEN** the front matter is `tags: [work, meeting]`
- **THEN** the tags are `work` and `meeting`, in that order

#### Scenario: Block list
- **WHEN** the front matter is `tags:\n  - work\n  - meeting`
- **THEN** the tags are `work` and `meeting`, in that order

#### Scenario: Non-string scalar
- **WHEN** the front matter is `tags: [2026, work]`
- **THEN** the tags are `2026` and `work`

#### Scenario: Duplicates and blanks
- **WHEN** the front matter is `tags: [Work, "", work, home]`
- **THEN** the tags are `Work` and `home`

#### Scenario: Tags is a plain string
- **WHEN** the front matter is `tags: work`
- **THEN** the front matter is invalid and the note has no tags

#### Scenario: Unclosed front matter
- **WHEN** a note's first line is `---` and no later line is exactly `---`
- **THEN** the front matter is invalid and the note has no tags

#### Scenario: Other keys
- **WHEN** the front matter is `title: Standup\ntags: [work]`
- **THEN** the tags are `work` and `title` has no effect

### Requirement: Tag query
A search query SHALL be split on whitespace into tokens. A token that starts
with `#` and has at least one more character is a tag filter for the text
after the `#`; a token that is only `#` SHALL be ignored. The remaining
tokens, joined with single spaces, form the name query. When the query has
no tag filter and no bare `#` token, the whole query SHALL be the name query
unchanged. An item matches when the name query fuzzy-matches its name and,
for every tag filter, at least one of its tags starts with that filter,
compared case-insensitively.

#### Scenario: Tag filter
- **WHEN** `standup` has the tags `work`, `ideas` has the tag `home`, and the
  query is `#work`
- **THEN** `standup` matches and `ideas` does not

#### Scenario: Prefix match
- **WHEN** `standup` has the tag `work` and the query is `#wo`
- **THEN** `standup` matches

#### Scenario: Case-insensitive
- **WHEN** `standup` has the tag `Work` and the query is `#WORK`
- **THEN** `standup` matches

#### Scenario: Several filters
- **WHEN** `standup` has the tags `work` and `meeting`, `review` has the tag
  `work`, and the query is `#work #meet`
- **THEN** `standup` matches and `review` does not

#### Scenario: Tags and name together
- **WHEN** `standup` and `review` both have the tag `work` and the query is
  `#work sta`
- **THEN** `standup` matches and `review` does not

#### Scenario: Filter position does not matter
- **WHEN** the query is `sta #work`
- **THEN** it matches the same items as `#work sta`

#### Scenario: Bare hash ignored
- **WHEN** the query is `# sta`
- **THEN** it matches the same items as `sta`

#### Scenario: Untagged item and a tag filter
- **WHEN** `ideas` has no tags and the query is `#work`
- **THEN** `ideas` does not match

### Requirement: Tag display
Wherever tags are shown, each tag SHALL be written as `#` followed by the
tag, separated by single spaces, in the item's tag order.

#### Scenario: Display
- **WHEN** a note has the tags `work` and `meeting`
- **THEN** its tags are shown as `#work #meeting`

### Requirement: Merging tags into a note
Merging a list of tags into a note's text SHALL add every tag that the note
does not already have, compared case-insensitively, after the existing tags
and in the list's order, and SHALL leave every other line of the text
unchanged:
- no front matter: a front matter block holding only the `tags` key SHALL be
  added before the first line;
- front matter without a `tags` key: a `tags` entry SHALL be added as the
  last line of the front matter;
- front matter with a `tags` key: the lines of the `tags` entry SHALL be
  replaced by a single line in flow form holding all tags.
Written tags SHALL be quoted when needed, so that the merged front matter
parses back to the merged tags. When no tag needs adding the text SHALL be
unchanged. When the note's front matter is invalid the merge SHALL be
refused with an error and the text left unchanged.

#### Scenario: Note without front matter
- **WHEN** `[meeting]` is merged into `# Standup`
- **THEN** the text is `---\ntags: [meeting]\n---\n# Standup`

#### Scenario: Front matter without tags
- **WHEN** `[meeting]` is merged into `---\ntitle: Standup\n---\nbody`
- **THEN** the text is `---\ntitle: Standup\ntags: [meeting]\n---\nbody`

#### Scenario: Existing block list
- **WHEN** `[meeting]` is merged into
  `---\ntags:\n  - work\ntitle: x\n---\nbody`
- **THEN** the text is `---\ntags: [work, meeting]\ntitle: x\n---\nbody`

#### Scenario: Already present
- **WHEN** `[Work]` is merged into a note with the tags `work`
- **THEN** the text is unchanged

#### Scenario: Tag needing quotes
- **WHEN** `["a, b"]` is merged into a note without front matter
- **THEN** the merged front matter parses to the single tag `a, b`

#### Scenario: Invalid front matter
- **WHEN** `[meeting]` is merged into a note whose front matter is
  `tags: [work`
- **THEN** the merge fails with an error and the text is unchanged
