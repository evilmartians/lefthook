constant VERSION = "2.1.3";
constant REPO-ROOT = $*PROGRAM.parent.parent.parent.Str;
constant DIST-PATH = REPO-ROOT ~ "/dist";

constant %DISTS = {
  amd64-linux   => "{DIST-PATH}/no_self_update_linux_amd64_v1/lefthook",
  amd64-windows => "{DIST-PATH}/no_self_update_windows_amd64_v1/lefthook.exe",
  amd64-darwin  => "{DIST-PATH}/no_self_update_darwin_amd64_v1/lefthook",
  amd64-freebsd => "{DIST-PATH}/no_self_update_freebsd_amd64_v1/lefthook",
  amd64-openbsd => "{DIST-PATH}/no_self_update_openbsd_amd64_v1/lefthook",

  arm64-linux   => "{DIST-PATH}/no_self_update_linux_arm64_v8.0/lefthook",
  arm64-windows => "{DIST-PATH}/no_self_update_windows_arm64_v8.0/lefthook.exe",
  arm64-darwin  => "{DIST-PATH}/no_self_update_darwin_arm64_v8.0/lefthook",
  arm64-freebsd => "{DIST-PATH}/no_self_update_freebsd_arm64_v8.0/lefthook",
  arm64-openbsd => "{DIST-PATH}/no_self_update_openbsd_arm64_v8.0/lefthook",
};

