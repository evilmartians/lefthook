# Lefthook in any environments

This is the guide to use the Lefthook git hook manager in any environment. You can find guides for Ruby and Node.js in [README.md](../README.md).

## Installation

Add Lefthook to your system or build it from source.

### go

```bash
go get github.com/evilmartians/lefthook
```

### Homebrew for MacOS and Linux

```bash
brew install lefthook
```

### Snap for Linux

```sh
snap install --classic lefthook
```

### AUR for Arch

You can install lefthook [package](https://aur.archlinux.org/packages/lefthook) from AUR

### Anything else

Or take it from [binaries](https://github.com/evilmartians/lefthook/releases) and install manually

### pip for Python

You can find Python wrapper here [package](https://github.com/life4/lefthook)

## Edit

Create and edit `lefthook.yml`:

```yml
pre-commit:
  parallel: true
  commands:
    audit:
      run: brakeman --no-pager
    rubocop:
      files: git diff --name-only @{push}
      glob: "*.rb"
      run: rubocop {files}
```

## Test it
```bash
lefthook install && lefthook run pre-commit
```

### More info
Have a question? Check the [wiki](https://github.com/evilmartians/lefthook/wiki).
