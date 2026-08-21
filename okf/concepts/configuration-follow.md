---
type: concept
title: follow
source: "https://lefthook.dev/configuration/follow/"
path: /configuration/follow/
updated: 2026-08-21
okf:
  generated_by: "@docmd/plugin-okf"
  generated_at: "2026-08-21T08:12:44.688Z"
---
---
title: "follow"
---

# `follow`

**Default: `false`**

Follow the STDOUT of the running commands and scripts.

#### Example

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

::: callout info Note
If used with [`parallel`](#parallel) the output can be a mess, so please avoid setting both options to `true`
:::
