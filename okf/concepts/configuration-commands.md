---
type: concept
title: commands
source: "https://lefthook.dev/configuration/Commands/"
path: /configuration/Commands/
updated: 2026-08-25
okf:
  generated_by: "@docmd/plugin-okf"
  generated_at: "2026-08-25T14:37:43.142Z"
---
---
title: "commands"
---

# `commands`

Commands to be executed for the hook. Each command has a name and associated run [options](#command).

#### Example

```yml
# lefthook.yml

pre-commit:
  commands:
    lint:
      ... # command options
```

#### Command options

- [`run`](./run.md)
- [`skip`](./skip.md)
- [`only`](./only.md)
- [`tags`](./tags.md)
- [`glob`](./glob.md)
- [`files`](./files.md)
- [`file_types`](./file_types.md)
- [`env`](./env.md)
- [`root`](./root.md)
- [`exclude`](./exclude.md)
- [`fail_text`](./fail_text.md)
- [`stage_fixed`](./stage_fixed.md)
- [`interactive`](./interactive.md)
- [`use_stdin`](./use_stdin.md)
- [`priority`](./priority.md)
