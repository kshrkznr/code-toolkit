# project-knowledge.design-note.concept-domain.md
============================================================

# Why Concept Domains Emerged

## Background

As CTK evolved, the number of independent concepts continued to grow.

Core concepts such as Cookbook and Build Lifecycle were joined by supporting concepts including Draft, Inspect, Documentation Resolver, and Code Environment.

Although these concepts were related, they represented different areas of responsibility.

Introducing a higher-level grouping made the overall architecture easier to understand without changing the concepts themselves.

---

## Design

Concept Domains group related Concept APIs by responsibility.

They provide a high-level view of the product while keeping each Concept API independently documented.

Concept Domains are intended for navigation and understanding rather than introducing additional runtime behavior.

## Note

Concept Domains were introduced after the Concept APIs rather than before.

The domains emerged by observing how the existing Concept APIs naturally clustered, instead of being designed as a top-down architecture.

## Boundary

This Design Note explains why the grouping emerged. Current Concept Domain
responsibilities remain in their accepted Node READMEs and the authoritative
Repository Map.
