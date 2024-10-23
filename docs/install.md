# Install lefthook

Choose your fighter:

- [Ruby](#ruby)
- [Node.js](#node)
- [Go](#go)
- [Python](#python)
- [Swift](#swift)
- [Scoop](#scoop)
- [Homebrew](#homebrew)
- [Winget](#winget)
- [Snap](#snap)
- [Debian-based distro](#deb)
- [RPM-based distro](#rpm)
- [Alpine](#alpine)
- [Arch Linux](#arch)
- [Manual](#else)

----

## <a id="ruby"></a> Ruby

```bash
gem install lefthook
```

**Troubleshooting**

If you see the error `lefthook: command not found` you need to check your $PATH. Also try to restart your terminal.


## <a id="node"></a> Node.js

Lefthook is available on NPM in the following flavors:

 1. [lefthook](https://www.npmjs.com/package/lefthook) that will install the proper binary:

    ```bash
    npm install lefthook --save-dev
    # or yarn:
    yarn add -D lefthook
    ```

 1. [@evilmartians/lefthook](https://www.npmjs.com/package/@evilmartians/lefthook) with pre-bundled binaries for all architectures:

    ```bash
    npm install @evilmartians/lefthook --save-dev
    # or yarn:
    yarn add -D @evilmartians/lefthook
    ```

 1. [@evilmartians/lefthook-installer](https://www.npmjs.com/package/@evilmartians/lefthook-installer) that will fetch binary file on installation:

    ```bash
    npm install @evilmartians/lefthook-installer --save-dev
    # or yarn:
    yarn add -D @evilmartians/lefthook-installer
    ```

> [!NOTE]
> If you use `pnpm` package manager make sure you set `side-effects-cache = false` in your .npmrc, otherwise the postinstall script of the lefthook package won't be executed and hooks won't be installed.

## <a id="go"></a> Go

```bash
go install github.com/evilmartians/lefthook@latest
```

## <a id="python"></a> Python

```sh
python -m pip install --user lefthook
```

## <a id="swift"></a> Swift

You can find the Swift wrapper plugin [here](https://github.com/csjones/lefthook-plugin).

Utilize lefthook in your Swift project using Swift Package Manager:

```swift
.package(url: "https://github.com/csjones/lefthook-plugin.git", exact: "1.8.1"),
```

Or, with [mint](https://github.com/yonaskolb/Mint):

```bash
mint run csjones/lefthook-plugin
```

## <a id="scoop"></a> Scoop for Windowss

```sh
scoop install lefthook
```

## <a id="homebrew"></a> Homebrew for MacOS and Linux

```bash
brew install lefthook
```

## <a id="winget"></a> Winget for Windows

```sh
winget install evilmartians.lefthook
```

## <a id="snap"></a> Snap for Linux

```sh
snap install --classic lefthook
```

## <a id="deb"></a> APT packages for Debian/Ubuntu Linux

```sh
curl -1sLf 'https://dl.cloudsmith.io/public/evilmartians/lefthook/setup.deb.sh' | sudo -E bash
sudo apt install lefthook
```
See all instructions: https://cloudsmith.io/~evilmartians/repos/lefthook/setup/#formats-deb

[![Hosted By: Cloudsmith](https://img.shields.io/badge/OSS%20hosting%20by-cloudsmith-blue?logo=cloudsmith&style=flat-square)](https://cloudsmith.com "Debian package repository hosting is graciously provided by Cloudsmith")

## <a id="rpm"></a> RPM packages for CentOS/Fedora Linux

```sh
curl -1sLf 'https://dl.cloudsmith.io/public/evilmartians/lefthook/setup.rpm.sh' | sudo -E bash
sudo yum install lefthook
```

See all instructions: https://cloudsmith.io/~evilmartians/repos/lefthook/setup/#repository-setup-yum

[![Hosted By: Cloudsmith](https://img.shields.io/badge/OSS%20hosting%20by-cloudsmith-blue?logo=cloudsmith&style=flat-square)](https://cloudsmith.com "RPM package repository hosting is graciously provided by Cloudsmith")

## <a id="alpine"></a> APK packages for Alpine

```sh
sudo apk add --no-cache bash curl
curl -1sLf 'https://dl.cloudsmith.io/public/evilmartians/lefthook/setup.alpine.sh' | sudo -E bash
sudo apk add lefthook
```

See all instructions: https://cloudsmith.io/~evilmartians/repos/lefthook/setup/#formats-alpine

[![Hosted By: Cloudsmith](https://img.shields.io/badge/OSS%20hosting%20by-cloudsmith-blue?logo=cloudsmith&style=flat-square)](https://cloudsmith.com "RPM package repository hosting is graciously provided by Cloudsmith")

## <a id="arch"></a> AUR for Arch

You can install lefthook [package](https://aur.archlinux.org/packages/lefthook) from AUR.

```sh
yay -S lefthook
```

## <a id="else"></a> Manuall installation with prebuilt executable

Or take it from [binaries](https://github.com/evilmartians/lefthook/releases) and install manually.

1. Download the executable for your OS and Arch
1. Put the executable under the $PATH (for unix systems)

### More info

Have a question?

<!-- :monocle_face: Check the [wiki](https://github.com/evilmartians/lefthook/wiki) -->

:thinking: Start a [discussion](https://github.com/evilmartians/lefthook/discussions)
