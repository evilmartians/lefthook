use Registry;

unit class Registries::AUR does Registry::Package;

use Constants;
use SystemAPI;
use Registries::AUR::Publishing;

my constant PKGBUILD = PKG-ROOT.child("aur").child("lefthook").child("PKGBUILD");

has SystemAPI $.sys is required;

method target(--> Registry::Target:D) { Registry::Target::aur }

method clean {}

method set-version {
  $!sys.replace(
    file => PKGBUILD,
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
    path-to-pkgbuild => PKGBUILD,
    sys => $!sys,
  );
}
