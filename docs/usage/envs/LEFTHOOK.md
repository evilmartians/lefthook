---
title: "LEFTHOOK"
---

## `LEFTHOOK`

Use `LEFTHOOK=0 git ...` or `LEFTHOOK=false git ...` to disable lefthook when running git commands.

**Example**

```bash
LEFTHOOK=0 git commit -am "Lefthook skipped"
```

When using NPM package `lefthook` in CI, and your CI sets `CI=true` automatically, use `LEFTHOOK=1` or `LEFTHOOK=true` to install hooks in the postinstall script:

**Example**

```bash
LEFTHOOK=1 npm install
LEFTHOOK=1 yarn install
LEFTHOOK=1 pnpm install
```

