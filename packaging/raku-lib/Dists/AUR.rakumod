use Config;
use System;
use Package;
use Dists::AUR::Publishing;

my constant aur = PKG-ROOT.child("aur");
my constant pkgbuild = aur.child("lefthook").child("PKGBUILD");

class Dists::AUR does Package::Dist {
  has System $.sys is required;

  submethod kind returns Package::Kind { Package::Kind::<aur> }

  method clean {}

  method set-version {
    $!sys.replace(
      file => pkgbuild,
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
      path-to-pkgbuild => pkgbuild,
      sys => $!sys,
    );
  }
}
