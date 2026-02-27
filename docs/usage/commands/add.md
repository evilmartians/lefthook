---
title: "lefthook add"
---

## `lefthook add`

Installs the given hook to Git hook.

With argument `--dirs` creates a directory `.git/hooks/<hook name>/` if it doesn't exist. Use it before adding a script to configuration.

**Example**

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
