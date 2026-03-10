use Registry;

unit class Registries::RubyGems does Registry::Package;

use Constants;
use SystemAPI;

my constant RUBYGEMS = PKG-ROOT.child("rubygems");
my constant %RUBYGEM-DISTS = (
  amd64-linux   => "{RUBYGEMS}/libexec/lefthook-linux-x64/lefthook",
  amd64-windows => "{RUBYGEMS}/libexec/lefthook-windows-x64/lefthook.exe",
  amd64-darwin  => "{RUBYGEMS}/libexec/lefthook-darwin-x64/lefthook",
  amd64-freebsd => "{RUBYGEMS}/libexec/lefthook-freebsd-x64/lefthook",
  amd64-openbsd => "{RUBYGEMS}/libexec/lefthook-openbsd-x64/lefthook",

  arm64-linux   => "{RUBYGEMS}/libexec/lefthook-linux-arm64/lefthook",
  arm64-windows => "{RUBYGEMS}/libexec/lefthook-windows-arm64/lefthook.exe",
  arm64-darwin  => "{RUBYGEMS}/libexec/lefthook-darwin-arm64/lefthook",
  arm64-freebsd => "{RUBYGEMS}/libexec/lefthook-freebsd-arm64/lefthook",
  arm64-openbsd => "{RUBYGEMS}/libexec/lefthook-openbsd-arm64/lefthook",
);

has SystemAPI $.sys is required;

method target(--> Registry::Target:D) { Registry::Target::rubygem }

method clean {
  $!sys.rm("{RUBYGEMS}/libexec/".IO.dir.grep(*.d));
  $!sys.rm("{RUBYGEMS}/pkg/".IO);
}

method set-version {
  $!sys.replace(
    file => "{RUBYGEMS}/lefthook.gemspec",
    regex => /$<spec-version>=(spec '.' version \s* '=') .* $/,
    replacement => -> $/ { qq[$<spec-version> "{VERSION}"] },
  );
}

method prepare {
  die "rubygems/ setup is not complete" unless %RUBYGEM-DISTS.keys.Set == %DISTS.keys.Set;

  for %DISTS.kv -> $platform, $source {
    $!sys.cp($source, %RUBYGEM-DISTS{$platform});
  }
}

method publish {
  say "Publish lefthook gem";

  $!sys.in-dir(RUBYGEMS, {
    $!sys.run("rake", "build");
  });

  my $pkg-dir = RUBYGEMS.child("pkg");
  my $last-pkg = $pkg-dir.IO.dir.sort(*.basename).tail
      // die "no gem found in rubygems/pkg/";

  $!sys.run("gem", "push", $last-pkg);
}
