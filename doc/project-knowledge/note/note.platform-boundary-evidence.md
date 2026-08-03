# project-knowledge.note.platform-boundary-evidence.md
============================================================

# Platform Differences as Boundary Evidence

Several Windows and Git Bash incidents during CTK development initially looked
like broad design problems. Inspecting the actual files, command output, and
host representation usually revealed a smaller boundary mismatch.

This Note records a loose development observation. It is not a Windows support
matrix or a required debugging method.

## Observation

Platform behavior should be treated as evidence before it becomes a shared
rule.

```text
Unexpected platform behavior
    ↓
Inspect the real file, command output, process, or link
    ↓
Identify the boundary that changed representation
    ↓
Normalize or adapt at that boundary
    ↓
Change shared responsibility only if the evidence requires it
```

This avoids turning one shell, filesystem representation, or line-ending issue
into a product-wide architecture decision.

## Examples observed in CTK

### CRLF at ingestion boundaries

External Platform output and editable Draft input may carry CRLF on Windows.
The useful correction was to normalize the data where it entered the parser or
lookup boundary, not to make every later operation understand stray carriage
returns.

### Junctions and symbolic links

A Windows CodeVenv Selection may be represented by a Junction. Checking only a
Unix-style symlink mode made a valid managed Selection appear absent.

The reusable responsibility was "resolve a managed link to its target," not
"accept only the filesystem representation reported as a symbolic link."

### MSYS argument conversion

Git Bash helps the Bash reference implementation use Unix-oriented tools on
Windows, but MSYS may transform paths and arguments when invoking native
Windows commands.

The shell boundary is therefore part of the evidence. A value printed before a
native command call is not always proof of the value received by that command.

### Binary and observed state can drift

During local development, source, release artifacts, links, and the binary
currently on `PATH` may represent different revisions. Verify which executable
ran and inspect the actual host state before changing a Contract to explain an
apparent platform failure.

## Loose guidance

- Observe the host representation before generalizing it.
- Normalize external data at the boundary that introduces it.
- Model the capability CTK needs rather than one operating system's mechanism.
- Keep shell- or language-specific requirements in that implementation's
  declaration.
- Prefer a bounded adapter fix when shared responsibility has not changed.

## Boundary

These observations do not imply that every adapter must hide every platform
difference. Safety, compatibility, or user-visible behavior may still require a
Contract change.

The point is to establish that need from evidence rather than infer it from the
first platform-specific symptom.

## Related documents

- `../../../go/README.md` — primary Go implementation declaration
- `../../../bash/scripts/README.md` — Bash implementation declaration
- `../../contract/contract.code-environment.md` — current host-integration
  safety boundary
