use Constants;
use SystemAPI;

unit module Registries::PHP::Publishing;

# Updates the PHP Git mirror repo with the new version.
sub publish-composer-package(
  Str:D     :$repo-url!,
  IO()      :$package-dir!,
  :%binaries!,
  SystemAPI :$sys!,
  --> Nil
) is export {
  my $clone-to = PKG-ROOT.child("lefthook-composer");

  $sys.in-dir($package-dir, {
    $sys.run("composer", "validate");
  });

  $sys.rm($clone-to);
  $sys.in-dir(PKG-ROOT, {
    $sys.run("git", "clone", $repo-url, $clone-to);
  });

  $sys.cp("{$package-dir}/composer.json", "{$clone-to}/composer.json");
  $sys.cp("{$package-dir}/bin/lefthook", "{$clone-to}/bin/lefthook");

  for %binaries.values -> $binary {
    $sys.cp($binary, $binary.subst($package-dir.Str, $clone-to.Str));
  }

  $sys.in-dir($clone-to, {
    $sys.run("git", "config", "user.name", "github-actions[bot]");
    $sys.run("git", "config", "user.email", "github-actions[bot]@users.noreply.github.com");
    $sys.run("git", "add", "-A");
    $sys.run("git", "commit", "-m", "release v{VERSION}");
    $sys.run("git", "tag", "v{VERSION}");
    $sys.run("git", "push", "origin", "HEAD", "--tags");
  });
}
