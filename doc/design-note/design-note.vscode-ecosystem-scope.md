# Knowledge.design-note.vscode-ecosystem-scope.md
============================================================

# Why CTK Keeps Platform Inside the VS Code Ecosystem

CTK uses *Platform* to distinguish VS Code-family applications such as Visual
Studio Code, VSCodium, Kiro, Cursor, and Devin Desktop. Repeated Platform intake
made it tempting to turn the shared Runtime implementation into a selectable
`Runtime Adapter`, leaving names such as `jetbrains` or `eclipse` open for the
future.

That abstraction does not match the current product boundary. CTK manages a VS
Code ecosystem Runtime: User Data, Extensions, Profiles, VSIX artifacts, and the
CLI behavior that connects them. Another IDE ecosystem would change that model,
not merely supply another Platform definition.

## A Platform definition describes application differences

Within the VS Code ecosystem, observed differences can be declared or resolved
through bounded strategies:

- Platform identity and command
- Host User Data and Extension paths
- process identity
- Extension Pool Repository order and download capability

Settings and Extension operations, Profile semantics, isolated launch, and
Runtime observation remain CTK's VS Code ecosystem behavior. A Platform file
therefore does not need `adapter: vscode`, because there is no alternative
adapter kind to select.

```text
VS Code ecosystem Runtime model
    ├── code Platform definition
    ├── codium Platform definition
    ├── kiro Platform definition
    ├── cursor Platform definition
    └── devin-desktop Platform definition
```

JetBrains or Eclipse support would require a separate product-scope decision
and concrete observation of its Runtime model. CTK does not reserve a Platform
configuration field in anticipation of that decision.

## Launch Override permits launcher-only interoperation

A Launch Override can replace how a Distribution is started. It can invoke
Eclipse, a JetBrains IDE, or another application, so CTK can be used as a plain
launcher outside the VS Code ecosystem.

That executable target does not become a CTK Platform. The override does not
teach CTK how to Build, Apply, observe, Lock, Archive, or recover the target
application's Runtime.

```text
Launch Override
    └── may start any caller-defined application

Platform integration
    └── owns the complete VS Code ecosystem Runtime lifecycle
```

Using CTK only as a launcher for another application is therefore supported by
the existing override boundary and is not a reason to add a Runtime Adapter
selector.

## Current decision

- Define CTK Platform support within the VS Code ecosystem.
- Do not expose a Runtime Adapter kind in built-in or external Platform
  definitions.
- Keep shared Runtime behavior implemented by CTK rather than configurable per
  Platform.
- Treat support for another IDE ecosystem as a new product decision, not an
  unimplemented Platform strategy name.
- Permit Launch Override to start applications outside the VS Code ecosystem
  while keeping that interoperation launch-only.

## Boundary

The [Code Environment Integration Contract](../contract/contract.code-environment.md)
defines the accepted Platform scope and lifecycle capabilities. The
[Platform Registry Future](../future/future.platform-registry.md) considers how
VS Code ecosystem application differences may become declarative. The
[Platform Definition Scope Design Note](design-note.platform-definition-scope.md)
explains why external definitions belong to Workspace integration rather than
Cookbook interpretation. The
[Go Distribution Contract](../../go/doc/contract/contract.distribution.md)
defines the current Launch Override behavior.
