use Registry;

unit class Registries::AUR-Bin does Registry::Package;

use Constants;
use SystemAPI;
use Registries::AUR::Publishing;

my constant PKGBUILD = PKG-ROOT.child("aur").child("lefthook-bin").child("PKGBUILD");

has SystemAPI $.sys is required;

method kind(--> Registry::Kind:D) { Registry::Kind::aur-bin }

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
    name => "lefthook-bin",
    sha256-urls => {
      sha256sum_linux_x86_64 => "https://github.com/evilmartians/lefthook/releases/download/v{VERSION}/lefthook_{VERSION}_Linux_x86_64.gz",
      sha256sum_linux_aarch64 => "https://github.com/evilmartians/lefthook/releases/download/v{VERSION}/lefthook_{VERSION}_Linux_aarch64.gz"
    },
    path-to-pkgbuild => PKGBUILD,
    sys => $!sys,
  );
}
