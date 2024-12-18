## `root`

You can change the CWD for the command you execute using `root` option.

This is useful when you execute some `npm` or `yarn` command but the `package.json` is in another directory.

For `pre-push` and `pre-commit` hooks and for the custom `files` command `root` option is used to filter file paths. If all files are filtered the command will be skipped.

**Example**

Format and stage files from a `client/` folder.

```bash
# Folders structure

$ tree .
.
├── client/
│   ├── package.json
│   ├── node_modules/
|   ├── ...
├── server/
|   ...
```

```yml
# lefthook.yml

pre-commit:
  commands:
    lint:
      root: "client/"
      glob: "*.{js,ts}"
      run: yarn eslint --fix {staged_files} && git add {staged_files}
```
