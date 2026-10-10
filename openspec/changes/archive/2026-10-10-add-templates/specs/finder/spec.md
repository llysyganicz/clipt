## ADDED Requirements

### Requirement: Template scope
The finder SHALL have two scopes: notes (the default) and templates.
`ctrl+t` SHALL switch to the other scope with an empty query and the top row
selected. In template scope the finder SHALL list the templates with the same
listing rules, order and fuzzy matching as notes, SHALL hide no row (there is
no current template), SHALL show the prompt as `templates> ` and SHALL show
hints that Enter edits and Esc returns to notes. A missing `.templates`
directory SHALL list as empty without an error. Entering the finder from any
other mode SHALL land in notes scope, except when returning from editing a
template.

#### Scenario: Toggle to templates
- **WHEN** the notes scope has the query `me` and the user presses `ctrl+t`
- **THEN** the template scope opens with an empty query listing every
  template, newest first

#### Scenario: Toggle back
- **WHEN** the user is in template scope and presses `ctrl+t`
- **THEN** the notes scope opens with an empty query

#### Scenario: Templates list only templates
- **WHEN** the notes directory holds `todo.md` and `.templates/meeting.md`
  and the user switches to template scope
- **THEN** the list shows `meeting` only

#### Scenario: No templates directory
- **WHEN** `.templates` does not exist and the user switches to template
  scope
- **THEN** the list is empty and no error is shown

#### Scenario: Current note is not hidden in template scope
- **WHEN** `todo` is the current note and a template is also named `todo`
- **THEN** template scope lists the template `todo`

### Requirement: Opening a template
Enter on a template row SHALL open that template in edit mode without
changing the current note. Leaving edit mode on a template SHALL return to
the finder in template scope with an empty query.

#### Scenario: Edit a template
- **WHEN** `todo` is the current note and the user presses Enter on the
  template `meeting`
- **THEN** edit mode opens on `.templates/meeting.md`, and after Esc the
  finder is back in template scope with `todo` still the current note

### Requirement: Creating a template
In template scope the create row SHALL create `.templates/<query>.md`,
creating the `.templates` directory first when needed, write the template
skeleton into it and open it in edit mode. Name validation and error display
SHALL be the same as for notes.

#### Scenario: Create a template
- **WHEN** no `.templates` directory exists, the query is `meeting` and the
  user presses Enter on the create row
- **THEN** `.templates/meeting.md` exists with the skeleton content and edit
  mode opens on it

#### Scenario: Invalid template name
- **WHEN** the query is `a/b` in template scope and the user presses Enter on
  the create row
- **THEN** nothing is created and the finder shows the invalid name error

## MODIFIED Requirements

### Requirement: Leaving the finder
In notes scope, Esc SHALL return to preview of the current note when there is
one, and SHALL quit clipt when there is none. Esc SHALL NOT clear the query.
In template scope, Esc SHALL return to notes scope with an empty query.
`ctrl+c` SHALL always quit. `ctrl+o` SHALL open the notes directory picker
from either scope, and accepting or cancelling it lands in notes scope.

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

#### Scenario: Esc in template scope
- **WHEN** no note is current, the user is in template scope and presses Esc
- **THEN** the finder shows notes scope with an empty query and clipt keeps
  running

#### Scenario: Change directory from template scope
- **WHEN** the user presses `ctrl+o` in template scope and cancels the picker
- **THEN** the finder is in notes scope

### Requirement: Empty directory
When the directory contains no notes and the query is empty, the finder
SHALL show a hint that typing a name and pressing Enter creates the first
note. In template scope with no templates and an empty query, it SHALL show
the same kind of hint for the first template.

#### Scenario: Fresh directory
- **WHEN** the finder opens on a directory with no notes
- **THEN** the list is empty and the hint is shown

#### Scenario: No templates yet
- **WHEN** the user switches to template scope and no template exists
- **THEN** the list is empty and a hint about creating the first template is
  shown
