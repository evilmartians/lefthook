## `skip`

You can skip all or specific commands and scripts using `skip` option. You can also skip when merging, rebasing, or being on a specific branch. Globs are available for branches.

Possible skip values:
- `rebase` - when in rebase git state
- `merge` - when in merge git state
- `merge-commit` - when current HEAD commit is the merge commit
- `ref: main` - when on a `main` branch
- `run: test ${SKIP_ME} -eq 1` - when `test ${SKIP_ME} -eq 1` is successful (return code is 0)

**Example**

Always skipping a command:

```yml
# lefthook.yml

pre-commit:
  commands:
    lint:
      skip: true
      run: yarn lint
```

Skipping on merging and rebasing:

```yml
# lefthook.yml

pre-commit:
  commands:
    lint:
      skip:
        - merge
        - rebase
      run: yarn lint
```

Or

```yml
# lefthook.yml

pre-commit:
  commands:
    lint:
      skip: merge
      run: yarn lint
```

Skipping when your are on a merge commit:

```yml
# lefthook.yml

pre-push:
  commands:
    lint:
      skip: merge-commit
      run: yarn lint
```

Skipping the whole hook on `main` branch:

```yml
# lefthook.yml

pre-commit:
  skip:
    - ref: main
  commands:
    lint:
      run: yarn lint
    test:
      run: yarn test
```

Skipping hook for all `dev/*` branches:

```yml
# lefthook.yml

pre-commit:
  skip:
    - ref: dev/*
  commands:
    lint:
      run: yarn lint
    test:
      run: yarn test
```

Skipping hook by running a command:

```yml
# lefthook.yml

pre-commit:
  skip:
    - run: test "${NO_HOOK}" -eq 1
  commands:
    lint:
      run: yarn lint
    test:
      run: yarn test
```

> TIP
>
> Always skipping is useful when you have a `lefthook-local.yml` config and you don't want to run some commands locally. So you just overwrite the `skip` option for them to be `true`.
>
> ```yml
> # lefthook.yml
>
> pre-commit:
>   commands:
>     lint:
>       run: yarn lint
> ```
>
> ```yml
> # lefthook-local.yml
>
> pre-commit:
>   commands:
>     lint:
>       skip: true
> ```
