# Contributing

CTK develops public changes through a focused branch and Pull Request against
`main`. Keep the Pull Request title representative of the complete change and
record relevant validation in its description.

Pull Requests are squash-merged. The Pull Request title becomes the commit
title on `main`; intermediate branch commits remain useful for review without
becoming permanent mainline history.

Before opening a Pull Request, run the repository verification from the
checkout root:

```bash
go/verify.sh
```

GitHub Actions repeats that verification and runs the Go tests natively on
macOS and Windows. The workflow is read-only and does not publish Releases.

Documentation-only Issues and Pull Requests are welcome. In particular, when
a CTK document is difficult to discover from its identity, path, Node alias,
title, or headings, report the question you tried to resolve. Improving that
navigation metadata is a useful project change even when the implementation is
already correct.

CLI context and documentation-source diagnostics can include Workspace or
repository paths outside your home directory. Review diagnostic output before
pasting it into a public Issue. Paths below your home are shortened, but an
organization or project name in another path is not automatically hidden.
