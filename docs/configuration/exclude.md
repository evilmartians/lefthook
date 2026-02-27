---
title: "exclude"
---

# `exclude`

This option allows to setup a list of globs for files to be excluded in files template.

> **Note:** The glob patterns used in `exclude` are affected by the [`glob_matcher`](./glob_matcher.md) setting. See the glob_matcher documentation for details on how `**` patterns behave.

**Example**

Run Rubocop on staged files with `.rb` extension except for `application.rb`, `routes.rb`, `rails_helper.rb`, and all Ruby files in `config/initializers/`.

```yml
# lefthook.yml

pre-commit:
  jobs:
    - name: lint
      glob: "*.rb"
      exclude:
        - config/routes.rb
        - config/application.rb
        - config/initializers/*.rb
        - spec/rails_helper.rb
      run: bundle exec rubocop --force-exclusion {staged_files}
```

If you've specified `exclude` but don't have a files template in [`run`](./run.md) option, lefthook will check `{staged_files}` for `pre-commit` hook and `{push_files}` for `pre-push` hook and apply filtering. If no files left, the command will be skipped.

```yml
# lefthook.yml

pre-commit:
  exclude:
    - "*/application.rb"
  jobs:
    - name: lint
      run: bundle exec rubocop # will skip if only application.rb was staged
```
