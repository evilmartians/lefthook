use Constants;
use SystemAPI;

unit module Registries::AUR::Publishing;

# Updates the AUR Git repo with the new version.
sub publish-aur-package(
  Str:D :$name!,
  :%sha256-urls!,
  IO() :$path-to-pkgbuild!,
  SystemAPI :$sys!,
  --> Nil
) is export {
  my $clone-to = PKG-ROOT.child("{$name}-aur");
  my $dest-pkgbuild = $clone-to.child("PKGBUILD");

  $sys.in-dir(PKG-ROOT, {
    clone-aur-repo($sys, $name, $clone-to);
    copy-pkgbuild($sys, $path-to-pkgbuild, $dest-pkgbuild);
    fill-sha256-sums($sys, $dest-pkgbuild, %sha256-urls);
  });

  $sys.in-dir($clone-to, {
    $sys.run("sh", "-c", "makepkg --printsrcinfo > .SRCINFO");
    $sys.run("makepkg", "--noconfirm");
    $sys.run("makepkg", "--install", "--noconfirm");

    $sys.run("git", "config", "user.name", "github-actions[bot]");
    $sys.run("git", "config", "user.email", "github-actions[bot]@users.noreply.github.com");
    $sys.run("git", "add", "PKGBUILD", ".SRCINFO");
    $sys.run("git", "commit", "-m", "release v{VERSION}");
    $sys.run("git", "push", "origin", "master");
  });
}

sub clone-aur-repo(SystemAPI $sys, Str:D $name, IO() $clone-to --> Nil) {
  $sys.run("git", "clone", "ssh://aur@aur.archlinux.org/{$name}.git", $clone-to);
}

sub copy-pkgbuild(SystemAPI $sys, IO() $from, IO() $to --> Nil) {
  $sys.cp($from, $to);
}

sub fill-sha256-sums(
  SystemAPI $sys,
  IO() $pkgbuild,
  %sha256-urls,
  --> Nil
) {
  for %sha256-urls.kv -> $template-name, $url {
    my $sha256sum = fetch-sha256($url);

    $sys.replace(
      file => $pkgbuild,
      regex => /'{{ ' $template-name ' }}'/,
      replacement => $sha256sum,
    );
  }
}

# Fetches the binary data by $url and returns SHA256 on it.
sub fetch-sha256(Str:D $url --> Str:D) {
  say "Fetching SHA256 for $url";

  my $curl = run("curl", "-fsSL", $url, :out, :bin);
  my $sha256sum = run("sha256sum", "-", :in($curl.out), :out);

  $sha256sum.out.slurp(:close).words.head;
}
