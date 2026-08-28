---
type: concept
title: script
source: "https://lefthook.dev/configuration/script/"
path: /configuration/script/
updated: 2026-08-28
okf:
  generated_by: "@docmd/plugin-okf"
  generated_at: "2026-08-28T10:12:38.535Z"
---
---
title: "script"
---

# `script`

Name of a script to execute. The rules are the same as for [`scripts`](./Scripts.md)

Use [`args`](./args.md) to append arguments to the script. Configuring `args`
replaces arguments passed by Git unless the `{0}` template is included.

#### Example

```yml
# lefthook.yml

pre-commit:
  jobs:
    - script: linter.sh
      runner: bash
      args: "{staged_files}"
```

```bash
# .lefthook/pre-commit/linter.sh

echo "Everything is OK"
```
