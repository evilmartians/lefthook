## `group`

Specifies a group of jobs and option to run them with.

- [`parallel`](./parallel.md)
- [`piped`](./piped.md)
- [`jobs`](./jobs.md)

### Example

```yml
# lefthook.yml

pre-commit:
  jobs:
    - group:
        parallel: true
        jobs:
          - run: echo hello from a group
```

> **Note:** To make a group mergeable with settings defined in local config or extends you have to specify the name of the job group belongs to:
> ```yml
> pre-commit:
>   jobs:
>     - name: a name of a group
>       group:
>         jobs:
>           - run: echo from a group job
> ```
