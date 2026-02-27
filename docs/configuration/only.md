---
title: "only"
---

# `only`

You can force a command, script, or the whole hook to execute only in certain conditions. This option acts like the opposite of [`skip`](./skip.md). It accepts the same values but skips execution only if the condition is not satisfied.

> **Note:** `skip` option takes precedence over `only` option, so if you have conflicting conditions the execution will be skipped.

**Example**

Execute a hook only for `dev/*` branches.

```yml
# lefthook.yml

pre-commit:
  only:
    - ref: dev/*
  commands:
    lint:
      run: yarn lint
    test:
      run: yarn test
```

When rebasing execute quick linter but skip usual linter and tests.

```yml
# lefthook.yml

pre-commit:
  commands:
    lint:
      skip: rebase
      run: yarn lint
    test:
      skip: rebase
      run: yarn test
    lint-on-rebase:
      only: rebase
      run: yarn lint-quickly
```
