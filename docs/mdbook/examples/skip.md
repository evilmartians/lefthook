## Skip or run on condition

Here are two hooks.

`pre-commit` hook will only be executed when you're committing something on a branch starting with `dev/` prefix.

In `pre-push` hook:
- `test` command will be skipped if `NO_TEST` env variable is set to `1`
- `lint` command will only be executed if you're pushing the `main` branch

```yml
# lefthook.yml

pre-commit:
  only:
    - ref: dev/*
  commands:
    lint:
      run: yarn lint {staged_files} --fix
      glob: "*.{ts,js}"
    test:
      run: yarn test

pre-push:
  commands:
    test:
      run: yarn test
      skip:
        - run: test "$NO_TEST" -eq 1
    lint:
      run: yarn lint
      only:
        - ref: main
```
