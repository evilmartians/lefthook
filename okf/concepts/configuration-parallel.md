---
type: concept
title: parallel
source: "https://lefthook.dev/configuration/parallel/"
path: /configuration/parallel/
updated: 2026-09-24
okf:
  generated_by: "@docmd/plugin-okf"
  generated_at: "2026-09-24T07:39:24.160Z"
---
---
title: "parallel"
---

# `parallel`

**Default: `false`**

::: callout info Note
Lefthook runs commands and scripts **sequentially** by default
:::

Run commands and scripts concurrently.

#### Example

```yml
# lefthook.yml

pre-commit:
  parallel: true
  commands:
    lint:
      run: yarn lint
    test:
      run: yarn test
```
