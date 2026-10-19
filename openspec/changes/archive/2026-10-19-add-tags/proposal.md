## Why

Notes can only be found by name, so a group of related notes (everything for
a project, all meeting notes) has no shared handle. Tags declared in YAML
front matter give notes that handle without folders, and the finder can use
them to narrow the list.

## What Changes

- Notes MAY start with YAML front matter (same `---` fences as templates)
  with a `tags` list. Front matter stays on disk and in the editor exactly as
  written.
- Finder: query tokens starting with `#` filter by tag (case-insensitive
  prefix, several tags AND-ed); the rest of the query is fuzzy-matched on the
  name as today. Rows show the note's tags dimmed after the name.
- Finder create row: creating `standup` from `#work standup` writes a note
  whose front matter holds `tags: [work]`.
- Template scope and the template picker filter templates by their own tags
  with the same syntax, and show them on rows.
- Preview: the front matter is no longer rendered as markdown; a header row
  above the content shows the note's tags.
- Templates: front matter gains a `tags` key next to `form`; the skeleton
  includes `tags: []` (filled from the query when created from a filtered
  template scope).
- Template insert: inserting a template merges its tags into the note's
  front matter (appended, deduplicated), creating the front matter when the
  note has none. When the note's front matter is broken, the body is still
  inserted, the merge is skipped and an error is shown.
- Front-matter splitting is shared by notes and templates so both agree on
  what front matter is.

## Non-goals

- Inline `#tags` in the note body.
- A tag browser, tag list, tag counts or tag completion.
- Renaming or deleting a tag across notes.
- Nested or hierarchical tags (`work/clientA` is just a tag string).
- Any other front matter field in the preview header or the finder.
- `{{placeholders}}` inside template tags; template tags are literal.
- Searching for a note whose name contains a `#`-prefixed word by that word.
- Note content search.

## Capabilities

### New Capabilities
- `tags`: note front matter, the `tags` key and its accepted forms, tag
  query matching, and merging tags into a note's front matter.

### Modified Capabilities
- `finder`: `#tag` filter tokens, tags shown on rows, create row writes the
  query's tags, template scope filters and creates with tags.
- `preview`: front matter is not rendered; a tags header row above the
  content.
- `templates`: `tags` key in template front matter; skeleton gains `tags`.
- `template-insert`: picker filters by `#tag` and shows tags; inserting a
  template merges its tags into the note.

## Impact

- New package for front matter splitting, used by `internal/templates` and
  the new note metadata parsing (in `internal/notes` or a sibling package).
- `internal/notes`: listing also reads each note's front matter (up to the
  closing fence) so the finder can filter and show tags; creating a note can
  write initial content.
- `internal/ui/finder`: query parsing, row rendering with tags, width
  handling, create with tags.
- `internal/ui/preview`: header row, body rendered without front matter.
- `internal/ui/tmpl`: `#tag` filtering, tags on rows, tags passed along with
  the expanded body.
- `internal/ui/editor`: insert that merges tags into the buffer's front
  matter and keeps the insert position.
- `internal/ui/theme`: styles for tags on rows and the preview header.
- No new dependencies (goccy/go-yaml is already used).
- `openspec/config.yaml` still lists tags as out of scope for v1; that line
  is stale and is not changed by this change.
