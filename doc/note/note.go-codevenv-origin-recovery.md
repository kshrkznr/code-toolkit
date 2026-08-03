# Knowledge.note.go-codevenv-origin-recovery.md
============================================================

# Go CodeVenv Origin Recovery

This Note records a specialized manual-maintenance procedure for the current Go
implementation. It is not a portable CodeVenv Contract or the normal
deactivation workflow.

Use it only when a completed Distribution already represents the exact host
state that should be restored while stopping CodeVenv management.

## Safety boundary

This procedure does not make arbitrary managed-state edits recoverable.

Do not manually remove or rewrite `current.<platform>`, transaction journals,
host connections, or trusted Lock files. If those have already been changed,
the remaining state may be ambiguous and CTK may correctly refuse recovery.

Prefer normal `ctk deactivate <platform>`. Use
`ctk deactivate <platform> --force` only when CTK reports that trusted recovery
is available. `--force-empty` is a separate destructive last resort; it does not
restore the imported Runtime.

## Required source state

The source must be a completed Distribution for the same Platform with a
complete trusted Go Lock, including:

- `.lock/recipe.yaml`
- `.lock/manifest.json`
- every observation required by that Lock

The Go deactivation path identifies origin primarily through the reserved
physical path `dist/origin.<platform>`. Changing `.meta/recipe.yaml` does not
select a deactivation origin, and renaming a Distribution alone does not make
it trustworthy.

Do not edit `.lock/recipe.yaml` or `.lock/manifest.json` independently. Their
provenance and contents must remain consistent.

## Cautious procedure

1. Prefer expressing the desired state as a Recipe and building a normal
   Distribution. Use direct adoption only when an existing Distribution is
   already the intended recovery source.
2. Stop the target Platform and confirm that the source Distribution belongs to
   the same Platform.
3. Refresh its trusted Lock when needed with `ctk lock <dist>` while it remains
   at its original path.
4. Preserve the existing `dist/origin.<platform>` under a non-reserved backup
   name.
5. Copy the complete source Distribution to `dist/origin.<platform>`.
6. Run `ctk deactivate <platform>` normally without launching, applying, or
   relocking the copied origin.

Prefer a copy over moving the source. Moving the target of
`current.<platform>` would break the active Selection before deactivation can
manage the transition.

## Treat the copy as recovery input

Do not launch, apply, or relock the copied origin. Extension state may contain
location-sensitive identity or fingerprint data. Normal Go deactivation
restores User Data and reconstructs Extensions from the trusted Lock at their
final host path; it does not adopt the copied `.ext` directory directly.

The reserved origin may be replaced by a later activation. Keep long-term
backups outside the reserved name.

If no valid trusted origin can be prepared and an empty host state is
acceptable, `ctk deactivate <platform> --force-empty` is the explicit
destructive alternative.

## Related documents

- `note.codevenv.md` — normal operational observations
- `../integration/integration.code-venv.md` — current lifecycle
- `../contract/contract.code-environment.md` — required safety boundary
