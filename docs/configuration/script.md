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
