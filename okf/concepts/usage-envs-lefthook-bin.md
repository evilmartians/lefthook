---
type: concept
title: LEFTHOOK_BIN
source: "https://lefthook.dev/usage/envs/LEFTHOOK_BIN/"
path: /usage/envs/LEFTHOOK_BIN/
updated: 2026-08-21
okf:
  generated_by: "@docmd/plugin-okf"
  generated_at: "2026-08-21T07:57:38.030Z"
---
---
title: "LEFTHOOK_BIN"
---

## `LEFTHOOK_BIN`

Set `LEFTHOOK_BIN` to a location where lefthook is installed to use that instead of trying to detect from the it the PATH or from a package manager.

Useful for cases when:

- lefthook is installed multiple ways, and you want to be explicit about which one is used (example: installed through homebrew, but also is in Gemfile but you are using a ruby version manager like rbenv that prepends it to the path)
- debugging and/or developing lefthook
