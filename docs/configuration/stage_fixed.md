---
title: "stage_fixed"
---

# `stage_fixed`

**Default: `false`**

> Works **only for `pre-commit`** hook

When set to `true` lefthook will automatically call `git add` on files after running the command or script. For a command if [`files`](./files.md) option was specified, the specified command will be used to retrieve files for `git add`. For scripts and commands without [`files`](./files.md) option `{staged_files}` template will be used. All filters ([`glob`](./glob.md), [`exclude`](./exclude.md)) will be applied if specified.

**Example**

```yml
# lefthook.yml

pre-commit:
  commands:
    lint:
      run: npm run lint --fix {staged_files}
      stage_fixed: true
```

