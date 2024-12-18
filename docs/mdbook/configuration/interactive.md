## `interactive`

**Default: `false`**

> **Note:** If you want to pass stdin to your command or script but don't need to get the input from CLI, use [`use_stdin`](./use_stdin.md) option instead.


Whether to use interactive mode. This applies the certain behavior:
- All `interactive` commands/scripts are executed after non-interactive. Exception: [`piped`](./piped.md) option is set to `true`.
- When executing, lefthook tries to open /dev/tty (Linux/Unix only) and use it as stdin.
- When [`no_tty`](./no_tty.md) option is set, `interactive` is ignored.
