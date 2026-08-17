---
type: concept
title: runner
source: "https://lefthook.dev/configuration/runner/"
path: /configuration/runner/
updated: 2026-08-17
okf:
  generated_by: "@docmd/plugin-okf"
  generated_at: "2026-08-17T13:11:15.262Z"
---
---
title: "runner"
---

# `runner`

You should specify a runner for the script. This is a command that should execute a script file. It will be called the following way: `<runner> <path-to-script>` (e.g. `ruby .lefthook/pre-commit/lint.rb`).

#### Example

```yml
# lefthook.yml

pre-commit:
  scripts:
    "lint.js":
      runner: node
    "check.go":
      runner: go run
```
