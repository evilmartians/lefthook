# lefthook-local.yml

::: callout tip Tip
You can put `lefthook-local.yml` into your `~/.gitignore`, so in every project you can have your local-only overrides.
:::

`lefthook-local.yml` overrides and extends the configuration of your main `lefthook.yml`.

```yml
# lefthook.yml

pre-commit:
  commands:
    lint:
      run: bundle exec rubocop {staged_files}
      glob: "*.rb"
    check-links:
      run: lychee {staged_files}
```

```yml
# lefthook-local.yml

pre-commit:
  parallel: true # run all commands concurrently
  commands:
    lint:
      run: docker-compose run backend {cmd} # wrap the original command with docker-compose
    check-links:
      skip: true # skip checking links

# Add another hook
post-merge:
  files: "git diff-tree -r --name-only --no-commit-id ORIG_HEAD HEAD"
  commands:
    dependencies:
      glob: "Gemfile*"
      run: docker-compose run backend bundle install
```

---

### The merged config lefthook will use

```yml

pre-commit:
  parallel: true
  commands:
    lint:
      run: docker-compose run backend bundle exec rubocop {staged_files}
      glob: "*.rb"
    check-links:
      run: lychee {staged_files}
      skip: true

post-merge:
  files: "git diff-tree -r --name-only --no-commit-id ORIG_HEAD HEAD"
  commands:
    dependencies:
      glob: "Gemfile*"
      run: docker-compose run backend bundle install
```
