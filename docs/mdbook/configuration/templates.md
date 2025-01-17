## `templates`

Provide custom replacement for templates in `run` values.

With `templates` you can specify what can be overridden via `lefthook-local.yml` without a need to overwrite every jobs in your configuration.

## Example

### Override with lefthook-local.yml

```yml
# lefthook.yml

templates:
  dip: # empty

pre-commit:
  jobs:
    # Will run: `bundle exec rubocop file1 file2 file3 ...`
    - run: {dip} bundle exec rubocop {staged_files}
```

```yml
# lefthook.yml

templates:
  dip: dip # Will run: `dip bundle exec rubocop file1 file2 file3 ...`
```

### Reduce redundancy

```yml
# lefthook.yml

templates:
  wrapper: docker-compose run --rm -v $(pwd):/app service

pre-commit:
  jobs:
    - run: {wrapper} yarn format
    - run: {wrapper} yarn lint
    - run: {wrapper} yarn test
```
