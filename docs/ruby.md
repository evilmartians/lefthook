# Lefthook in Ruby

This is guide of using Lefthook git hook manager in Ruby projects. You can find guides for other environments in [README.md](../README.md).

## Install

```bash
gem install lefthook
```

## Edit

Create and edit `lefthook.yml`:

```yml
pre-commit:
  parallel: true
  commands:
    audit:
      run: brakeman
    rubocop:
      files: git diff --name-only HEAD master
      glob: "*.{rb}"
      run: rubocop {files}
```

## Test it
```bash
lefthook install && lefthook run pre-commit
```

### Troubleshooting
If you see the error `lefthook: command not found` you need to check your $PATH. Also try to restart your terminal.

### More info
Have a question? Check the [wiki](https://github.com/Arkweid/lefthook/wiki).
