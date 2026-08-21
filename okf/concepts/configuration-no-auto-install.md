---
type: concept
title: no_auto_install
source: "https://lefthook.dev/configuration/no_auto_install/"
path: /configuration/no_auto_install/
updated: 2026-08-21
okf:
  generated_by: "@docmd/plugin-okf"
  generated_at: "2026-08-21T08:12:44.693Z"
---
---
title: "no_auto_install"
---

# `no_auto_install`

**Default: `false`**

Disable automatic installation and synchronization of git hooks when running lefthook. By default, lefthook automatically installs and updates hooks when you run `lefthook run` if the configuration has changed. Setting this to `true` disables that behavior.

This can also be controlled with the `--no-auto-install` option for the `lefthook run` command.

#### Example

```yml
# lefthook.yml

no_auto_install: true

pre-commit:
  commands:
    lint:
      run: npm run lint
```
