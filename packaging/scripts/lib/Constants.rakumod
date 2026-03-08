constant VERSION = "2.1.3";

constant REPO-ROOT = $?FILE.IO.parent(4);
constant PKG-ROOT  = $?FILE.IO.parent(3).child("registries");

my constant dist-path = REPO-ROOT.child("dist").Str;
constant %DISTS = (
  amd64-linux   => "{dist-path}/no_self_update_linux_amd64_v1/lefthook",
  amd64-windows => "{dist-path}/no_self_update_windows_amd64_v1/lefthook.exe",
  amd64-darwin  => "{dist-path}/no_self_update_darwin_amd64_v1/lefthook",
  amd64-freebsd => "{dist-path}/no_self_update_freebsd_amd64_v1/lefthook",
  amd64-openbsd => "{dist-path}/no_self_update_openbsd_amd64_v1/lefthook",

  arm64-linux   => "{dist-path}/no_self_update_linux_arm64_v8.0/lefthook",
  arm64-windows => "{dist-path}/no_self_update_windows_arm64_v8.0/lefthook.exe",
  arm64-darwin  => "{dist-path}/no_self_update_darwin_arm64_v8.0/lefthook",
  arm64-freebsd => "{dist-path}/no_self_update_freebsd_arm64_v8.0/lefthook",
  arm64-openbsd => "{dist-path}/no_self_update_openbsd_arm64_v8.0/lefthook",
);
