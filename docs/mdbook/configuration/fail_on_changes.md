# fail_on_changes

The behaviour of lefthook when files (tracked by git) are modified can set by modifying the `fail_on_changes` configuration parameter. The possible values are:

- `never`: never exit with a non-zero status if files were modified (default).
- `always`: always exit with a non-zero status if files were modified.
- `ci`: exit with a non-zero status only when the `CI` environment variable is set. This can be useful when combined with `stage_fixed` to ensure a frictionless devX locally, and a robust CI.
- `non-ci`: exit with a non-zero status only when the `CI` environment variable is _not_ set. This can be useful in setups where the CI pipeline commits changes automatically, such as [autofix.ci](https://autofix.ci).

```yml
# lefthook.yml
pre-commit:
  parallel: true
  fail_on_changes: "always"
  commands:
    lint:
      run: yarn lint
    test:
      run: yarn test
```
