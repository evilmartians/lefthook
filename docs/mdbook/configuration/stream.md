## `stream`

**Default: `false`**

Enable real-time streaming output for commands or scripts without requiring a TTY. This is useful when you want to see command output as it happens in non-interactive environments like VS Code's git commit interface, CI/CD pipelines, or any environment where a TTY is not available.

Unlike [`interactive`](./interactive.md), `stream` does not attempt to open `/dev/tty` for user input, making it work seamlessly in environments without TTY access.

**Example**

Use this option when you want to see real-time output from long-running commands or scripts, such as linters or test runners, even when running commits from editors or IDEs:

```yml
# lefthook.yml
pre-commit:
  commands:
    tests:
      run: npm test
      stream: true
    lint:
      run: npm run lint
      stream: true
```

**Comparison with other options**

- `stream: true` - Streams output in real-time, no TTY required, no user input
- `interactive: true` - Requires TTY for user input, will fail if TTY is unavailable
- `use_stdin: true` - Passes stdin to the command but doesn't control output streaming
- Default (all false) - Uses pseudo-TTY for output buffering

**When to use**

- You want real-time output visibility
- You're running hooks from non-TTY environments (VS Code, GitHub Desktop, etc.)
- Your command produces output you want to see immediately
- You don't need user input during execution
