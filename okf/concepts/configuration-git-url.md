---
type: concept
title: git_url
source: "https://lefthook.dev/configuration/git_url/"
path: /configuration/git_url/
updated: 2026-08-17
okf:
  generated_by: "@docmd/plugin-okf"
  generated_at: "2026-08-17T13:11:15.250Z"
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
