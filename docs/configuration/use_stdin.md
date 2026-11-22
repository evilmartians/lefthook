---
title: "use_stdin"
---

# `use_stdin`

::: callout info Note
Each stdin-enabled command or script receives its own copy of stdin. For `pre-push` hooks invoked by Git, this contains the ref-update lines.
:::

Pass the stdin from the OS to the command/script.

A `pre-push` command that reads refs directly can omit file templates and file
filters to receive deletion-only pushes too. See [`run`](./run.md) for file-based
skipping.

#### Example

Use this option for the `pre-push` hook when you have a script that does `while read ...`. Without this option lefthook will hang: lefthook uses [pseudo TTY](https://github.com/creack/pty) by default, and it doesn't close stdin when all data is read.

```bash
# .lefthook/pre-push/do-the-magic.sh

remote="$1"
url="$2"

while read local_ref local_oid remote_ref remote_oid; do
  # ...
done
```

```yml
# lefthook.yml
pre-push:
  scripts:
    "do-the-magic.sh":
      runner: bash
      use_stdin: true
```
