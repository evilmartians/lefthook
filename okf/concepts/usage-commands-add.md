---
type: concept
title: "lefthook add"
source: "https://lefthook.dev/usage/commands/add/"
path: /usage/commands/add/
updated: 2026-08-28
okf:
  generated_by: "@docmd/plugin-okf"
  generated_at: "2026-08-28T10:12:38.551Z"
---
---
title: "lefthook add"
---

## `lefthook add`

Installs the given hook to Git hook.

With argument `--dirs` creates a directory `.git/hooks/<hook name>/` if it doesn't exist. Use it before adding a script to configuration.

#### Example

```bash
$ lefthook add pre-push  --dirs
```

Describe pre-push commands in `lefthook.yml`:

```yml
pre-push:
  jobs:
    - script: "audit.sh"
      runner: bash
```

Edit the script:

```bash
$ vim .lefthook/pre-push/audit.sh
...
```

Run `git push` and lefthook will run `bash audit.sh` as a pre-push hook.
