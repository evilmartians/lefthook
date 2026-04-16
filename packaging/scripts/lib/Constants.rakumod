# Current lefthook version.
constant VERSION = "2.1.6";

# Git root.
constant REPO-ROOT = $?FILE.IO.parent(4);

# /packages/registries/
constant PKG-ROOT  = $?FILE.IO.parent(3).child("registries");

my constant DIST-ROOT = REPO-ROOT.child("dist");

# Supported platforms and architectures.
constant %DISTS = (
  amd64-linux   => "{DIST-ROOT}/no_self_update_linux_amd64_v1/lefthook",
  amd64-windows => "{DIST-ROOT}/no_self_update_windows_amd64_v1/lefthook.exe",
  amd64-darwin  => "{DIST-ROOT}/no_self_update_darwin_amd64_v1/lefthook",
  amd64-freebsd => "{DIST-ROOT}/no_self_update_freebsd_amd64_v1/lefthook",
  amd64-openbsd => "{DIST-ROOT}/no_self_update_openbsd_amd64_v1/lefthook",

  arm64-linux   => "{DIST-ROOT}/no_self_update_linux_arm64_v8.0/lefthook",
  arm64-windows => "{DIST-ROOT}/no_self_update_windows_arm64_v8.0/lefthook.exe",
  arm64-darwin  => "{DIST-ROOT}/no_self_update_darwin_arm64_v8.0/lefthook",
  arm64-freebsd => "{DIST-ROOT}/no_self_update_freebsd_arm64_v8.0/lefthook",
  arm64-openbsd => "{DIST-ROOT}/no_self_update_openbsd_arm64_v8.0/lefthook",
);
