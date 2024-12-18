## `files`

A custom git command for files or directories to be referenced in `{files}` template for [`run`](./run.md) setting.

If the result of this command is empty, the execution of commands will be skipped.

This option overwrites the [hook-level `files`](./files-global.md) option.

**Example**

Provide a git command to list files.

```yml
# lefthook.yml

pre-push:
  commands:
    stylelint:
      tags:
        - frontend
        - style
      files: git diff --name-only master
      glob: "*.js"
      run: yarn stylelint {files}
```

Call a custom script for listing files.

```yml
# lefthook.yml

pre-push:
  commands:
    rubocop:
      tags: backend
      glob: "**/*.rb"
      files: node ./lefthook-scripts/ls-files.js # you can call your own scripts
      run: bundle exec rubocop --force-exclusion --parallel {files}
```
