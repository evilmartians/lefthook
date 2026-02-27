---
title: "follow"
---

# `follow`

**Default: `false`**

Follow the STDOUT of the running commands and scripts.

**Example**

```yml
# lefthook.yml

pre-push:
  follow: true
  commands:
    backend-tests:
      run: bundle exec rspec
    frontend-tests:
      run: yarn test
```

> **Note:** If used with [`parallel`](#parallel) the output can be a mess, so please avoid setting both options to `true`
