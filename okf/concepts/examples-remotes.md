---
type: concept
title: Remotes
source: "https://lefthook.dev/examples/remotes/"
path: /examples/remotes/
updated: 2026-08-28
okf:
  generated_by: "@docmd/plugin-okf"
  generated_at: "2026-08-28T10:16:43.990Z"
---
# Remotes

Use configurations from other Git repositories via `remotes` feature.

Lefthook will automatically download the remote config files and merge them into existing configuration.

```yml
remotes:
  - git_url: https://github.com/evilmartians/lefthook
    configs:
      - examples/remote/ping.yml
```
