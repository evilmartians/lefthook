use Config;

unit module Dists::AUR::Publishing;

# Updates the AUR Git repo with the new version.
sub publish-aur-package(:$name!, :%sha256-urls!, :$path-to-pkgbuild!, :$sys!) is export
{
  my $clone-to = "{$name}-aur";
  my $dest-pkgbuild = "$clone-to/PKGBUILD";

  run("git", "clone", "ssh://aur@aur.archlinux.org/{$name}.git", $clone-to);
  $path-to-pkgbuild.IO.copy($dest-pkgbuild);

  for %sha256-urls.kv -> $template-name, $url {
    my $sha256sum = fetch-sha256($url);

    $sys.replace(
      file => $dest-pkgbuild,
      regex => /'{{ ' $template-name ' }}'/,
      replacement => $sha256sum,
    );
  }

  $sys.cd($clone-to);

  $sys.run("makepkg --printsrcinfo > .SRCINFO");
  $sys.run("makepkg --noconfirm");
  $sys.run("makepkg --install --noconfirm");

  $sys.run("git config user.name 'github-actions[bot]'");
  $sys.run("git config user.email 'github-actions[bot]@users.noreply.github.com'");
  $sys.run("git add PKGBUILD .SRCINFO");
  $sys.run("git commit -m 'release v{VERSION}'");
  $sys.run("git push origin master");
}

# Fetches the binary data by $url and returns SHA256 on it.
sub fetch-sha256(Str:D $url --> Str:D)
{
  my $curl = run("curl", "-fsSL", $url, :out, :bin);
  my $sha256sum = run("sha256sum", "-", :in($curl.out), :out);

  $sha256sum.out.slurp(:close).words.head;
}
