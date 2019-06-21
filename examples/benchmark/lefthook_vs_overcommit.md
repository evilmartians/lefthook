Benchmark based on [discourse](https://github.com/discourse/discourse/blob/master/.travis.yml#L77-L83) project.
We take all commands from Travis CI "as is".

## Lefthook

```yml
pre-commit:
  parallel: true
  commands:
    rubocop:
      run: bundle exec rubocop --parallel
    danger:
      run: bundle exec danger
    eslint-assets:
      run: npx yarn eslint --ext .es6 app/assets/javascripts
    eslint-test:
      run: npx yarn eslint --ext .es6 test/javascripts
    eslint-plugins-assets:
      run: npx yarn eslint --ext .es6 plugins/**/assets/javascripts
    eslint-plugins-test:
      run: npx yarn eslint --ext .es6 plugins/**/test/javascripts
    eslint-assets-tests:
      run: npx yarn eslint app/assets/javascripts test/javascripts

```

Run it:
```bash
lefthook run pre-commit
```

## Overcommit

```yml
PreCommit:
  RuboCop:
    enabled: true
    command: ['bundle', 'exec', 'rubocop', '--parallel']
  Danger:
    enabled: true
    command: ['bundle', 'exec', 'danger']
  EsLintAssets:
    enabled: true
    command: ['yarn', 'eslint', '--ext', '.es6', 'app/assets/javascripts']
  EsLintTest:
    enabled: true
    command: ['yarn', 'eslint', '--ext', '.es6', 'test/javascripts']
  EsLintPAssets:
    enabled: true
    command: ['yarn', 'eslint', '--ext', '.es6', 'plugins/**/assets/javascripts']
  EsLintPTest:
    enabled: true
    command: ['yarn', 'eslint', '--ext', '.es6', 'plugins/**/test/javascripts']
  EsLintAT:
    enabled: true
    command: ['yarn', 'eslint', 'app/assets/javascripts', 'test/javascripts']
```

Run it
```bash
overcommit --run
```

## Result
Lefthook ~20% faster than Overcommit.
