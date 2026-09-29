#! /usr/bin/env bash

[[ -n "${_RUBYGEM_SH:-}" ]] && return
_RUBYGEM_SH=1

lib=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
source "$lib/globals.sh"
source "$lib/utils.sh"

RUBYGEMS="$REGISTRIES_ROOT/rubygems"

declare -rA RUBYGEM_BINARIES=(
  [amd64-linux]="${RUBYGEMS}/libexec/lefthook-linux-x64/lefthook"
  [amd64-windows]="${RUBYGEMS}/libexec/lefthook-windows-x64/lefthook.exe"
  [amd64-darwin]="${RUBYGEMS}/libexec/lefthook-darwin-x64/lefthook"
  [amd64-freebsd]="${RUBYGEMS}/libexec/lefthook-freebsd-x64/lefthook"
  [amd64-openbsd]="${RUBYGEMS}/libexec/lefthook-openbsd-x64/lefthook"
  [arm64-linux]="${RUBYGEMS}/libexec/lefthook-linux-arm64/lefthook"
  [arm64-windows]="${RUBYGEMS}/libexec/lefthook-windows-arm64/lefthook.exe"
  [arm64-darwin]="${RUBYGEMS}/libexec/lefthook-darwin-arm64/lefthook"
  [arm64-freebsd]="${RUBYGEMS}/libexec/lefthook-freebsd-arm64/lefthook"
  [arm64-openbsd]="${RUBYGEMS}/libexec/lefthook-openbsd-arm64/lefthook"
)

set_version_rubygem() {
  replace "$RUBYGEMS/lefthook.gemspec" \
    '(spec\.version\s*)=.*$' \
    "\1= \"${VERSION}\""
}

prepare_rubygem() {
  [[ ${#BINARIES[@]} -eq ${#RUBYGEM_BINARIES[@]} ]]

  for kind in "${!BINARIES[@]}"
  do
    copy "${BINARIES[$kind]}" "${RUBYGEM_BINARIES[$kind]}"
  done
}

publish_rubygem() {
  echo "Publishing a ruby gem"

  pushd "$RUBYGEMS"
  rake build

  # Ensure we publish just one package
  local -a packages=$(ls ./pkg/*)
  [[ ${#packages[@]} -eq 1 ]]

  for package in "${packages[@]}"
  do
    gem push "$package"
  done

  popd
}
