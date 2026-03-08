use Constants;
use System;
use Registry;
use Registries::AUR::Publishing;

my constant aur = PKG-ROOT.child("aur");
my constant pkgbuild = aur.child("lefthook-bin").child("PKGBUILD");

class Registries::AUR-Bin does Registry::Package {
  has System $.sys is required;

  submethod kind returns Registry::Kind { Registry::Kind::<aur-bin> }

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
      name => "lefthook-bin",
      sha256-urls => {
        sha256sum_linux_x86_64 => "https://github.com/evilmartians/lefthook/releases/download/v{VERSION}/lefthook_{VERSION}_Linux_x86_64.gz",
        sha256sum_linux_aarch64 => "https://github.com/evilmartians/lefthook/releases/download/v{VERSION}/lefthook_{VERSION}_Linux_aarch64.gz"
      },
      path-to-pkgbuild => pkgbuild,
      sys => $!sys,
    );
  }
}
