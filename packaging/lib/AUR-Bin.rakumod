use Constants;
use System;
use Lefthook::Package;
use AUR::Publishing;

my constant aur = $*PROGRAM.parent.parent.child("aur");

class AUR-Bin does Lefthook::Package {
  has System $.sys is required;

  submethod kind returns Lefthook::PackageKind { Lefthook::PackageKind::<aur-bin> }

  method clean {}

  method set-version {
    $!sys.replace(
      file => "{aur}/lefthook-bin/PKGBUILD",
      regex => /pkgver\s*'='.*$/,
      replacement => "pkgver={VERSION}",
    );
  }

  method prepare {}

  method publish {
    publish-aur-package(
      name => "lefthook-bin",
      sha256-urls => {
        sha256sum_linux_x86_64 => "https://github.com/evilmartians/lefthook/releases/download/v{VERSION}/lefthook_{VERSION}_Linux_x86_64.gz",
        sha256sum_linux_aarch64 => "https://github.com/evilmartians/lefthook/releases/download/v{VERSION}/lefthook_{VERSION}_Linux_aarch64.gz"
      },
      path-to-pkgbuild => "{aur}/lefthook-bin/PKGBUILD",
      sys => $!sys,
    );
  }
}
