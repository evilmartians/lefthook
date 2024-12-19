## `jobs`

Job can either be a command or a script. Configuring `jobs` is more flexible than configuring `commands` and `scripts`, although all options are supported now.

```yml
# lefthook.yml

pre-commit:
  jobs:
    - run: yarn lint
    - run: yarn test
```

This is how jobs configuration differ from commands and scripts:

- Jobs have optional names. Lefthook merges named jobs across [extends](./extends.md) and [local configs](../examples/lefthook-local.md). Unnamed jobs get appended in the definition order.
- Jobs can have groups of other jobs. For groups you can specify [`parallel`](./parallel.md) or [`piped`](./piped.md) flow for a bunch of jobs. Also [`glob`](./glob.md) and [`root`](./root.md) parameters of a group apply to all its jobs (even nested).

### Job options

- [`name`](./name.md)
- [`run`](./run.md)
- [`script`](./script.md)
- [`runner`](./runner.md)
- [`group`](./group.md)
  - [`parallel`](./parallel.md)
  - [`piped`](./piped.md)
  - [`jobs`](./jobs.md)
- [`skip`](./skip.md)
- [`only`](./only.md)
- [`tags`](./tags.md)
- [`glob`](./glob.md)
- [`files`](./files.md)
- [`file_types`](./file_types.md)
- [`env`](./env.md)
- [`root`](./root.md)
- [`exclude`](./exclude.md)
- [`fail_text`](./fail_text.md)
- [`stage_fixed`](./stage_fixed.md)
- [`interactive`](./interactive.md)
- [`use_stdin`](./use_stdin.md)

### Example

> **Note:** Currently only `root` and `glob` options are applied to group jobs. Other options must be set for each job separately. If you find this inconvenient, please submit a [feature request](https://github.com/evilmartians/lefthook/issues/new?assignees=&labels=feature+request&projects=&template=feature_request.md).

A simple configuration with one piped group which executes in parallel with other jobs.

```yml
# lefthook.yml

pre-commit:
  parallel: true
  jobs:
    - name: migrate
      root: backend/
      glob: "db/migrations/*"
      group:
        piped: true
        jobs:
          - run: bundle install
          - run: rails db:migrate
    - run: yarn lint --fix {staged_files}
      root: frontend/
      stage_fixed: true
    - run: bundle exec rubocop
      root: backend/
    - run: golangci-lint
      root: proxy/
    - script: verify.sh
      runner: bash
```
