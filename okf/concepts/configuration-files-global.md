---
type: concept
title: "files (hook-level)"
source: "https://lefthook.dev/configuration/files-global/"
path: /configuration/files-global/
updated: 2026-10-06
okf:
  generated_by: "@docmd/plugin-okf"
  generated_at: "2026-10-06T18:07:24.581Z"
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
