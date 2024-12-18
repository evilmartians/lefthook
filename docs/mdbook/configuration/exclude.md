## `exclude`

For the `exclude` option two variants are supported:

- A list of globs to be excluded
- A single regular expression (deprecated)


> **Note:** The regular expression is matched against full paths to files in the repo,
> relative to the repo root, using `/` as the directory separator on all platforms.
> File paths do not begin with the separator or any other prefix.

**Example**

Run Rubocop on staged files with `.rb` extension except for `application.rb`, `routes.rb`, `rails_helper.rb`, and all Ruby files in `config/initializers/`.

```yml
# lefthook.yml

pre-commit:
  commands:
    lint:
      glob: "*.rb"
      exclude:
        - config/routes.rb
        - config/application.rb
        - config/initializers/*.rb
        - spec/rails_helper.rb
      run: bundle exec rubocop --force-exclusion {staged_files}
```

The same example using a regular expression.

```yml
# lefthook.yml

pre-commit:
  commands:
    lint:
      glob: "*.rb"
      exclude: '(^|/)(application|routes|rails_helper|initializers/\w+)\.rb$'
      run: bundle exec rubocop --force-exclusion {staged_files}
```

**Important**

Be careful with the config file format's string quoting and escaping rules when writing regexps in it. For YAML, single quotes are often the simplest choice.

If you've specified `exclude` but don't have a files template in [`run`](./run.md) option, lefthook will check `{staged_files}` for `pre-commit` hook and `{push_files}` for `pre-push` hook and apply filtering. If no files left, the command will be skipped.

```yml
# lefthook.yml

pre-commit:
  commands:
    lint:
      exclude: '(^|/)application\.rb$'
      run: bundle exec rubocop # skipped if only application.rb was staged
```
