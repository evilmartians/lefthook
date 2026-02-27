---
title: "Hook"
---

# Git hook

Contains settings for the git hook (commands, scripts, skip rules, etc.). You can specify any Git hook or your own custom, e.g. `test`


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
