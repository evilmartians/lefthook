---
type: concept
title: "Wrap commands in local config"
source: "https://lefthook.dev/examples/wrap-commands/"
path: /examples/wrap-commands/
updated: 2026-08-03
okf:
  generated_by: "@docmd/plugin-okf"
  generated_at: "2026-08-03T09:38:01.708Z"
---
# Wrap commands in local config

Wrapping some commands defined in a main config with `dip`[^1].

```yml
# lefthook.yml

pre-commit:
  jobs:
    - name: rubocop
      run: bundle exec rubocop -A -- {staged_files}
```

```yml
# lefthook-local.yml

pre-commit:
  jobs:
    - name: rubocop
      run: dip {cmd}
```

[^1]: [dip](https://github.com/bibendi/dip) – dockerized dev experience with, similar to `docker-compose run`
