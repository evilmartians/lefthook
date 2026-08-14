---
title: "jobs"
---

# `jobs`

::: callout tip New feature
Added in lefthook `1.10.0`
:::

Jobs provide a flexible way to define tasks, supporting both commands and scripts. Jobs can be grouped for advanced flow control.

Named jobs are merged across [`extends`](./extends.md) and local config; unnamed jobs are appended in definition order. Groups can include other jobs with their own parallel or piped flow — `glob`, `root`, and `exclude` on a group apply to all nested jobs.

#### Example

::: callout info Note
Currently, only `root`, `glob`, and `exclude` options are applied to group jobs. Other options must be set for each job individually. Submit a [feature request](https://github.com/evilmartians/lefthook/issues/new?assignees=&labels=feature+request&projects=&template=feature_request.md) if this limits your workflow.
:::

A configuration demonstrating a piped group running in parallel with other jobs:

```yml
# lefthook.yml

pre-commit:
  parallel: true
  jobs:
    - name: migrate
      root: backend/
      glob: "db/migrations/*"
      group:
        piped: true
        jobs:
          - run: bundle install
          - run: rails db:migrate
    - run: yarn lint --fix {staged_files}
      root: frontend/
      stage_fixed: true
    - run: bundle exec rubocop
      root: backend/
    - run: golangci-lint
      root: proxy/
    - script: verify.sh
      runner: bash
```

This configuration runs migrate jobs in a piped flow while other jobs execute in parallel.
