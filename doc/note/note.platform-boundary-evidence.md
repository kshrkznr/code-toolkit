# Knowledge.note.platform-boundary-evidence.md
============================================================

# Platform Differences as Boundary Evidence

Several Windows and Git Bash incidents initially looked like broad CTK design
problems. Inspecting the actual files, command output, processes, and Host
representation usually revealed a smaller integration-boundary mismatch.

This Note preserves operational guidance for implementing and diagnosing CTK
Platform integration. It is not a Windows support matrix or a required
debugging method.

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

This avoids turning one shell, filesystem representation, process shape, or
line-ending issue into a product-wide architecture decision.

## Examples observed in CTK

### Application identities and owned paths

The current VS Code-family integrations began from application-owned
identities rather than CTK naming conventions.

| Application | Platform command | Host data identity | Extension root | Observed primary Repository |
| --- | --- | --- | --- | --- |
| Visual Studio Code | `code` | `Code` | `.vscode/extensions` | Visual Studio Marketplace |
| VSCodium | `codium` | `VSCodium` | `.vscode-oss/extensions` | Open VSX |
| Kiro | `kiro` | `Kiro` | `.kiro/extensions` | Open VSX |
| Cursor | `cursor` | `Cursor` | `.cursor/extensions` | Cursor Marketplace |
| Devin Desktop | `devin-desktop` | `Devin` | `.devin/extensions` | Windsurf Marketplace |

On macOS, the Host data identity resolves below
`~/Library/Application Support`; on Windows it resolves below the roaming
application data directory. Extension roots resolve from the user's home
directory.

These are retained application observations, not names invented to make the
Platforms uniform. A product update does not require routine re-observation,
but packaging, storage, command, or Repository changes are reasons to revisit
the affected boundary.

The primary Repository column records the Platform-owned installation boundary.
It does not include CTK's secondary Pool candidates or CTK-owned exact-artifact
acquisition policy. Those belong to
[Extension Resolution](note.extension-resolve.md) and the current implementation
inventory.

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

### Process identity and ancestry

VS Code-family applications may use the same executable identity for their
desktop root, helpers, language servers, and other workers. Some helpers expose
`--type=...`, but the absence of that argument is not sufficient evidence that a
process is the desktop root.

Cursor on Windows made this boundary explicit: workers can use `Cursor.exe`
without `--type`, while the intended desktop root can be distinguished through
same-name process ancestry. The reusable observation is to inspect the process
tree and Runtime paths rather than infer ownership from an application name or
one argument alone.

The current Go implementation retains its concrete process strategies in the
[Go Platform Support Inventory](../../go/doc/platform-support.md). The shared
Code Environment Contract remains responsible for the requirement to stop only
processes that can access the Runtime paths being replaced.

### MSYS argument conversion

Git Bash helps the Bash reference implementation use Unix-oriented tools on
Windows, but MSYS may transform paths and arguments when invoking native
Windows commands.

The shell boundary is therefore part of the evidence. A value printed before a
native command call is not always proof of the value received by that command.

### Runtime path length and application IPC

A filesystem accepting a Runtime path does not prove that the Platform can use
the same path for every internal resource. Desktop applications may derive IPC,
lock, or other internal paths below the supplied User Data directory and impose
a substantially smaller limit than the filesystem.

This was observed while building a fresh isolated Devin Desktop Runtime. The
same Recipe, Profile, and binary succeeded with a shallow Runtime path and
failed to persist the Profile with a deeper path. Shortening the Distribution
name also succeeded because the name had previously been copied into CTK's
staging directory name; Recipe weight, Profile ownership, and the temporary
filesystem itself were not the deciding differences.

CTK should avoid adding unnecessary length to application-facing paths. The Go
Build lifecycle therefore uses short internal staging names that do not repeat
the Distribution identity. This cannot remove limits introduced by an
arbitrarily deep `CTK_HOME` or by the application itself. When diagnosing a
fresh Runtime that exits or does not persist a Profile without a useful error,
repeat the observation from a shallow Workspace before changing Platform or
Recipe semantics. Keep disposable Platform-validation roots short for the same
reason.

No portable numeric maximum is declared here. Applications may append different
internal suffixes, and their limits may change independently of the enclosing
filesystem.

### Binary and observed state can drift

During local development, source, release artifacts, links, and the binary
currently on `PATH` may represent different revisions. Verify which executable
ran and inspect the actual Host state before changing a Contract to explain an
apparent Platform failure.

## Loose guidance

- Observe the Host representation before generalizing it.
- Normalize external data at the boundary that introduces it.
- Model the capability CTK needs rather than one operating system's mechanism.
- Keep shell- or language-specific requirements in that implementation's
  declaration.
- Prefer a bounded Adapter fix when shared responsibility has not changed.

## Boundary

These observations do not imply that every Adapter must hide every Platform
difference. Safety, compatibility, or user-visible behavior may still require a
Contract change.

The point is to establish that need from evidence rather than infer it from the
first Platform-specific symptom.

## Related documents

- [Go implementation](../../go/README.md)
- [Bash reference implementation](../../bash/scripts/README.md)
- [Code Environment Integration Contract](../contract/contract.code-environment.md)
- [VS Code Ecosystem Platform Intake](../project-knowledge/note/note.vscode-ecosystem-platform-intake.md)
