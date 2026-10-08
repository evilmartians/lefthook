#! /usr/bin/env bash

[[ -n "${_PHP_SH:-}" ]] && return
_PHP_SH=1

lib=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
source "$lib/globals.sh"
source "$lib/utils.sh"

set_version_php() {
  local php_dir="$REGISTRIES_ROOT/php"

  replace "$php_dir/bin/lefthook" \
    'LEFTHOOK_VERSION\s*=\s*.*;$' \
    "LEFTHOOK_VERSION = '${VERSION}';"
}

publish_php() {
  local php_dir="$REGISTRIES_ROOT/php"
  git clone git@github.com:evilmartians/lefthook-php.git lefthook-php

  copy "$php_dir/composer.json" lefthook-php/composer.json
  copy "$php_dir/bin/lefthook" lefthook-php/bin/lefthook

  pushd lefthook-php

  git config user.name "github-actions[bot]"
  git config user.email "github-actions[bot]@users.noreply.github.com"

  git add composer.json bin/lefthook
  git commit -m "${VERSION}: automated bump"
  git tag "v${VERSION}" -m "${VERSION}: automated bump"
  git push origin main --follow-tags

  popd
}
