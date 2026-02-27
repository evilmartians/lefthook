---
title: "group"
---

# `group`

You can define a group of jobs and configure how they should execute using the following options:

- [`parallel`](./parallel.md): Executes all jobs in the group simultaneously.
- [`piped`](./piped.md): Executes jobs sequentially, passing output between them.
- [`jobs`](./jobs.md): Specifies the jobs within the group.

### Example

```yml
# lefthook.yml

pre-commit:
  jobs:
    - group:
        parallel: true
        jobs:
          - run: echo 1
          - run: echo 2
          - run: echo 3
```

If you specify `env`, `root`, `glob`, or `exclude` on a group, they will be inherited to the underlying jobs.

```yml
# lefthook.yml

pre-commit:
  jobs:
    - env:
        E1: hello
      glob:
        - "*.md"
      exclude:
        - "README.md"
      root: "subdir/"
      group:
        parallel: true
        jobs:
          - run: echo $E1
          - run: echo $E1
            env:
              E1: bonjour
```

> **Note:** To make a group mergeable with settings defined in local config or extends you have to specify the name of the job group belongs to:
> ```yml
> pre-commit:
>   jobs:
>     - name: a name of a group
>       group:
>         jobs:
>           - name: lint
>             run: yarn lint
>           - name: test
>             run: yarn test
> ```
