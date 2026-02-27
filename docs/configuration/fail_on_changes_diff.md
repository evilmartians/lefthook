---
title: "fail_on_changes_diff"
---

# `fail_on_changes_diff`

When Lefthook exits with a non-zero status as a result of [`fail_on_changes`](./fail_on_changes.md) triggering,
it can optionally output a diff of the detected changes.

The default behavior is to output the diff when run in a CI pipeline.
The `fail_on_changes_diff` boolean configuration parameter can be used to override this.

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
