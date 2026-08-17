---
type: concept
title: fail_text
source: "https://lefthook.dev/configuration/fail_text/"
path: /configuration/fail_text/
updated: 2026-08-17
okf:
  generated_by: "@docmd/plugin-okf"
  generated_at: "2026-08-17T13:12:03.061Z"
---
---
title: "fail_text"
---

# `fail_text`

You can specify a text to show when the command or script fails.

#### Example

```yml
# lefthook.yml

pre-commit:
  commands:
    lint:
      run: yarn lint
      fail_text: Add node executable to $PATH
```

```bash
$ git commit -m 'fix: Some bug'

Lefthook v1.1.3
RUNNING HOOK: pre-commit

  EXECUTE > lint

SUMMARY: (done in 0.01 seconds)
🥊  lint: Add node executable to $PATH env
```
