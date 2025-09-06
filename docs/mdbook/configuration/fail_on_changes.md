# fail_on_changes

By default, lefthook exits with a non-zero status if any files (tracked by git) were modified by the hook if and only if a `CI` environment variable is set.

This behaviour can be overridden by setting `fail_on_changes` to `true` or `false`. In that case, lefthook will always, or never, exit with a non-zero status if any files (tracked by git) were modified by the hook.

```yml
pre-commit:
  fail_on_changes: true
  jobs:
    - run: yarn lint --fix
```
