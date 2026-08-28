---
type: concept
title: piped
source: "https://lefthook.dev/configuration/piped/"
path: /configuration/piped/
updated: 2026-08-28
okf:
  generated_by: "@docmd/plugin-okf"
  generated_at: "2026-08-28T10:16:43.982Z"
---
---
title: "piped"
---

# `piped`

**Default: `false`**

::: callout info Note
Lefthook will return an error if both `piped: true` and `parallel: true` are set
:::

Stop running commands and scripts if one of them fail.

#### Example

```yml
# lefthook.yml

database:
  piped: true # Stop if one of the steps fail
  commands:
    1_create:
      run: rake db:create
    2_migrate:
      run: rake db:migrate
    3_seed:
      run: rake db:seed
```
