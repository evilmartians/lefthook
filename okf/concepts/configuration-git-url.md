---
type: concept
title: git_url
source: "https://lefthook.dev/configuration/git_url/"
path: /configuration/git_url/
updated: 2026-10-08
okf:
  generated_by: "@docmd/plugin-okf"
  generated_at: "2026-10-08T08:35:24.030Z"
---
---
title: "git_url"
---

# `git_url`

A URL to Git repository. It will be accessed with privileges of the machine lefthook runs on.

#### Example

```yml
# lefthook.yml

remotes:
  - git_url: git@github.com:evilmartians/lefthook
```

Or

```yml
# lefthook.yml

remotes:
  - git_url: https://github.com/evilmartians/lefthook
```
