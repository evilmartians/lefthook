#! /usr/bin/env bash

[[ -n "${_GLOBALS_SH:-}" ]] && return
_GLOBALS_SH=1

VERSION="2.1.15"

lib="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

REPO_ROOT=$(cd ${lib}/../../../ && pwd)
BINS_ROOT="$REPO_ROOT/dist"
REGISTRIES_ROOT="$(cd ${lib}/../../ && pwd)/registries"

declare -rA BINARIES=(
  [amd64-linux]="${BINS_ROOT}/no_self_update_linux_amd64_v1/lefthook"
  [amd64-windows]="${BINS_ROOT}/no_self_update_windows_amd64_v1/lefthook.exe"
  [amd64-darwin]="${BINS_ROOT}/no_self_update_darwin_amd64_v1/lefthook"
  [amd64-freebsd]="${BINS_ROOT}/no_self_update_freebsd_amd64_v1/lefthook"
  [amd64-openbsd]="${BINS_ROOT}/no_self_update_openbsd_amd64_v1/lefthook"
  [arm64-linux]="${BINS_ROOT}/no_self_update_linux_arm64_v8.0/lefthook"
  [arm64-windows]="${BINS_ROOT}/no_self_update_windows_arm64_v8.0/lefthook.exe"
  [arm64-darwin]="${BINS_ROOT}/no_self_update_darwin_arm64_v8.0/lefthook"
  [arm64-freebsd]="${BINS_ROOT}/no_self_update_freebsd_arm64_v8.0/lefthook"
  [arm64-openbsd]="${BINS_ROOT}/no_self_update_openbsd_arm64_v8.0/lefthook"
)
