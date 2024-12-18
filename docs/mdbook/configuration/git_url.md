## `git_url`

A URL to Git repository. It will be accessed with privileges of the machine lefthook runs on.

**Example**

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
