use Registry;

unit class Registries::PHP does Registry::Package;

use Constants;
use SystemAPI;
use Registries::PHP::Publishing;

my constant PHP = PKG-ROOT.child("php");
my constant %PHP-DISTS = (
  amd64-linux   => "{PHP}/libexec/lefthook-linux-x64/lefthook",
  amd64-windows => "{PHP}/libexec/lefthook-windows-x64/lefthook.exe",
  amd64-darwin  => "{PHP}/libexec/lefthook-darwin-x64/lefthook",
  amd64-freebsd => "{PHP}/libexec/lefthook-freebsd-x64/lefthook",
  amd64-openbsd => "{PHP}/libexec/lefthook-openbsd-x64/lefthook",

  arm64-linux   => "{PHP}/libexec/lefthook-linux-arm64/lefthook",
  arm64-windows => "{PHP}/libexec/lefthook-windows-arm64/lefthook.exe",
  arm64-darwin  => "{PHP}/libexec/lefthook-darwin-arm64/lefthook",
  arm64-freebsd => "{PHP}/libexec/lefthook-freebsd-arm64/lefthook",
  arm64-openbsd => "{PHP}/libexec/lefthook-openbsd-arm64/lefthook",
);

has SystemAPI $.sys is required;

method target(--> Registry::Target:D) { Registry::Target::php }

method clean {
  $!sys.rm("{PHP}/libexec/".IO.dir.grep(*.d));
}

method set-version {
  # Nothing to do: Composer/Packagist derives the package version from the
  # repository's git tags, so composer.json must not carry a version field.
}

method prepare {
  die "php/ setup is not complete" unless %PHP-DISTS.keys.Set == %DISTS.keys.Set;

  for %DISTS.kv -> $platform, $source {
    $!sys.cp($source, %PHP-DISTS{$platform});
  }
}

method publish {
  my $repo-url = %*ENV<LEFTHOOK_COMPOSER_REPO>;
  die "LEFTHOOK_COMPOSER_REPO is not set (git URL of the Packagist mirror repository)"
    unless $repo-url;

  publish-composer-package(
    repo-url    => $repo-url,
    package-dir => PHP,
    binaries    => %PHP-DISTS,
    sys         => $!sys,
  );
}
