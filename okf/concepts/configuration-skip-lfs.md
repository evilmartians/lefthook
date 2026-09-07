---
type: concept
title: skip_lfs
source: "https://lefthook.dev/configuration/skip_lfs/"
path: /configuration/skip_lfs/
updated: 2026-09-07
okf:
  generated_by: "@docmd/plugin-okf"
  generated_at: "2026-09-07T07:37:00.172Z"
---
---
title: "skip_lfs"
---

# `skip_lfs`

**Default:** `false`

Skip running LFS hooks even if it exists on your system.

#### Example

```yml
# lefthook.yml

skip_lfs: true

pre-push:
  commands:
    test:
      run: yarn test
```
