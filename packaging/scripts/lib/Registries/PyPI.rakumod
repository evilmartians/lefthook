use Constants;
use System;
use Registry;

my constant @platforms = <linux darwin windows> X <x86_64 arm64>;
my constant pypi = PKG-ROOT.child("pypi");

class Registries::PyPI does Registry::Package {
  has System $.sys is required;

  my constant %PYPI-DISTS = {
    amd64-linux   => "{pypi}/lefthook/bin/lefthook-linux-x86_64/lefthook",
    amd64-windows => "{pypi}/lefthook/bin/lefthook-windows-x86_64/lefthook.exe",
    amd64-darwin  => "{pypi}/lefthook/bin/lefthook-darwin-x86_64/lefthook",
    amd64-freebsd => "{pypi}/lefthook/bin/lefthook-freebsd-x86_64/lefthook",
    amd64-openbsd => "{pypi}/lefthook/bin/lefthook-openbsd-x86_64/lefthook",

    arm64-linux   => "{pypi}/lefthook/bin/lefthook-linux-arm64/lefthook",
    arm64-windows => "{pypi}/lefthook/bin/lefthook-windows-arm64/lefthook.exe",
    arm64-darwin  => "{pypi}/lefthook/bin/lefthook-darwin-arm64/lefthook",
    arm64-freebsd => "{pypi}/lefthook/bin/lefthook-freebsd-arm64/lefthook",
    arm64-openbsd => "{pypi}/lefthook/bin/lefthook-openbsd-arm64/lefthook",
  };

  submethod kind returns Registry::Kind { Registry::Kind::<pypi> }

  method clean {
    $!sys.rm(
      "{pypi}/lefthook/__pycache__/",
      "{pypi}/lefthook/bin/",
      "{pypi}/lefthook.egg-info/",
      "{pypi}/build/",
    )
  }

  method set-version {
    $!sys.replace(
      file => "{pypi}/pyproject.toml",
      regex => /^ \s* version \s* '=' .+ $/,
      replacement => qq[version = "{VERSION}"],
    );
  }

  method prepare {
    die "pypi/ setup is not complete" unless %PYPI-DISTS.keys.Set == %DISTS.keys.Set;

    $!sys.cp(.value, %PYPI-DISTS{.key}) for %DISTS.pairs;
  }

  method publish {
    $!sys.cd(pypi);

    for @platforms {
      my ($os, $arch) = $_;

      say "Build wheel for $os-$arch";
      %*ENV<LEFTHOOK_TARGET_PLATFORM> = $os;
      %*ENV<LEFTHOOK_TARGET_ARCH> = $arch;
      $!sys.run("uv build --wheel");
    }

    $!sys.run("uv publish");
  }
}
