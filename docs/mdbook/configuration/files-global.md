### `files` (global)

A custom git command for files to be referenced in `{files}` template. See [`run`](#run) and [`files`](#files).

If the result of this command is empty, the execution of commands will be skipped.

**Example**

```yml
# lefthook.yml

pre-commit:
  files: git diff --name-only master # custom list of files
  commands:
    ...
```
