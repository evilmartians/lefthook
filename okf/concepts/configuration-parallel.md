---
type: concept
title: parallel
source: "https://lefthook.dev/configuration/parallel/"
path: /configuration/parallel/
updated: 2026-11-25
okf:
  generated_by: "@docmd/plugin-okf"
  generated_at: "2026-11-25T07:43:04.394Z"
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
