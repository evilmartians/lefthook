#! /usr/bin/env bash

[[ -n "${_NPM_SH:-}" ]] && return
_NPM_SH=1

lib=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
source "$lib/globals.sh"
source "$lib/utils.sh"

NPM="$REGISTRIES_ROOT/npm"
NPM_BUNDLED="$REGISTRIES_ROOT/npm-bundled"
NPM_INSTALLER="$REGISTRIES_ROOT/npm-installer"

declare -ra PACKAGES=(
  ${NPM}/lefthook-darwin-arm64/
  ${NPM}/lefthook-darwin-x64/
  ${NPM}/lefthook-linux-arm64/
  ${NPM}/lefthook-linux-x64/
  ${NPM}/lefthook-windows-arm64/
  ${NPM}/lefthook-windows-x64/
  ${NPM}/lefthook-freebsd-arm64/
  ${NPM}/lefthook-freebsd-x64/
  ${NPM}/lefthook-openbsd-arm64/
  ${NPM}/lefthook-openbsd-x64/
  ${NPM}/lefthook/
  ${NPM_BUNDLED}/
  ${NPM_INSTALLER}/
)

declare -rA NPM_BINARIES=(
  [amd64-linux]="${NPM}/lefthook-linux-x64/bin/lefthook"
  [amd64-windows]="${NPM}/lefthook-windows-x64/bin/lefthook.exe"
  [amd64-darwin]="${NPM}/lefthook-darwin-x64/bin/lefthook"
  [amd64-freebsd]="${NPM}/lefthook-freebsd-x64/bin/lefthook"
  [amd64-openbsd]="${NPM}/lefthook-openbsd-x64/bin/lefthook"
  [arm64-linux]="${NPM}/lefthook-linux-arm64/bin/lefthook"
  [arm64-windows]="${NPM}/lefthook-windows-arm64/bin/lefthook.exe"
  [arm64-darwin]="${NPM}/lefthook-darwin-arm64/bin/lefthook"
  [arm64-freebsd]="${NPM}/lefthook-freebsd-arm64/bin/lefthook"
  [arm64-openbsd]="${NPM}/lefthook-openbsd-arm64/bin/lefthook"
)

declare -rA NPM_BUNDLED_BINARIES=(
  [amd64-linux]="${NPM_BUNDLED}/bin/lefthook-linux-x64/lefthook"
  [amd64-windows]="${NPM_BUNDLED}/bin/lefthook-windows-x64/lefthook.exe"
  [amd64-darwin]="${NPM_BUNDLED}/bin/lefthook-darwin-x64/lefthook"
  [amd64-freebsd]="${NPM_BUNDLED}/bin/lefthook-freebsd-x64/lefthook"
  [amd64-openbsd]="${NPM_BUNDLED}/bin/lefthook-openbsd-x64/lefthook"
  [arm64-linux]="${NPM_BUNDLED}/bin/lefthook-linux-arm64/lefthook"
  [arm64-windows]="${NPM_BUNDLED}/bin/lefthook-windows-arm64/lefthook.exe"
  [arm64-darwin]="${NPM_BUNDLED}/bin/lefthook-darwin-arm64/lefthook"
  [arm64-freebsd]="${NPM_BUNDLED}/bin/lefthook-freebsd-arm64/lefthook"
  [arm64-openbsd]="${NPM_BUNDLED}/bin/lefthook-openbsd-arm64/lefthook"
)

prepare_npm() {
  for package in ${PACKAGES[@]}
  do
    copy "$REPO_ROOT/README.md" "$package/README.md"
  done

  local -ra schemas=(
    "${NPM}/lefthook/schema.json"
    "${NPM_BUNDLED}/schema.json"
    "${NPM_INSTALLER}/schema.json"
  )
  for schema in ${schemas[@]}
  do
    copy "$REPO_ROOT/schema.json" "$schema"
  done

  [[ ${#BINARIES[@]} -eq ${#NPM_BINARIES[@]} ]]
  [[ ${#BINARIES[@]} -eq ${#NPM_BUNDLED_BINARIES[@]} ]]

  for kind in ${!BINARIES[@]}
  do
    copy "${BINARIES[$kind]}" "${NPM_BINARIES[$kind]}"
    copy "${BINARIES[$kind]}" "${NPM_BUNDLED_BINARIES[$kind]}"
  done
}

publish_npm() {
  for package in ${PACKAGES[@]}
  do
    pushd "$package"
    echo "Publishing $(basename "$package")"

    npm publish --access public
    popd
  done
}
