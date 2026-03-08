use Constants;
use System;
use SystemAPI;
use Registry;
use Registries::AUR::Publishing;

my constant aur = PKG-ROOT.child("aur");
my constant pkgbuild = aur.child("lefthook").child("PKGBUILD");

class Registries::AUR does Registry::Package {
  has SystemAPI $.sys is required;

  method kind(--> Registry::Kind:D) { Registry::Kind::aur }

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
