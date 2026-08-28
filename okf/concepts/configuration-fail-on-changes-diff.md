---
type: concept
title: fail_on_changes_diff
source: "https://lefthook.dev/configuration/fail_on_changes_diff/"
path: /configuration/fail_on_changes_diff/
updated: 2026-08-28
okf:
  generated_by: "@docmd/plugin-okf"
  generated_at: "2026-08-28T10:12:38.518Z"
---
---
title: "fail_on_changes_diff"
---

# `fail_on_changes_diff`

**Default:** outputs diff only in CI

When [`fail_on_changes`](./fail_on_changes.md) triggers, lefthook can optionally print a diff of the detected changes. Set this boolean to explicitly enable or disable the diff output regardless of environment.

#### Example

```yml
# lefthook.yml
pre-commit:
  parallel: true
  fail_on_changes: "always"
  fail_on_changes_diff: true
  commands:
    lint:
      run: yarn lint
    test:
      run: yarn test
```
