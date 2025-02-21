## `glob`

You can set a glob to filter files for your command. This is only used if you use a file template in [`run`](./run.md) option or provide your custom [`files`](./files.md) command.

**Example**

```yml
# lefthook.yml

pre-commit:
  jobs:
    - name: lint
      run: yarn eslint {staged_files}
      glob: "*.{js,ts,jsx,tsx}"
```

> **Note:** from lefthook version `1.10.10` you can also provide a list of globs:
>
> ```yml
> # lefthook.yml
>
> pre-commit:
>   jobs:
>     - run: yarn lint {staged_files}
>       glob:
>         - "*.ts"
>         - "*.js"
> ```

**Notes**

For patterns that you can use see [this](https://tldp.org/LDP/GNU-Linux-Tools-Summary/html/x11655.htm) reference. We use [glob](https://github.com/gobwas/glob) library.

***When using `root:`***

Globs are still calculated from the actual root of the git repo, `root` is ignored.

***Behaviour of `**`***

Note that the behaviour of `**` is different from typical glob implementations, like `ls` or tools like `lint-staged` in that a double-asterisk matches 1+ directories deep, not zero or more directories.
If you want to match *both* files at the top level and nested, then rather than:

```yaml
glob: "src/**/*.js"
```

You'll need:

```yaml
glob: "src/*.js"
```

***Using `glob` without a files template in`run`***

If you've specified `glob` but don't have a files template in [`run`](./run.md) option, lefthook will check `{staged_files}` for `pre-commit` hook and `{push_files}` for `pre-push` hook and apply filtering. If no files left, the command will be skipped.

```yml
# lefthook.yml

pre-commit:
  jobs:
    - name: lint
      run: npm run lint # skipped if no .js files staged
      glob: "*.js"
```
