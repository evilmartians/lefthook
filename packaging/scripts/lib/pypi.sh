#! /usr/bin/env bash

[[ -n "${_PYPI_SH:-}" ]] && return
_PYPI_SH=1

lib=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
source "$lib/globals.sh"
source "$lib/utils.sh"

PYPI="$REGISTRIES_ROOT/pypi"

declare -rA PYPI_BINARIES=(
  [amd64-linux]="${PYPI}/lefthook/bin/lefthook-linux-x86_64/lefthook"
  [amd64-windows]="${PYPI}/lefthook/bin/lefthook-windows-x86_64/lefthook.exe"
  [amd64-darwin]="${PYPI}/lefthook/bin/lefthook-darwin-x86_64/lefthook"
  [amd64-freebsd]="${PYPI}/lefthook/bin/lefthook-freebsd-x86_64/lefthook"
  [amd64-openbsd]="${PYPI}/lefthook/bin/lefthook-openbsd-x86_64/lefthook"
  [arm64-linux]="${PYPI}/lefthook/bin/lefthook-linux-arm64/lefthook"
  [arm64-windows]="${PYPI}/lefthook/bin/lefthook-windows-arm64/lefthook.exe"
  [arm64-darwin]="${PYPI}/lefthook/bin/lefthook-darwin-arm64/lefthook"
  [arm64-freebsd]="${PYPI}/lefthook/bin/lefthook-freebsd-arm64/lefthook"
  [arm64-openbsd]="${PYPI}/lefthook/bin/lefthook-openbsd-arm64/lefthook"
)

set_version_pypi() {
  replace "$PYPI/pyproject.toml" \
    '^\s*version\s*=.+$' \
    "version = \"${VERSION}\""
}

prepare_pypi() {
  [[ ${#BINARIES[@]} -eq ${#PYPI_BINARIES[@]} ]]

  for kind in "${!BINARIES[@]}"
  do
    copy "${BINARIES[$kind]}" "${PYPI_BINARIES[$kind]}"
  done
}

publish_pypi() {
  pushd "$PYPI"

  uv_build "linux"   "x86_64"
  uv_build "linux"   "arm64"
  uv_build "windows" "x86_64"
  uv_build "windows" "arm64"
  uv_build "darwin"  "x86_64"
  uv_build "darwin"  "arm64"

  uv publish

  popd
}

uv_build() {
  local os=$1
  local arch=$2
  echo "Build wheel for $os-$arch"

  LEFTHOOK_TARGET_PLATFORM=$os LEFTHOOK_TARGET_ARCH=$arch uv build --wheel
}
