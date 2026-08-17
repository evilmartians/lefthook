---
type: concept
title: "Local config"
source: "https://lefthook.dev/usage/features/local/"
path: /usage/features/local/
updated: 2026-08-17
okf:
  generated_by: "@docmd/plugin-okf"
  generated_at: "2026-08-17T13:11:15.287Z"
---
---
title: "Local config"
---

# Local config

You can extend and override options of your main configuration with `lefthook-local.yml`. Don't forget to add the file to `.gitignore`.

You can also use `lefthook-local.yml` without a main config file. This is useful when you want to use lefthook locally without imposing it on your teammates.

```yml
# lefthook.yml (committed into your repo)

pre-commit:
  jobs:
    - name: linter
      run: yarn lint
    - name: tests
      run: yarn test
```

```yml
# lefthook-local.yml (ignored by git)

pre-commit:
  jobs:
    - name: tests
      skip: true # don't want to run tests on every commit
    - name: linter
      run: yarn lint {staged_files} # lint only staged files
```
