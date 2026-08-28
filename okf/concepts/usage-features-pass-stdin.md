---
type: concept
title: "Pass stdin to a command or script"
source: "https://lefthook.dev/usage/features/pass-stdin/"
path: /usage/features/pass-stdin/
updated: 2026-08-28
okf:
  generated_by: "@docmd/plugin-okf"
  generated_at: "2026-08-28T10:12:38.562Z"
---
---
title: "Pass stdin to a command or script"
---

# Pass stdin to a command or script

When you need to read the data from stdin – specify [`use_stdin: true`](../../configuration/use_stdin.md). This option is good when you write a command or script that receives data from git using stdin (for the `pre-push` hook, for example).


