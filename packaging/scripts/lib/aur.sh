#! /usr/bin/env bash

[[ -n "${_AUR_SH:-}" ]] && return
_AUR_SH=1

lib=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
source "$lib/globals.sh"
source "$lib/utils.sh"

set_version_aur() {
  local pkgbuild="$REGISTRIES_ROOT/aur/lefthook/PKGBUILD"

  replace "$pkgbuild" \
    'pkgver\s*=.*$' \
    "pkgver=${VERSION}"
}

publish_aur() {
  pushd "$REPO_ROOT"

  local pkgbuild="$REGISTRIES_ROOT/aur/lefthook/PKGBUILD"

  git clone ssh://aur@aur.archlinux.org/lefthook.git lefthook-aur

  local shasum
  shasum=$(fetch_sha256sum "https://github.com/evilmartians/lefthook/archive/v${VERSION}.tar.gz")

  cat "$pkgbuild" | sed -e "s/{{ sha256sum }}/$shasum/" > lefthook-aur/PKGBUILD

  push_new_version lefthook-aur

  popd
}

push_new_version() {
  pushd "$1"

  makepkg --printsrcinfo > .SRCINFO
  makepkg --noconfirm
  makepkg --install --noconfirm

  git config user.name "github-actions[bot]"
  git config user.email "github-actions[bot]@users.noreply.github.com"
  git add PKGBUILD .SRCINFO
  git commit -m release "v${VERSION}"
  git push origin master

  popd
}
