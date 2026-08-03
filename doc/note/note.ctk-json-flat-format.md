# Knowledge.note.ctk-json-flat-format.md
============================================================

# CTK JSON Flat Format

## Status

This Note documents CTK JSON Flat Format, an optional line-oriented Workbench
representation for JSON-like Artifacts.

CTK does not require every implementation or Workbench to use this format. An
implementation may adopt it when a deterministic, reversible Flat
representation is useful, or may use another representation that preserves the
relevant Workbench responsibilities.

The current Go implementation uses this format. That is an implementation
resolution, not a declaration that Go defines the canonical CTK representation.
The format is not Required Source Compatibility and is not a canonical Cookbook
format.

The retained Bash implementation uses gron for Freeze Draft and does not need to
adopt this format. Workbench files from different implementations do not require
direct representation compatibility unless a separate Contract establishes it.

## Intent

CTK JSON Flat Format keeps gron's recognizable path-assignment model while
using JSON syntax for both property paths and values.

```text
<path>=<json-value>
```

The format is intended to be:

- line-oriented and easy to compare
- deterministic
- readable as key and value assignments
- reversible to JSON
- safe for dotted property names
- easy to divide among and combine from Ingredient candidates
- small enough for an implementation to parse and validate directly

It does not attempt to implement JavaScript or the complete gron statement
grammar. Review headings and first-property grouping belong to the Workbench
Draft presentation layer, not this Flat Format.

## Object representation

Each Object property is a quoted string path segment. Object containers are
declared explicitly.

```text
[]={}
["editor.fontSize"]=16
["editor"]={}
["editor", "fontSize"]=14
["files"]={}
["files", "associations"]={}
["files", "associations", "*.go"]="go"
```

This distinguishes a dotted property name from a nested path:

```text
["editor.fontSize"]=16
["editor", "fontSize"]=14
```

The empty path declares the root value:

```text
[]={}
```

## Replace and Union representation

The Flat Format provides two editing forms for JSON array composition:

- `[0]` is an indexed replacement value
- `[*]` and `[@name]` are Union members

No independent collection type is persisted. In the current Go integration,
the Merge Rule is simply the operation applied across Settings Resources:
`replace` or `union`.

### Indexed replacement

A numeric selector preserves explicit element order:

```text
["key"]=[]
["key", [0]]=100
["key", [1]]=200
```

The result is:

```json
{
  "key": [100, 200]
}
```

Replacement indices must be unique, start at zero, and remain contiguous. CTK does
not silently pad a missing index with `null`.

```text
["key", [0]]=100
["key", [2]]=200
```

The example above is invalid because index `1` is absent.

### Anonymous Union member

A wildcard selector declares an anonymous atomic value that participates in
Union:

```text
["key"]=[]
["key", [*]]=100
["key", [*]]=200
```

Union retains deterministic first-occurrence order from Cookbook Core Resource
resolution and each source array. Elements with identical canonical JSON values
are deduplicated.

### Named Union member

Workbench editing and Freeze Commit use a named Union selector when a
structured member must be edited through multiple assignments:

```text
["bindings"]=[]
["bindings", [@copy]]={}
["bindings", [@copy], "before"]=[]
["bindings", [@copy], "before", [0]]="<C-y>"
["bindings", [@copy], "commands"]=[]
["bindings", [@copy], "commands", [0]]="editor.action.clipboardCopyAction"
```

The Union selectors have different editing roles:

- `[*]` is an anonymous atomic Union member
- `[@name]` is a named structured Union member whose children may be traversed

Anonymous and named members may coexist at one Union path. Their completed
canonical values, rather than temporary names, determine membership. Indexed
replacement selectors such as `[0]` cannot coexist with Union selectors at the
same path because one Draft operation cannot request both replace and union.

`@name` is scoped to one Union path in one Workbench. It is an editing handle,
not persistent identity. Both `*` and `@name` selectors are discarded when
Freeze Commit materializes a normal JSON/JSONC Ingredient, and a later Freeze
does not regenerate either selector from Kitchen Notes.

Objects and arrays may be anonymous atomic Union members:

```text
["languages"]=[]
["languages", [*]]={"name":"go","enabled":true}
["languages", [*]]={"name":"rust","enabled":true}
```

One path must not mix replacement and Union selectors:

```text
["key"]=[]
["key", [0]]=100
["key", [*]]=200
```

The example above is an ambiguous merge operation rather than a collection type
error.

### Observation and author intent

Freeze Draft and View observe ordinary Runtime JSON arrays and therefore always
generate indexed selectors:

```text
["editor.rulers"]=[]
["editor.rulers", [0]]=120
["editor.rulers", [1]]=80
```

Kitchen Notes are not used to reconstruct `[*]` or `[@name]` during observation.
A human or AI author changes indexed selectors to Union selectors while
reviewing the Draft Workbench when values should compose across Settings
Resources:

```text
["editor.rulers"]=[]
["editor.rulers", [*]]=80
["editor.rulers", [*]]=120
```

Freeze Commit materializes ordinary JSON/JSONC arrays and discards both `*` and
temporary `@name` selectors. Neither selector is Ingredient data. Commit
records only the exact path as `union` in the Go Merge Rules Kitchen Note; rule
absence means later-value `replace`.

An empty array requires only its container declaration:

```text
["key"]=[]
```

## Ingredient division and composition

Assignments may be moved between Draft files when deciding Ingredient
responsibility. Object and array container declarations may appear in more than
one input and can be deduplicated when their declarations agree.

Editor candidate:

```text
[]={}
["editor"]={}
["editor", "fontSize"]=14
```

Files candidate:

```text
[]={}
["files"]={}
["files", "associations"]={}
["files", "associations", "*.go"]="go"
```

Union members may be collected from multiple Draft fragments committed into the
same target Settings Resource. Across resolved Settings Resources, paths absent
from Go Merge Rules retain current later-value `replace`; listed paths use
Union. Merge Rules never alter Cookbook Core Resource resolution order.

Go stores its Cookbook-local rules at:

```text
cookbook/kitchen-notes/go.merge-rules.yaml
```

Rules apply to exact logical Settings paths across the Cookbook. Go does not
select Merge Rules per Recipe. Freeze Commit automatically adds a Union Rule
when `[*]` or `[@name]` is committed. Because generated observation always uses
indexed selectors, `[0]` alone does not automatically remove an existing Union
Rule; initial removal is a direct Kitchen Note edit.

## Validation boundary

An implementation adopting this documented form should reject at least:

- malformed path or JSON value syntax
- missing root or required container declarations
- duplicate replacement indices
- gaps in replacement indices
- replacement and Union selectors under the same path
- duplicate `@name` declarations with incompatible values at one Union path
- traversal beneath an anonymous `[*]` element
- conflicting values at one Object or indexed replacement path
- a scalar used as the parent of another assignment
- Object, array, and scalar declaration conflicts

Identical Object and array container declarations may be deduplicated. Identical
Union values are deduplicated by canonical JSON value.

Errors should identify the source file and line. Invalid Workbench content must
not partially update a canonical Ingredient.

## Format boundary

The currently documented parser boundary is:

- UTF-8 input and output
- one assignment per physical line
- the first `=` outside JSON strings separates path and value
- surrounding whitespace is accepted on input
- canonical Go output does not add spaces around `=`
- no trailing semicolon
- JSON string escapes and values follow standard JSON
- numbers are decoded without unnecessary floating-point conversion
- generated output uses LF and readers accept LF or CRLF
- blank lines may separate assignments
- assignments are emitted in deterministic path order

The path language contains four segment forms:

```text
"property"  Object property segment
[0]         indexed replacement selector
[*]         anonymous atomic Union member
[@name]     named structured Union member
```

Property strings and right-hand values use JSON syntax. The path container,
array selectors, assignment delimiter, and whitespace rules above form the
current CTK JSON Flat boundary. An implementation may refine its parser and
diagnostics without making this optional representation a Core requirement.

Comments, Markdown headings, fenced blocks, and other review annotations are
not CTK JSON Flat Format. A Workbench renderer or reader is responsible for
separating presentation from Assignment content before calling the Flat Format
codec.

The current Go codec has round-trip, conflict, real CTK Settings, and Ingredient
split/composition coverage. Those tests provide evidence for that
implementation. They do not make the format a Bash requirement, a Core source
format, or the required representation for another implementation.
