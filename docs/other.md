# Lefthook in any environments

This is guide of using Lefthook git hook manager in any environments. You can find guides for Ruby and Node.js in [README.md](../README.md).

## Installation

Add Lefthook to your system or build it from sources.

### go

```bash
go get github.com/Arkweid/lefthook
```

### Homebrew for macOS

```bash
brew install Arkweid/lefthook/lefthook
```

### snap for Ubuntu

```bash
sudo snap install --devmode lefthook
```

### AUR for Arch

You can install lefthook [package](https://aur.archlinux.org/packages/lefthook) from AUR

Or take it from [binaries](https://github.com/Arkweid/lefthook/releases) and install manually


## Edit

Create and edit `lefthook.yml`:

```yml
pre-commit:
  parallel: true
  commands:
    audit:
      run: brakeman
    rubocop:
      files: git diff --name-only @{push}
      glob: "*.{rb}"
      run: rubocop {files}
```

## Test it
```bash
lefthook install && lefthook run pre-commit
```
