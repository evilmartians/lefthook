use Registry;

unit class Registries::PyPI does Registry::Package;

use Constants;
use System;
use SystemAPI;

my constant PYPI = PKG-ROOT.child("pypi");
my constant @PLATFORMS = (
  ("linux",   "x86_64"),
  ("windows", "x86_64"),
  ("darwin",  "x86_64"),
  ("linux",   "arm64"),
  ("windows", "arm64"),
  ("darwin",  "arm64"),
);
my constant %PYPI-DISTS = {
  amd64-linux   => "{PYPI}/lefthook/bin/lefthook-linux-x86_64/lefthook",
  amd64-windows => "{PYPI}/lefthook/bin/lefthook-windows-x86_64/lefthook.exe",
  amd64-darwin  => "{PYPI}/lefthook/bin/lefthook-darwin-x86_64/lefthook",
  amd64-freebsd => "{PYPI}/lefthook/bin/lefthook-freebsd-x86_64/lefthook",
  amd64-openbsd => "{PYPI}/lefthook/bin/lefthook-openbsd-x86_64/lefthook",

  arm64-linux   => "{PYPI}/lefthook/bin/lefthook-linux-arm64/lefthook",
  arm64-windows => "{PYPI}/lefthook/bin/lefthook-windows-arm64/lefthook.exe",
  arm64-darwin  => "{PYPI}/lefthook/bin/lefthook-darwin-arm64/lefthook",
  arm64-freebsd => "{PYPI}/lefthook/bin/lefthook-freebsd-arm64/lefthook",
  arm64-openbsd => "{PYPI}/lefthook/bin/lefthook-openbsd-arm64/lefthook",
};

has SystemAPI $.sys is required;


method kind(--> Registry::Kind:D) { Registry::Kind::PYPI }

method clean {
  $!sys.rm(
    "{PYPI}/lefthook/__pycache__/",
    "{PYPI}/lefthook/bin/".IO.dir.grep(*.basename ne ".keep"),
    "{PYPI}/lefthook.egg-info/",
    "{PYPI}/build/",
  )
}

method set-version {
  $!sys.replace(
    file => "{PYPI}/pyproject.toml",
    regex => /^ \s* version \s* '=' .+ $/,
    replacement => qq[version = "{VERSION}"],
  );
}

method prepare {
  die "PYPI/ setup is not complete" unless %PYPI-DISTS.keys.Set == %DISTS.keys.Set;

  $!sys.cp(.value, %PYPI-DISTS{.key}) for %DISTS.pairs;
}

method publish {
  $!sys.in-dir(PYPI, {
    for @PLATFORMS {
      my ($os, $arch) = $_;

      say "Build wheel for $os-$arch";
      %*ENV<LEFTHOOK_TARGET_PLATFORM> = $os;
      %*ENV<LEFTHOOK_TARGET_ARCH> = $arch;
      $!sys.run("uv", "build", "--wheel");
    }

    $!sys.run("uv", "publish");
  });
}
