## Local config

You can extend and override options of your main configuration with `lefthook-local.yml`. Don't forge to add the file to `.gitignore`.

You can also use `lefthook-local.yml` without a main config file. This is useful when you want to use lefthook locally without imposing it on your teammates.

```yml
# lefthook.yml (committed into your repo)

pre-commit:
  jobs:
    - name: linter
      run: yarn lint
    - name: tests
      run: yarn test
```

```yml
# lefthook-local.yml (ignored by git)

pre-commit:
  jobs:
    - name: tests
      skip: true # don't want to run tests on every commit
    - name: linter
      run: yarn lint {staged_files} # lint only staged files
```
