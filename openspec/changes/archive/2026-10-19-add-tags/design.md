## Context

See proposal.md for motivation and the specs for behavior. Today:

- Only templates have front matter. `templates.Parse` splits it inline
  (first line `---`, next `---` closes, CRLF normalized) and unmarshals
  `form` with goccy/go-yaml.
- `notes.Store.List` builds `Note{Name, ModTime}` from `os.ReadDir` alone;
  no file is opened. The finder (`internal/ui/finder`) and the template
  picker (`internal/ui/tmpl`) both list through a `notes.Store` and filter
  with `finder.Match(query, name)`.
- Preview feeds the whole file to Glamour. A `---` / `tags: [...]` / `---`
  block renders as a horizontal rule plus a setext heading.
- The editor wraps a Bubbles textarea; `Model.Insert(text)` inserts at the
  cursor. The textarea has no line-level setter, so edits elsewhere replace
  the whole value and restore the cursor (as `dedent` and `reflow` do).
- New notes are created empty with `O_EXCL`; templates get `Skeleton`.

## Goals / Non-Goals

**Goals:**
- One definition of "front matter" shared by notes and templates.
- Tags available on every listed item without a second listing pass.
- Tag merging that never reformats front matter it does not own.

**Non-Goals:**
- A cache or index of tags across finder entries; the listing is re-read
  on every entry, as today.
- A general front matter model; only `tags` (and templates' `form`) are
  parsed.

## Decisions

### 1. A `frontmatter` package owns the fence rule

```
internal/frontmatter
  Split(content) (fm Block, body string, err error)
      Block{Present bool; Lines []string; Start, End int}  // line indexes
  Tags(raw string) ([]string, error)        // sequence-of-scalars + dedupe
  MergeTags(content string, add []string) (string, int /*line delta*/, error)
```

`templates.Parse` calls `Split` instead of its own loop; its error message
for an unclosed block stays the same. Tag parsing lives here too, so notes
and templates share the `tags` rules (scalar values via their text, trim,
drop empty, case-insensitive dedupe). Keeping it outside `notes` avoids a
`templates -> notes` dependency for parsing (templates already import
`notes` for the store, which is fine).

Alternative: put everything in `notes`. Rejected: templates would then
parse their front matter through the notes package, coupling two formats
that only share the fence rule.

### 2. `Note` carries tags; the listing reads only the head of each file

`notes.Note` gains `Tags []string`. `Store.List` opens each listed file
and reads line by line with a `bufio.Scanner` until the closing fence (or
stops after the first line when it is not `---`). Unclosed front matter
means reading the whole file, which is acceptable for a rare broken note.
Errors reading or parsing a single file give that note no tags and are not
reported (finder spec: broken front matter shows no error).

Both the notes store and the templates store are `notes.Store`, so the
finder's template scope and the template picker get template tags for free.
Template tags read this way use the same rules as `templates.Parse`.

Alternative: lazy loading tags only when the query contains `#`. Rejected:
rows always show tags.

### 3. Query parsing in the finder package

`finder.ParseQuery(q string) Query{Name string; Tags []string}` follows the
tag query rule (no `#` token -> `Name == q` unchanged, keeping today's
literal behavior including inner whitespace). `BuildRows` takes the parsed
query; a new `MatchItem(Query, Note)` combines `Match(q.Name, n.Name)` with
prefix tag checks. The template picker calls the same functions, so the
two lists cannot drift. The create row carries `Create` (name) and
`CreateTags` (deduped filter texts).

### 4. Creating with tags

`Store.Create(name)` stays as is (empty file, `O_EXCL`). The finder writes
initial content right after creating, as it already does for template
skeletons: notes get `---\ntags: [..]\n---\n`, templates get a skeleton
built by `templates.SkeletonWithTags(tags)` (`Skeleton` itself becomes the
no-tags case with `tags: []`). Tag lists are written with the same flow
serializer used by `MergeTags` (decision 6).

The editor loads content with one trailing newline trimmed, so a note that
is only front matter would put the cursor at the end of the closing `---`.
For a note created with tags the root opens the editor and calls
`Insert("\n")`, putting the cursor on the empty line after the fence. On
save, `notes.Normalize` trims that newline back, so an untouched note is
not rewritten differently.

### 5. Preview layout

```
+-----------------------------+
| #work #meeting              |  header: 0..n lines, fixed
|-----------------------------|
| rendered body (viewport)    |  height = h - 1 - headerLines
|                             |
| status bar                  |
+-----------------------------+
```

Preview calls `frontmatter.Split`; Glamour renders `body` (whole file when
unclosed). The header is rendered with a new `theme.Tag` style and
word-wrapped with Lip Gloss to the width; its line count is subtracted from
the viewport height in `render()`. Invalid front matter puts a single
`theme.Error` line in the header. No separator line is drawn beyond the
header itself, keeping the screen minimal.

### 6. Merge edits lines, not YAML

`MergeTags` works on lines of the text:

- No block -> prepend 3 lines (`---`, `tags: [...]`, `---`).
- Block without a `tags` key -> insert one line before the closing fence.
- Block with `tags` -> find the top-level `tags:` line (column 0) and the
  following lines that belong to it (indented, or starting with `- `),
  replace them with one flow line.

The YAML is still parsed (to read existing tags and to refuse invalid front
matter), but never re-marshalled as a whole, so comments, key order and
unknown keys survive. The flow line is produced by marshalling just the
string slice with goccy/go-yaml in flow style, which quotes values that
need it; a round-trip test guards this.

Alternative: unmarshal into `yaml.MapSlice` and marshal back. Rejected:
loses comments and formatting of keys clipt does not own.

### 7. Insert then merge, cursor by line delta

The template picker's `InsertMsg` gains `Tags []string`. A new
`editor.InsertTemplate(body, tags)`:

1. `Insert(body)` as today (cursor ends after the body).
2. If `tags` is empty, stop.
3. `MergeTags(value, tags)` -> new text and the number of lines added
   (positive) or removed (negative, e.g. a 3-line block list collapsed)
   above the cursor; all edited lines are inside the front matter, which
   starts at line 0, so the delta applies whenever the cursor is below the
   closing fence.
4. Replace the value and restore cursor at `(row+delta, col)`, the same way
   `dedent` restores it.
5. On error, keep the step 1 result and set the editor error line.

Merging after the insert keeps the definition purely textual and keeps the
"body at the cursor" behavior identical for untagged templates.

## Risks / Trade-offs

- [Finder entry now opens every note] -> read only up to the closing fence
  with a small buffer; a notes folder of a few thousand files stays well
  under a frame. Revisit with a cache only if it is noticed.
- [Line-based detection of the `tags` entry misreads exotic YAML (anchors,
  multi-line flow lists spanning lines, `tags` nested under another key)]
  -> only a column-0 `tags:` key counts; after replacing, the result is
  re-parsed and the merge is refused with an error if it no longer yields
  the expected tags, leaving the text unchanged.
- [Cursor inside the front matter when inserting] -> the body lands inside
  the front matter, which usually makes it invalid; the merge is then
  refused with an error and the user sees both. Accepted as an edge case.
- [Notes whose names contain `#word` cannot be searched by that word] ->
  stated as a non-goal; the rest of the name still fuzzy-matches.
- [Prefix filters become exact tags on create (`#wo` -> `wo`)] -> accepted;
  the tag is visible in the editor right away.
