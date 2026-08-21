---
type: concept
title: Remotes
source: "https://lefthook.dev/examples/remotes/"
path: /examples/remotes/
updated: 2026-08-21
okf:
  generated_by: "@docmd/plugin-okf"
  generated_at: "2026-08-21T08:12:44.705Z"
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
