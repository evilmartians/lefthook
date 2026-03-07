use Constants;
use System;
use Package;
use Dists::AUR::Publishing;

my constant aur = $*PROGRAM.parent.parent.child("aur");

class Dists::AUR does Package::Dist {
  has System $.sys is required;

  submethod kind returns Package::Kind { Package::Kind::<aur> }

  method clean {}

  method set-version {
    $!sys.replace(
      file => "{aur}/lefthook/PKGBUILD",
      regex => /pkgver\s*'='.*$/,
      replacement => "pkgver={VERSION}",
    );
  }

  method prepare {}

  method publish {
    publish-aur-package(
      name => "lefthook",
      sha256-urls => {
        sha256sum => "https://github.com/evilmartians/lefthook/archive/v{VERSION}.tar.gz",
      },
      path-to-pkgbuild => "{aur}/lefthook/PKGBUILD",
      sys => $!sys,
    );
  }
}
