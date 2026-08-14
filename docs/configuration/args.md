---
title: "args"
---

# `args`

::: callout tip New feature
Added in lefthook `2.0.5`
:::

Sometimes you want to pass arguments to the scripts or be able to overwrite arguments to the commands in `lefthook-local.yml`. For this you can use `args` option which will simply be appended to the command. You can use the same templates as in [`run`](./run.md).

Arguments passed by Git will be omitted if you specify `args` in the config. Providing no `args` or providing `args: "{0}"` works the same way.

See [`run`](./run.md) for supported templates.

#### Example

```yml
# lefthook.yml

pre-commit:
  jobs:
    - script: check-python-files.sh
      runner: bash
      args: "{staged_files}"
      glob: "*.py"

    - run: yarn lint
      args: "{staged_files}"
      glob:
        - "*.ts"
        - "*.js"
```
