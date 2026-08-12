# test.platform-validation.README.md
============================================================

# Platform Validation Cookbook

This directory contains a small, public Cookbook Source for targeted
real-machine validation of CTK's built-in VS Code ecosystem Platforms.

It is intentionally separate from both the user-editable repository Cookbook
and `go/internal/cookbook/testdata`:

- the repository Cookbook is product sample data;
- the Go fixture is stable input for Resolver unit tests;
- this Cookbook drives an installed application and CTK lifecycle operations.

## Coverage

The validation Recipe keeps only enough content to observe important Platform
boundaries:

- one default-scope Setting;
- one named Profile with a Profile-local Setting;
- one named-Profile Extension observed in every built-in Platform's primary
  Repository;
- exact Extension Pool refresh for Archive validation.

It does not model the author's daily environment or attempt broad editor
coverage.

## Isolated setup

Copy the static Cookbook Source into a disposable CTK Workspace. Do not copy an
existing `draft/` or `inspect/` Workbench.

```bash
# Keep the temporary Runtime path short. Some desktop products impose a
# substantially smaller internal IPC path limit than the filesystem does.
validation_home="$(mktemp -d /tmp/ctk-validation.XXXXXX)"
mkdir -p "$validation_home/cookbook"
cp -R test/platform-validation/cookbook/ingredient "$validation_home/cookbook/"
cp -R test/platform-validation/cookbook/recipe "$validation_home/cookbook/"

CTK_HOME="$validation_home" bin/ctk view recipe \
  "$validation_home/cookbook/recipe/vscode-platform-validation.macos.yaml"
```

Use the matching OS Recipe for Build, Apply, and Archive. A typical macOS pass
starts with:

```bash
CTK_HOME="$validation_home" bin/ctk build \
  "$validation_home/cookbook/recipe/vscode-platform-validation.macos.yaml"
```

The macOS and Windows Recipes cover all five built-in Platform identities. Use
the OS-specific Recipe matching the machine under observation. Recipe identity
includes the OS where both variants may be viewed in one Workspace.

Keep the Windows Workspace shallow as well, for example `C:\ctk-validation`.
Run the same Build, Apply, and Archive sequence from the shell used for CTK
validation. If a TLS-inspecting environment prevents a Platform Repository from
using its normal certificate chain, record the failure before retrying with the
already-observed process-local `NODE_OPTIONS=--use-system-ca`; do not weaken
Extension signature or certificate verification in the Recipe.

`activate`, `use`, and `deactivate` operate on the real Host paths declared by
the Platform. Run them only after confirming that another CTK Workspace is not
already managing those paths and that the Host state is recoverable.

## Scope

These files are scenario input, not automated proof that an installed product
works. Record real-machine results in the
[Go Platform Support Inventory](../../go/doc/platform-support.md), leaving
unexecuted operations explicit.
