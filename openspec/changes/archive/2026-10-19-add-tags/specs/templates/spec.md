## MODIFIED Requirements

### Requirement: Form definition
The front matter's `form` key SHALL be a list of fields. Each field SHALL have
a `name` and a `type`; `label` is optional and defaults to the name. `name`
SHALL be non-empty and SHALL NOT contain whitespace, `{` or `}`. Two fields
SHALL NOT share a name. Front matter without a `form` key, or with an empty
list, defines no fields. Any key in the front matter other than `form` and
`tags` SHALL be ignored.

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

### Requirement: New template skeleton
A template created from clipt SHALL start with a front matter skeleton: a
`tags` list (empty unless the creating flow supplies tags), a `form` list
with one example `text` field, followed by a body containing the matching
placeholder, so the user sees the expected shape.

#### Scenario: Skeleton parses
- **WHEN** the user creates a template and changes nothing
- **THEN** the file parses without error, has no tags and exactly one
  `text` field

#### Scenario: Skeleton shows the tags key
- **WHEN** the user creates a template without tags
- **THEN** its front matter contains `tags: []`

## ADDED Requirements

### Requirement: Template tags
The template front matter's `tags` key SHALL follow the tags capability's
tags key rules. A `tags` value that is not a sequence of scalars SHALL be a
parse error. Template tags are literal: `{{name}}` placeholders inside tags
SHALL NOT be expanded.

#### Scenario: Template with tags
- **WHEN** a template's front matter is `tags: [meeting]` with a `form` list
- **THEN** it has the tag `meeting` and the fields from `form`

#### Scenario: Invalid tags
- **WHEN** a template's front matter is `tags: meeting`
- **THEN** parsing fails with an error about `tags`

#### Scenario: Placeholder in tags
- **WHEN** a template has `tags: ["{{topic}}"]` and a field `topic`
- **THEN** its tag is `{{topic}}` literally
