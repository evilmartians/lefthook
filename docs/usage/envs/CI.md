---
title: "CI"
---

## `CI`

When using NPM package `lefthook`, set `CI=true` in your CI (if it does not set it automatically) to prevent lefthook from installing hooks in the postinstall script:

```bash
CI=true npm install
CI=true yarn install
CI=true pnpm install
```

::: callout info Note
Set `LEFTHOOK=1` or `LEFTHOOK=true` to override this behavior and install hooks in the postinstall script (despite `CI=true`).
:::

