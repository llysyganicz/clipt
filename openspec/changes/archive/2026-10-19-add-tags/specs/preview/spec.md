## MODIFIED Requirements

### Requirement: Rendered content
Preview SHALL show the current note's body (see the tags capability's note
front matter) rendered as markdown, word-wrapped to the terminal width, in a
scrollable area that fills the terminal between the tags header, when shown,
and the status bar. The front matter SHALL NOT be rendered as markdown. When
the front matter is not closed, the whole file SHALL be rendered as the body.
The note SHALL be re-read from disk every time preview is entered, and the
content SHALL start scrolled to the top. The rendered markdown SHALL use the
application's theme palette. Resizing the terminal SHALL re-render the
content at the new width.

#### Scenario: Open a note
- **WHEN** `todo` contains `# Todo` and a list, and the user opens it from
  the finder
- **THEN** preview shows the heading and list rendered, starting at the top

#### Scenario: Front matter not rendered
- **WHEN** `standup` reads `---\ntags: [work]\n---\n# Standup\n`
- **THEN** the content area shows the heading `Standup` and no `tags:` text

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

## ADDED Requirements

### Requirement: Tags header
When the current note has at least one tag, preview SHALL show a header
above the content area listing the tags in the tag display form, in the tag
style, wrapped to the terminal width. When the note has no tags and valid or
no front matter, there SHALL be no header and the content starts on the
first line. When the front matter is invalid, the header SHALL show an error
saying so instead of tags. The header SHALL stay in place while the content
scrolls.

#### Scenario: Tagged note
- **WHEN** `standup` has the tags `work` and `meeting`
- **THEN** the first line shows `#work #meeting` and the rendered body
  follows below it

#### Scenario: Untagged note
- **WHEN** `ideas` has front matter `title: Ideas` and no tags
- **THEN** there is no header and the first line belongs to the rendered
  body

#### Scenario: Header stays while scrolling
- **WHEN** a tagged note is taller than the content area and the user
  presses PgDn
- **THEN** the tags header is still shown at the top

#### Scenario: Invalid front matter
- **WHEN** the current note's front matter is `tags: work`
- **THEN** the header shows an error that the front matter is invalid
