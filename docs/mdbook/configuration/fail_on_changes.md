# fail_on_changes

Exit with a non-zero status if any files were modified by the hook.
This can be useful for hooks that automatically fix code.

Defaults to `true` if `CI` environment variable is `true`, otherwise `false`.

Can be used in a hook definition.

```yaml
pre-commit:
  fail_on_changes: true
  jobs:
    - run: yarn lint --fix
