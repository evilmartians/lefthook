---
type: concept
title: assert_lefthook_installed
source: "https://lefthook.dev/configuration/assert_lefthook_installed/"
path: /configuration/assert_lefthook_installed/
updated: 2026-08-03
okf:
  generated_by: "@docmd/plugin-okf"
  generated_at: "2026-08-03T09:38:01.693Z"
---
---
title: "assert_lefthook_installed"
---

# `assert_lefthook_installed`

**Default: `false`**

When set to `true`, fail (with exit status 1) if `lefthook` executable can't be found in $PATH, under node_modules/, as a Ruby gem, or other supported method. This makes sure git hook won't omit `lefthook` rules if `lefthook` ever was installed.

#### Example

```yml
# lefthook.yml

assert_lefthook_installed: true
```
