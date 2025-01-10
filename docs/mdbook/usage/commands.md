## Commands

> **Tip:** Use `lefthook help` or `lefthook <command> -h/--help` to discover available commands and their options

- [`lefthook install`](#lefthook-install)
- [`lefthook uninstall`](#lefthook-uninstall)
- [`lefthook add`](#lefthook-add)
- [`lefthook run`](#lefthook-run)
- [`lefthook version`](#lefthook-version)
- [`lefthook self-update`](#lefthook-self-update)
- [`lefthook validate`](#lefthook-validate)
- [`lefthook dump`](#lefthook-dump)

### `lefthook install`

`lefthook install` creates an empty `lefthook.yml` if a configuration file does not exist and updates git hooks to use lefthook. Run `lefthook install` after cloning the git repo.

> **Note:** NPM package `lefthook` installs the hooks in a postinstall script automatically

### `lefthook uninstall`

`lefthook uninstall` clears git hooks affected by lefthook. If you have lefthook installed as an NPM package you should remove it manually.

### `lefthook add`

`lefthook add pre-commit` will create a file `.git/hooks/pre-commit`. This is the same lefthook does for [`install`](#lefthook-install) command but you don't need to create a configuration first.

To use custom scripts as hooks create the required directories with `lefthook add pre-commit --dirs`.

**Example**

```bash
$ lefthook add pre-push --dirs
```

Describe pre-push commands in `lefthook.yml`:

```yml
pre-push:
  scripts:
    "audit.sh":
      runner: bash
```

Edit the script:

```bash
$ vim .lefthook/pre-push/audit.sh
...
```

Run `git push` and lefthook will run `bash audit.sh` as a pre-push hook.

### `lefthook run`

`lefthook run` executes the commands and scripts configured for a given hook. Generated hooks call `lefthook run` implicitly.

**Example**

```yml
# lefthook.yml

pre-commit:
  commands:
    lint:
      run: yarn lint --fix

test:
  commands:
    js-test:
      run: yarn test
```

Install the hook.

```bash
$ lefthook install
```

Run `test`.

```bash
$ lefthook run test # will run 'yarn test'
```

Commit changes.

```bash
$ git commit # will run pre-commit hook ('yarn lint --fix')
```

Or run manually also

```bash
$ lefthook run pre-commit
```

You can also specify a flag to run only some commands:

```bash
$ lefthook run pre-commit --commands lint
```

and optionally run either on all files (any `{staged_files}` placeholder acts as `{all_files}`) or a list of files:

```bash
$ lefthook run pre-commit --all-files
$ lefthook run pre-commit --file file1.js --file file2.js
```

(if both are specified, `--all-files` is ignored)

### `lefthook version`

`lefthook version` prints the current binary version. Print the commit hash with `lefthook version --full`

**Example**

```bash
$ lefthook version --full

1.1.3 bb099d13c24114d2859815d9d23671a32932ffe2
```

### `lefthook self-update`

`lefthook self-update` updates the binary with the latest lefthook release on Github. This command is available only if you install lefthook from sources or download the binary from the Github Releases. For other ways use package-specific commands to update lefthook.

### `lefthook validate`

`lefthook validate` loads JSON schema from the Github repo and validates you main lefthook config (e.g. `lefthook.yml`) and secondary configs (`lefthook-local.yml`, configs from `extends` and `remotes`). Use `lefthook dump` to get the full config and locate the issue.

### `lefthook dump`

`lefthook dump` prints the merged config. This is the actual config lefthook uses, it can be build from the main config (`lefthook.yml`), remotes, extends, and `lefthook-local.yml` overrides.
