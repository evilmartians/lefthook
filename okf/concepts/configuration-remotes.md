---
type: concept
title: remotes
source: "https://lefthook.dev/configuration/remotes/"
path: /configuration/remotes/
updated: 2026-08-28
okf:
  generated_by: "@docmd/plugin-okf"
  generated_at: "2026-08-28T10:16:43.984Z"
---
---
title: "remotes"
---

# `remotes`

You can provide multiple remote configs if you want to share yours lefthook configurations across many projects. Lefthook will automatically download and merge configurations into your local `lefthook.yml`.

You can use [`extends`](./extends.md) but the paths must be relative to the remote repository root.

If you provide [`scripts`](./scripts.md) in a remote config file, the [script `source_dir`](./source_dir.md) must also be in the **root of the remote repository**.

::: callout info Note
Configs are merged in this order: `lefthook.yml` → `remotes` → `lefthook-local.yml`. For simplicity, keep jobs in remote configs independent from other steps.
:::

#### Example

```yml
# lefthook.yml

remotes:
  - git_url: git@github.com:evilmartians/lefthook
    ref: v1.0.0
    configs:
      - examples/ruby-linter.yml
```
