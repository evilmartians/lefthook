---
type: concept
title: LEFTHOOK
source: "https://lefthook.dev/usage/envs/LEFTHOOK/"
path: /usage/envs/LEFTHOOK/
updated: 2026-08-21
okf:
  generated_by: "@docmd/plugin-okf"
  generated_at: "2026-08-21T08:12:44.716Z"
---
---
title: "LEFTHOOK"
---

## `LEFTHOOK`

Use `LEFTHOOK=0 git ...` or `LEFTHOOK=false git ...` to disable lefthook when running git commands.

#### Example

```bash
LEFTHOOK=0 git commit -am "Lefthook skipped"
```

When using NPM package `lefthook` in CI, and your CI sets `CI=true` automatically, use `LEFTHOOK=1` or `LEFTHOOK=true` to install hooks in the postinstall script:

#### Example

```bash
LEFTHOOK=1 npm install
LEFTHOOK=1 yarn install
LEFTHOOK=1 pnpm install
```
