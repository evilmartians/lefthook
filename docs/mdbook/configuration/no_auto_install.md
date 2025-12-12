## `no_auto_install`

**Default: `false`**

Disable automatic installation and synchronization of git hooks when running lefthook. By default, lefthook automatically installs and updates hooks when you run `lefthook run` if the configuration has changed. Setting this to `true` disables that behavior.

This can also be controlled with the `--no-auto-install` option for the `lefthook run` command.

**Example**

```yml
# lefthook.yml

no_auto_install: true

pre-commit:
  commands:
    lint:
      run: npm run lint
```
