## lefthook-local.yml

`lefthook-local.yml` overrides and extends the configuration of your main `lefthook.yml` (or `lefthook.toml`, [etc.](../configuration)) file.

> **Tip:** You can put `lefthook-local.yml` into your `~/.gitignore`, so in every project you can have your local-only overrides.

*Special feature* of `lefthook-local.yml`: you can wrap the commands using `{cmd}` template.

```yml
# lefthook.yml

pre-commit:
  commands:
    lint-frontend:
      run: yarn lint
      glob: ".{ts,tsx}"
    lint-backend:
      run: bundle exec rubocop {staged_files}
      glob: "*.rb"
    test-frontend:
      run: yarn test
      glob: "*.tsx"
    test-backend:
      run: bundle exec rspec
      glob: "spec/*"
    check-typos:
      run: typos {staged_files}
    check-links:
      run: lychee {staged_files}
```

```yml
# lefthook-local.yml

pre-commit:
  parallel: true # run all commands concurrently
  commands:
    lint-backend:
      run: docker-compose run backend {cmd} # wrap the original command with docker-compose
    test-backend:
      run: docker-compose run backend {cmd}
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

```yml
# The resulting config would look like this

pre-commit:
  parallel: true
  commands:
    lint-frontend:
      run: yarn lint
      glob: "*.{ts,tsx}"
    lint-backend:
      run: docker-compose run backend bundle exec rubocop {staged_files}
      glob: "*.rb"
    test-frontend:
      run: yarn test
      glob: "*.tsx"
    test-backend:
      run: docker-compose run backend bundle exec rspec
      glob: "spec/*"
    check-links:
      run: lychee {staged_files}
      skip: true
    check-typos:
      run: typos {staged_files}

post-merge:
  files: "git diff-tree -r --name-only --no-commit-id ORIG_HEAD HEAD"
  commands:
    dependencies:
      glob: "Gemfile*"
      run: docker-compose run backend bundle install
```
