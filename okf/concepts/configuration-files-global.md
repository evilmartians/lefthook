---
type: concept
title: "files (hook-level)"
source: "https://lefthook.dev/configuration/files-global/"
path: /configuration/files-global/
updated: 2026-08-03
okf:
  generated_by: "@docmd/plugin-okf"
  generated_at: "2026-08-03T09:38:01.696Z"
---
---
title: "files (hook-level)"
---

# `files`

A custom command executed by the `sh` shell that returns the files or directories to be referenced in `{files}` template. See [`run`](./run.md) and [`files`](./files.md).

If the result of this command is empty, the execution of commands will be skipped.

#### Example

```yml
# lefthook.yml

pre-commit:
  files: git diff --name-only master # custom list of files
  commands:
    ...
```
