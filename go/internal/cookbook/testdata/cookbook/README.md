# Cookbook test fixture

This Cookbook is stable test input for the Go Resolver. Tests must not use the
workspace-level `cookbook/`, because it is user-editable product data and may
change independently of Go implementation contracts.

Keep this fixture minimal. Add a Recipe or Ingredient only when a test needs a
concrete cross-file resolution pattern; focused cases should continue to use
temporary directories created by the test.

The `vscode-golang` macOS and Windows Recipes are representative Extension Set
fixtures. Runtime `golang` and Profile `core` reuse `set:editor-core`; Profile
`core` also declares the same member directly to verify final deduplication.
Profile `inspect` references a Set with no `.extensions` Resource, while
Profile `ops` references a present empty Resource. The Set member also owns an
Extension Settings Resource so the fixture exercises the existing concrete
Extension resolution path.
