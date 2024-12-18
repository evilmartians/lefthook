## Stage fixed files

> Works only for `pre-commit` Git hook

Sometimes your linter fixes the changes and you usually want to commit them automatically. To enable auto-staging of the fixed files use [`stage_fixed`](../configuration/stage_fixed.md) option.

```yml
# lefthook.yml

pre-commit:
  commands:
    lint:
      run: yarn lint {staged_files} --fix
      stage_fixed: true
```
