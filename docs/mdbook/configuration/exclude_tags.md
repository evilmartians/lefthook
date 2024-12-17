### `exclude_tags`

[Tags](./tags.md) or command names that you want to exclude. This option can be overwritten with `LEFTHOOK_EXCLUDE` env variable.

**Example**

```yml
# lefthook.yml

pre-commit:
  exclude_tags: frontend
  commands:
    lint:
      tags: frontend
      ...
    test:
      tags: frontend
      ...
    check-syntax:
      tags: documentation
```

```bash
lefthook run pre-commit # will only run check-syntax command
```

**Notes**

This option is good to specify in `lefthook-local.yml` when you want to skip some execution locally.

```yml
# lefthook.yml

pre-push:
  commands:
    packages-audit:
      tags:
        - frontend
        - security
      run: yarn audit
    gems-audit:
      tags:
        - backend
        - security
      run: bundle audit
```

You can skip commands by tags:

```yml
# lefthook-local.yml

pre-push:
  exclude_tags:
    - frontend
```
