## `run`

This is a mandatory option for a command. This is actually a command that is executed for the hook.

You can use files templates that will be substituted with the appropriate files on execution:

- `{files}` - custom [`files`](./files.md) command result.
- `{staged_files}` - staged files which you try to commit.
- `{push_files}` - files that are committed but not pushed.
- `{all_files}` - all files tracked by git.
- `{cmd}` - shorthand for the command from `lefthook.yml`.
- `{0}` - shorthand for the single space-joint string of git hook arguments.
- `{N}` - shorthand for the N-th git hook argument.

> **Note:** Command line length has a limit on every system. If your list of files is quite long, lefthook splits your files list to fit in the limit and runs few commands sequentially.

**Example**

Run `yarn lint` on `pre-commit` hook.

```yml
# lefthook.yml

pre-commit:
  commands:
    lint:
      run: yarn lint
```

#### `{files}` template

Run `go vet` only on files listed with `git ls-files -m` command with `.go` extension.

```yml
# lefthook.yml

pre-commit:
  commands:
    govet:
      files: git ls-files -m
      glob: "*.go"
      run: go vet {files}
```

#### `{staged_files}` template

Run `yarn eslint` only on staged files with `.js`, `.ts`, `.jsx`, and `.tsx` extensions.

```yml
# lefthook.yml

pre-commit:
  commands:
    eslint:
      glob: "*.{js,ts,jsx,tsx}"
      run: yarn eslint {staged_files}
```

#### `{push_files}` template

If you want to lint files only before pushing them.

```yml
# lefthook.yml

pre-push:
  commands:
    eslint:
      glob: "*.{js,ts,jsx,tsx}"
      run: yarn eslint {push_files}
```

#### `{all_files}` template

Simply run `bundle exec rubocop` on all files with `.rb` extension excluding `application.rb` and `routes.rb` files.

> **Note:** `--force-exclusion` will apply `Exclude` configuration setting of Rubocop

```yml
# lefthook.yml

pre-commit:
  commands:
    rubocop:
      tags:
        - backend
        - style
      glob: "*.rb"
      exclude:
        - config/application.rb
        - config/routes.rb
      run: bundle exec rubocop --force-exclusion {all_files}
```

#### `{cmd}` template

```yml
# lefthook.yml

pre-commit:
  commands:
    lint:
      run: yarn lint
  scripts:
    "good_job.js":
      runner: node
```

You can wrap it in docker runner locally:

```yml
# lefthook-local.yml

pre-commit:
  commands:
    lint:
      run: docker run -it --rm <container_id_or_name> {cmd}
  scripts:
    "good_job.js":
      runner: docker run -it --rm <container_id_or_name> {cmd}
```

#### Git arguments

Make sure commits are signed.

```yml
# lefthook.yml

# Note: commit-msg hook takes a single parameter,
#       the name of the file that holds the proposed commit log message.
# Source: https://git-scm.com/docs/githooks#_commit_msg
commit-msg:
  commands:
    multiple-sign-off:
      run: 'test $(grep -c "^Signed-off-by: " {1}) -lt 2'
```

#### Rubocop

If using `{all_files}` with RuboCop, it will ignore RuboCop's `Exclude` configuration setting. To avoid this, pass `--force-exclusion`.

#### Quotes

If you want to have all your files quoted with double quotes `"` or single quotes `'`, quote the appropriate shorthand:

```yml
# lefthook.yml

pre-commit:
  commands:
    lint:
      glob: "*.js"
      # Quoting with double quotes `"` might be helpful for Windows users
      run: yarn eslint "{staged_files}" # will run `yarn eslint "file1.js" "file2.js" "[strange name].js"`
    test:
      glob: "*.{spec.js}"
      run: yarn test '{staged_files}' # will run `yarn eslint 'file1.spec.js' 'file2.spec.js' '[strange name].spec.js'`
    format:
      glob: "*.js"
      # Will quote where needed with single quotes
      run: yarn test {staged_files} # will run `yarn eslint file1.js file2.js '[strange name].spec.js'`
```

#### Scripts

```yml
# lefthook.yml

pre-commit:
  jobs:
    - name: a whole script in a run
      run: |
        for file in $(ls .); do
          yarn lint $file
        done
```
