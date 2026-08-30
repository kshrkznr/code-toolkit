# Knowledge.integration.workspace.md
============================================================

# Concept API: Workspace

## Concept Domain

Integration

## Why

A Cookbook may be maintained in its own versioned repository while generated
Distributions and review output remain local to the machine using it. Treating
every path as one `cookbook/` directory makes that separation accidental: a
Draft can enter Source, or generated state can become coupled to a Source
checkout.

Workspace makes the ownership boundary explicit without introducing a global
CTK profile or requiring Source to move.

## Definition

A CTK Workspace is the selected integration context for one invocation. It
connects Cookbook Source to the mutable and generated state used while CTK
builds, observes, reviews, archives, and integrates Runtimes.

```text
selected CTK Workspace
    ├── selects Cookbook Source
    ├── owns Workbench review state
    ├── selects the Dist root
    └── owns Archive and Extension Pool state
```

Workspace is distinct from a project-owned `.code-workspace` file and from a
future Workspace Build target.

## Responsibility

Workspace is responsible for:

- identifying the integration context selected for one CTK invocation;
- resolving static Recipe, Ingredient, and Kitchen Note Source;
- keeping Draft and Inspect output in Workspace-local review state;
- resolving the Dist root used by lifecycle and CodeVenv operations;
- owning Archive and Extension Pool state unless a later observed requirement
  establishes another boundary;
- providing a Workspace-local ownership boundary for integration
  configuration.

Cookbook Source and the Dist root may be located outside the Workspace root.
Workbench output remains below the Workspace so generated review state does not
enter independently versioned Cookbook Source accidentally.

This responsibility establishes where machine integration configuration
belongs. Loading external Platform definitions from that boundary remains an
unimplemented Future.

## Selection and defaults

One invocation selects one Workspace. An explicit `CTK_HOME` may select it;
implementations may also provide deterministic local discovery.

### Where Workspace state is stored

Absent location overrides, the Workspace resolves the familiar local layout:

```text
CTK_HOME/
├── .config/
├── cookbook/
│   ├── recipe/
│   ├── ingredient/
│   ├── draft/
│   └── inspect/
├── dist/
├── archive/
└── .vsix/
```

Changing a selected location changes subsequent resolution only. Workspace
does not imply migration, an old-location registry, persistence of `CTK_HOME`,
or a user-global active-Workspace selector.

## Ownership boundary: Dist, Archive, and Workbench

Package managers may install the CTK binary and compiled Built-in Platform
definitions. They do not own or update Workspace configuration, Cookbook
Source, Dist, Archive, Pool, or Workbench state.

Recipes select portable desired Runtime content and a Platform identity. They
do not own machine integration paths. Kitchen Notes remain Cookbook Source and
may guide Cookbook interpretation; they do not select Workspace locations.

## See also

- [Why Platform Definitions Belong to Workspace Integration](../design-note/design-note.platform-definition-scope.md)
- [Freeze and Inspect Workbench Contract](../contract/contract.workbench.md)
- [Future: Workspace-loaded Platform definitions](../future/future.platform-registry.md)
