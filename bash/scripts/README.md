# CTK Bash Implementation

This directory contains the Bash reference implementation of CTK.

This README is the documentation boundary between shared CTK documents and the
Bash source. Readers resolving Bash behavior should read its declared Kitchen
Notes and implementation guidance here before consulting scripts as behavioral
evidence.

## Windows / Git Bash execution boundary

The Bash reference implementation uses Git Bash / MSYS when it runs on
Windows. This is a requirement of this implementation, not a shared CTK Windows
requirement. The primary Go Windows binary does not require Git Bash.

Current Bash implementation details include:

- `/usr/bin` is prepended to `PATH` under MINGW/MSYS so the scripts resolve the
  expected Unix tools.
- Windows VSIX downloads may pass curl's `--ssl-no-revoke` option.
- Platform process discovery and termination delegate to `powershell.exe`.
- Calls from MSYS to native Windows commands cross MSYS argument and path
  conversion; new integrations should inspect the actual arguments received by
  the native command.

These statements document the preserved Bash environment. They are not
promises that another shell can run the Bash implementation, nor requirements
for other CTK implementations.

## Applied Kitchen Notes

The Bash implementation currently adopts the Merge Rules Kitchen Note for
Settings composition:

- Settings Resources enter the merge stream in Cookbook Core resolution order.
- Objects merge recursively through `jq` object multiplication.
- Arrays, scalars, and `null` use the later value.
- Array composition is `replace`; `union` Merge Rules are not implemented.

This section is the Bash implementation's declaration of applied Kitchen
Notes. Notes not declared here are not part of Bash Cookbook interpretation.

The historical Bash implementation does not adopt the [Extension Set Kitchen
Note](../../doc/note/note.extension-set.md). It provides no behavior guarantee
for Extension Set declarations. Because Kitchen Notes are optional and Bash is
retained as reference evidence, Bash does not need Go's strict `set:`
reserved-prefix guard.
