---
type: concept
title: source_dir_local
source: "https://lefthook.dev/configuration/source_dir_local/"
path: /configuration/source_dir_local/
updated: 2026-08-28
okf:
  generated_by: "@docmd/plugin-okf"
  generated_at: "2026-08-28T10:16:43.987Z"
---
---
title: "source_dir_local"
---

# `source_dir_local`

**Default: `.lefthook-local/`**

Change a directory for *local* script files (not stored in VCS).

This option is useful if you have a `lefthook-local.yml` config file and want to reference different scripts there.

#### Example

```yml
# lefthook-local.yml

source_dir_local: .lefthook-local/
```
