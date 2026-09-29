#! /usr/bin/env bash

[[ -n "${_AUR_BIN_SH:-}" ]] && return
_AUR_BIN_SH=1

lib=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
source "$lib/globals.sh"
source "$lib/utils.sh"
source "$lib/aur.sh" # for push_new_version

set_version_aur_bin() {
  local pkgbuild="$REGISTRIES_ROOT/aur/lefthook-bin/PKGBUILD"

  replace "$pkgbuild" \
    'pkgver\s*=.*$' \
    "pkgver=${VERSION}"
}

publish_aur_bin() {
  pushd "$REPO_ROOT"

  local pkgbuild="$REGISTRIES_ROOT/aur/lefthook-bin/PKGBUILD"

  git clone ssh://aur@aur.archlinux.org/lefthook-bin.git lefthook-bin-aur

  local sha256sum_linux_x86_64
  local sha256sum_linux_aarch64
  sha256sum_linux_x86_64=$(fetch_sha256sum "https://github.com/evilmartians/lefthook/releases/download/v${VERSION}/lefthook_${VERSION}_Linux_x86_64.gz")
  sha256sum_linux_aarch64=$(fetch_sha256sum "https://github.com/evilmartians/lefthook/releases/download/v${VERSION}/lefthook_${VERSION}_Linux_aarch64.gz")

  cat "$pkgbuild" \
    | sed -e "s/{{ sha256sum_linux_x68_64 }}/$sha256sum_linux_x86_64/" \
          -e "s/{{ sha256sum_linux_aarch64 }}/$sha256sum_linux_aarch64/" \
    > lefthook-bin-aur/PKGBUILD

  push_new_version lefthook-bin-aur

  popd
}
