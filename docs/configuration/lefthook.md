---
title: "lefthook"
---

# `lefthook`

**Default:** `null`

> Added in lefthook `1.10.5`

Provide a full path to lefthook executable or a command to run lefthook. Bourne shell (`sh`) syntax is supported.

> **Important:** This option does not merge from `remotes` or `extends` for security reasons. But it gets merged from lefthook local config if specified.

There are three reasons you may want to specify `lefthook`:

1. You want to force using specific lefthook version from your dependencies (e.g. npm package)
1. You use PnP loader for your JS/TS project, and your `package.json` with lefthook dependency locates in a subfolder
1. You want to make sure you use concrete lefthook executable path and want to defined it in `lefthook-local.yml`

### Examples

#### Specify lefthook executable

```yml
# lefthook.yml

lefthook: /usr/bin/lefthook

pre-commit:
  jobs:
    - run: yarn lint
```

#### Specify a command to run lefthook

```yml
# lefthook.yml

lefthook: |
  cd project-with-lefthook
  pnpm lefthook

pre-commit:
  jobs:
    - run: yarn lint
      root: project-with-lefthook
```

#### Force using a version from Rubygems

```yml
# lefthook.yml

lefthook: bundle exec lefthook

pre-commit:
  jobs:
    - run: bundle exec rubocop {staged_files}
```

#### Enable debug logs

```yml
# lefthook-local.yml

lefthook: lefthook --verbose
```
