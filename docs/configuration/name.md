---
title: "name"
---

# `name`

Name of a job. Will be printed in summary. If specified, the jobs can be merged with a jobs of the same name in a [local config](../examples/lefthook-local.md) or [extends](./extends.md).

### Example

```yml
# lefthook.yml

pre-commit:
  jobs:
    - name: lint and fix
      run: yarn run eslint --fix {staged_files}
```
