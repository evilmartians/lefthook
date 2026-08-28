---
type: concept
title: "files (job-level)"
source: "https://lefthook.dev/configuration/files/"
path: /configuration/files/
updated: 2026-08-28
okf:
  generated_by: "@docmd/plugin-okf"
  generated_at: "2026-08-28T10:12:38.520Z"
---
---
title: "files (job-level)"
---

# `files`

A custom command executed by the `sh` shell that returns the files or directories to be referenced in `{files}` template for [`run`](./run.md) setting.

If the result of this command is empty, the execution of commands will be skipped.

This option overwrites the [hook-level `files`](./files-global.md) option.

#### Example

Provide a git command to list files.

```yml
# lefthook.yml

pre-push:
  commands:
    stylelint:
      tags:
        - frontend
        - style
      files: git diff --name-only master
      glob: "*.js"
      run: yarn stylelint {files}
```

Call a custom script for listing files.

```yml
# lefthook.yml

pre-push:
  commands:
    rubocop:
      tags: backend
      glob: "**/*.rb"
      files: node ./lefthook-scripts/ls-files.js # you can call your own scripts
      run: bundle exec rubocop --force-exclusion --parallel -- {files}
```
