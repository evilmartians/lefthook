use Constants;

unit module Aur::Publishing;

sub publish-aur-package(:$name!, :%sha256-urls!, :$path-to-pkgbuild!, :$sys!) is export
{
  my $clone-to = "{$name}-aur";
  my $dest-pkgbuild = "$clone-to/PKGBUILD";

  $sys.run("git clone ssh://aur@aur.archlinux.org/{$name}.git {$clone-to}");
  $sys.cp($path-to-pkgbuild, $dest-pkgbuild);

  for %sha256-urls.kv -> $name, $url {
    my $sha256sum = fetch-sha256($url);

    $sys.replace(
      file => $dest-pkgbuild,
      regex => /'{{ ' $name ' }}'/,
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

sub fetch-sha256(Str:D $url --> Str:D)
{
  my $data = run("curl", "-fsSL", $url, :out).out.slurp(:close, :bin);
  run("sha256sum", "-", :in($data), :out).out.slurp(:close).words.head;
}
