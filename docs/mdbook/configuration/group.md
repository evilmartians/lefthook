## `group`

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
