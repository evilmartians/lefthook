---
type: concept
title: configs
source: "https://lefthook.dev/configuration/configs/"
path: /configuration/configs/
updated: 2026-08-17
okf:
  generated_by: "@docmd/plugin-okf"
  generated_at: "2026-08-17T13:12:03.059Z"
---
---
title: "configs"
---

# `configs`

**Default:** `[lefthook.yml]`

An optional array of config paths from remote's root.

#### Example

```yml
# lefthook.yml

remotes:
  - git_url: git@github.com:evilmartians/lefthook
    ref: v1.0.0
    configs:
      - examples/ruby-linter.yml
      - examples/test.yml
```

Example with multiple remotes merging multiple configurations.

```yml
# lefthook.yml

remotes:
  - git_url: git@github.com:org/lefthook-configs
    ref: v1.0.0
    configs:
      - examples/ruby-linter.yml
      - examples/test.yml
  - git_url: https://github.com/org2/lefthook-configs
    configs:
      - lefthooks/pre_commit.yml
      - lefthooks/post_merge.yml
  - git_url: https://github.com/org3/lefthook-configs
    ref: feature/new
    configs:
      - configs/pre-push.yml

```
