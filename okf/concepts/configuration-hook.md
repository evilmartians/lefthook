---
type: concept
title: Hook
source: "https://lefthook.dev/configuration/Hook/"
path: /configuration/Hook/
updated: 2026-08-17
okf:
  generated_by: "@docmd/plugin-okf"
  generated_at: "2026-08-17T13:11:15.241Z"
---
---
title: "Hook"
---

# Git hook

Contains settings for the git hook (commands, scripts, skip rules, etc.). You can specify any Git hook or your own custom, e.g. `test`

#### Example

```yml
# lefthook.yml

# Git hook
pre-commit:
  jobs:
    - run: yarn lint {staged_files} --fix
      stage_fixed: true

# Custom hook
check-docs:
  jobs:
    - run: yarn check-docs
    - run: typos
```
