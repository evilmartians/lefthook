---
title: "lefthook run"
---

## `lefthook run`

Executes the commands and scripts configured for a given hook. Installed Git hooks call `lefthook run` implicitly.

**Example**

```yml
# lefthook.yml

pre-commit:
  jobs:
    - name: lint
      run: yarn lint --fix {staged_files}

test:
  jobs:
    - name: test
      run: yarn test
```

Install the hook.

```bash
$ lefthook install
```

```bash
$ lefthook run test # will run 'yarn test'
$ git commit # will run pre-commit hook ('yarn lint --fix')
$ lefthook run pre-commit # will run pre-commit hook (`yarn lint --fix`)
```

### Run specific jobs

You can specify which jobs to run (also `--tag` supported).

```bash
$ lefthook run pre-commit --job lints --job pretty --tag checks
```

### Specify files

You can force replacing files templates (like `{staged_files}`) with either all files (will acts as `{all_files}` template) or a list of files.

```bash
$ lefthook run pre-commit --all-files
$ lefthook run pre-commit --file file1.js --file file2.js
```

(if both are specified, `--all-files` is ignored)
