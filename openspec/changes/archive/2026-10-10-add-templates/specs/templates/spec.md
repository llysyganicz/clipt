## Purpose

Reusable markdown skeletons stored next to the notes: a template file is
markdown with an optional YAML front matter that declares a small form, and
its body is expanded by replacing `{{name}}` placeholders with the values
entered in that form.

## ADDED Requirements

### Requirement: Template location
A template is a regular file whose name ends in `.md` and does not start
with `.`, stored in the `.templates` subdirectory of the notes directory. Its
displayed name is the file name without the trailing `.md`. Templates SHALL
NOT appear in the notes listing. A missing `.templates` directory SHALL be
treated as an empty template set, and it SHALL be created when the first
template is created from clipt.

#### Scenario: Templates are not notes
- **WHEN** the notes directory contains `todo.md` and `.templates/meeting.md`
- **THEN** the notes listing shows `todo` only and the template set contains
  `meeting` only

#### Scenario: No templates directory
- **WHEN** the notes directory has no `.templates` subdirectory
- **THEN** the template set is empty and no error is reported

#### Scenario: Directory created on first template
- **WHEN** no `.templates` directory exists and the user creates the template
  `meeting`
- **THEN** `.templates/meeting.md` exists afterwards

### Requirement: Front matter
A template SHALL start with front matter when its first line is exactly
`---`; the front matter ends at the next line that is exactly `---`. Front
matter is YAML. A template whose first line is not `---` has no front matter
and no fields. Windows line endings SHALL be treated as `\n` before parsing.

#### Scenario: Template with front matter
- **WHEN** a template reads `---\nform:\n  - name: topic\n    type: text\n---\n# {{topic}}\n`
- **THEN** it has one field `topic` and the body `# {{topic}}`

#### Scenario: Template without front matter
- **WHEN** a template reads `# Daily\n\n- [ ] \n`
- **THEN** it has no fields and the body is the whole file

#### Scenario: Unterminated front matter
- **WHEN** a template starts with `---` and no later line is exactly `---`
- **THEN** parsing fails with an error saying the front matter is not closed

### Requirement: Form definition
The front matter's `form` key SHALL be a list of fields. Each field SHALL have
a `name` and a `type`; `label` is optional and defaults to the name. `name`
SHALL be non-empty and SHALL NOT contain whitespace, `{` or `}`. Two fields
SHALL NOT share a name. Front matter without a `form` key, or with an empty
list, defines no fields. Any other key in the front matter SHALL be ignored.

#### Scenario: Label defaults to name
- **WHEN** a field is declared with `name: topic` and `type: text` and no
  label
- **THEN** its label is `topic`

#### Scenario: Missing name
- **WHEN** a field is declared with `type: text` and no `name`
- **THEN** parsing fails with an error naming the field's position in the
  list

#### Scenario: Duplicate name
- **WHEN** two fields are both named `topic`
- **THEN** parsing fails with an error naming `topic`

#### Scenario: Name with braces
- **WHEN** a field is named `to}pic`
- **THEN** parsing fails with an error naming the invalid name

#### Scenario: Other front matter keys
- **WHEN** the front matter contains `title: Meeting` and a `form` list
- **THEN** the fields come from `form` and `title` has no effect

#### Scenario: Invalid YAML
- **WHEN** the front matter is `form: [`
- **THEN** parsing fails with the YAML error

### Requirement: Field types
The supported field types are `text`, `number` and `option`. A `text` field
has an empty default. A `number` field has the default `0` and accepts any
decimal number, as typed. An `option` field SHALL declare a non-empty
`options` list of strings and defaults to the first entry. Any other type,
and an `option` field without options, SHALL be a parse error.

#### Scenario: Number default
- **WHEN** a field has `type: number`
- **THEN** its default value is `0`

#### Scenario: Option default
- **WHEN** a field has `type: option` and `options: [low, high]`
- **THEN** its default value is `low`

#### Scenario: Unknown type
- **WHEN** a field has `type: date`
- **THEN** parsing fails with an error naming the unsupported type

#### Scenario: Option without options
- **WHEN** a field has `type: option` and no `options`
- **THEN** parsing fails with an error naming the field

### Requirement: Body
The body is everything after the closing `---` line, or the whole file when
there is no front matter. Leading lines that are empty or contain only
whitespace SHALL be removed from the body, and trailing newlines SHALL be
removed.

#### Scenario: Blank line after front matter
- **WHEN** the file is `---\nform: []\n---\n\n# Title\n`
- **THEN** the body is `# Title`

#### Scenario: Several blank lines after front matter
- **WHEN** the closing `---` is followed by three empty lines and `# Title`
- **THEN** the body is `# Title`

#### Scenario: Blank lines inside the body are kept
- **WHEN** the body text is `# Title\n\ntext\n`
- **THEN** the body is `# Title\n\ntext`

### Requirement: Expansion
Expanding a template SHALL replace every occurrence of `{{name}}`, for each
declared field `name`, with that field's value. A placeholder that matches
no field SHALL be left unchanged. Values SHALL be inserted literally; a
placeholder inside a value SHALL NOT be expanded. Placeholders are matched
exactly; `{{ name }}` with spaces is not a placeholder.

#### Scenario: Replace placeholders
- **WHEN** the body is `# {{topic}} ({{prio}})` and the values are
  `topic=Standup`, `prio=high`
- **THEN** the result is `# Standup (high)`

#### Scenario: Repeated placeholder
- **WHEN** the body is `{{a}} and {{a}}` and `a=x`
- **THEN** the result is `x and x`

#### Scenario: Unknown placeholder
- **WHEN** the body is `{{topic}} {{date}}` and only `topic` is a field
- **THEN** the result keeps `{{date}}` literally

#### Scenario: Value containing a placeholder
- **WHEN** the body is `{{a}} {{b}}`, `a={{b}}` and `b=y`
- **THEN** the result is `{{b}} y`

### Requirement: New template skeleton
A template created from clipt SHALL start with a front matter skeleton: a
`form` list with one example `text` field, followed by a body containing the
matching placeholder, so the user sees the expected shape.

#### Scenario: Skeleton parses
- **WHEN** the user creates a template and changes nothing
- **THEN** the file parses without error and has exactly one `text` field
