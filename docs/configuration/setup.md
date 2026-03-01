---
title: 'setup'
---

# `setup`

::: callout tip New feature
Added in lefthook `2.1.2`
:::

A list of instructions to run before any job. Supports templates and Git args like in [`run`](./run.md).

::: callout info Note
When merging configs (with `lefthook-local.yml` or files from [`extends`](./extends.md)) `setup` instructions get **prepended**. When there are multiple `extends`, they get **appended** in the same order as extend files are specified.
:::

#### Example

```yml
# lefthook.yml

pre-commit:
  setup:
    - run: |
        if ! command -v golangci-lint >/dev/null 2>&1; then
          go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.10.1
        fi
  jobs:
    - run: golangci-lint {staged_files}
      glob: "*.go"
```
