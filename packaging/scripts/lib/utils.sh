#! /usr/bin/env bash

[[ -n "${_UTILS_SH:-}" ]] && return
_UTILS_SH=1

# copy source_file dest_file
#
# Creates missing directories for the `dest_file`
copy() {
  local src=$1
  local dest=$2

  local dest_dir
  dest_dir="$(dirname "$dest")"
  if [[ ! -e "$dest_dir" ]]
  then
    mkdir -p "$dest_dir"
  fi

  cp "$src" "$dest"
}

replace() {
  local file=$1
  local regex=$2
  local replacement=$3

  echo "updating version in $file"

  perl -pi -e "s/$regex/$replacement/" "$file"
}

fetch_sha256sum() {
  local url=$1

  curl -fsSL "$url" | sha256sum | cut -d' ' -f1
}
