#! /usr/bin/env bash

[[ -n "${_UTILS_SH:-}" ]] && return
_UTILS_SH=1

# copy source_file dest_file
#
# Creates missing directories for the `dest_file`
copy() {
  local source=$1
  local dest=$2

  if [[ ! -e $(dirname "$dest") ]]
  then
    mkdir -p $(dirname "$dest")
  fi

  cp "$source" "$dest"
}

fetch_sha256sum() {
  local url=$1

  curl -fsSL "$url" | sha256sum | cut -d' ' -f1
}
