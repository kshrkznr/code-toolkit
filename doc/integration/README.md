# Knowledge.integration.md
============================================================

# Concept Domain: Integration

Integration connects CTK's accepted concepts to the environments and
information structures through which they are used.

It makes Core and Workbench responsibilities accessible through documentation,
repository navigation, Platform commands, and Runtime selection without
redefining those responsibilities.

## Responsibility

Integration is responsible for mapping CTK concepts onto an external or
repository-facing context.

An Integration Concept API should make both sides of that mapping visible:

```text
CTK responsibility
        ↓ integrated through
Documentation, repository, or development environment
```

Integration may define the boundary and lifecycle of that connection. It does
not move implementation-specific representation into Core or make one external
environment part of every CTK implementation.

## Navigate by question

Start with the question closest to the current work:

- How do I find the document responsible for the current question?
  → [Documentation Resolver](../README.md)
- How do CTK concepts and document roles map to repository paths, Canonical
  identities, and headings?
  → [Project Structure](integration.project-structure.md)
- How does CTK select, launch, and restore IDE Runtimes through Platform
  commands?
  → [Code Environment](integration.code-venv.md)
- Where is everything stored in one selected CTK Workspace, and how can
  Cookbook Source remain separate from generated review state?
  → [Workspace](integration.workspace.md)

These routes identify the current Integration Concept APIs without duplicating
the authoritative Repository Map maintained by the Documentation Resolver.

## Boundary

Integration does not define:

- Cookbook, Build, or Persistence responsibilities owned by Core
- Draft or Inspect responsibilities owned by Workbench
- one canonical repository layout for every project or Cookbook
- one Platform representation that every implementation must use
- implementation behavior already owned by a language README or Contract

Return to the [Documentation Resolver](../README.md) when the current
question belongs to another Concept Domain or supporting document role.

## Shared documentation guidance

An Integration document should name the CTK responsibility being exposed, the
external context it connects to, and the boundary between them. Definition,
responsibility, lifecycle, representation, and boundary headings are useful
when they make that mapping easier to inspect.

The directory README provides entry routes only. Detailed Concept API documents
remain responsible for their own behavior and terminology.
