---
type: concept
title: "Protect a branch from direct push"
source: "https://lefthook.dev/examples/protect-branch/"
path: /examples/protect-branch/
updated: 2026-10-07
okf:
  generated_by: "@docmd/plugin-okf"
  generated_at: "2026-10-07T19:00:00.695Z"
---
---
title: "Protect a branch from direct push"
---

# Protect a branch from direct push

This `pre-push` command blocks a push when the current branch is `main` or
`master`. You can still push by setting an env var for that one push.

```yml
# lefthook.yml

pre-push:
  commands:
    protect-main:
      run: |
        branch="$(git rev-parse --abbrev-ref HEAD)"
        case "$branch" in
          main|master)
            if [ "$ALLOW_PUSH_TO_MAIN" != "1" ]; then
              echo "Refusing direct push from protected branch '$branch'." >&2
              echo "Push a feature branch and open a PR instead, or set ALLOW_PUSH_TO_MAIN=1 to override." >&2
              exit 1
            fi
            ;;
        esac
```

Override for one push:

```
ALLOW_PUSH_TO_MAIN=1 git push
```

::: callout info Note
This command does not reference `{push_files}`, so lefthook skips it when a
push has no file changes (for example a push of an empty commit). It runs
normally for any push that changes a file.
:::
