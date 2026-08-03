# Cookbook test fixture

This Cookbook is stable test input for the Go Resolver. Tests must not use the
workspace-level `cookbook/`, because it is user-editable product data and may
change independently of Go implementation contracts.

Keep this fixture minimal. Add a Recipe or Ingredient only when a test needs a
concrete cross-file resolution pattern; focused cases should continue to use
temporary directories created by the test.
