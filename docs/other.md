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

### APT packages for Debian/Ubuntu Linux

```sh
curl -1sLf 'https://dl.cloudsmith.io/public/evilmartians/lefthook/setup.deb.sh' | sudo -E bash
sudo apt install lefthook
```

See all instructions: https://cloudsmith.io/~evilmartians/repos/lefthook/setup/#formats-deb

[![Hosted By: Cloudsmith](https://img.shields.io/badge/OSS%20hosting%20by-cloudsmith-blue?logo=cloudsmith&style=flat-square)](https://cloudsmith.com "Debian package repository hosting is graciously provided by Cloudsmith")

### RPM packages for CentOS/Fedora Linux

```sh
curl -1sLf 'https://dl.cloudsmith.io/public/evilmartians/lefthook/setup.rpm.sh' | sudo -E bash
sudo yum install lefthook
```

See all instructions: https://cloudsmith.io/~evilmartians/repos/lefthook/setup/#repository-setup-yum

[![Hosted By: Cloudsmith](https://img.shields.io/badge/OSS%20hosting%20by-cloudsmith-blue?logo=cloudsmith&style=flat-square)](https://cloudsmith.com "RPM package repository hosting is graciously provided by Cloudsmith")

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
