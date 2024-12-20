## `jobs`

Jobs provide a flexible way to define tasks, supporting both commands and scripts. Jobs can be grouped which provides advanced flow control.

### Basic example

Define jobs in your `lefthook.yml` file under a specific hook like `pre-commit`:

```yml
# lefthook.yml

pre-commit:
  jobs:
    - run: yarn lint
    - run: yarn test
```

### Differences from Commands and Scripts

**Optional Job Names**

- Named jobs are merged across [`extends`](./extends.md) and local config.
- Unnamed jobs are appended in the order of their definition.

**Job Groups**

- Groups can include other jobs.
- Flow within groups can be parallel or piped. Options like `glob` and `root` apply to all jobs in the group, including nested ones.

### Job options

Below are the available options for configuring jobs.

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

> **Note:** Currently, only root and glob options are applied to group jobs. Other options must be set for each job individually. Submit a [feature request](https://github.com/evilmartians/lefthook/issues/new?assignees=&labels=feature+request&projects=&template=feature_request.md) if this limits your workflow.

A configuration demonstrating a piped group running in parallel with other jobs:

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

This configuration runs migrate jobs in a piped flow while other jobs execute in parallel.
