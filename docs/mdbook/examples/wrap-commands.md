# Wrap commands in local config

Wrapping some commands defined in a main config with `dip`[^1].

```yml
# lefthook.yml

pre-commit:
  jobs:
    - name: rubocop
      run: bundle exec rubocop -A {staged_files}
```

```yml
# lefthook-local.yml

pre-commit:
  jobs:
    - name: rubocop
      run: dip {cmd}
```

[^1]: [dip](https://github.com/bibendi/dip) – dockerized dev experience with, similar to `docker-compose run`
