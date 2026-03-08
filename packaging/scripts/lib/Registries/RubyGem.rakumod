use Constants;
use System;
use Registry;

my constant rubygems = PKG-ROOT.child("rubygems").Str;

class Registries::RubyGem does Registry::Package {
  has System $.sys is required;

  my constant %RUBYGEM-DISTS = (
    amd64-linux   => "{rubygems}/libexec/lefthook-linux-x64/lefthook",
    amd64-windows => "{rubygems}/libexec/lefthook-windows-x64/lefthook.exe",
    amd64-darwin  => "{rubygems}/libexec/lefthook-darwin-x64/lefthook",
    amd64-freebsd => "{rubygems}/libexec/lefthook-freebsd-x64/lefthook",
    amd64-openbsd => "{rubygems}/libexec/lefthook-openbsd-x64/lefthook",

    arm64-linux   => "{rubygems}/libexec/lefthook-linux-arm64/lefthook",
    arm64-windows => "{rubygems}/libexec/lefthook-windows-arm64/lefthook.exe",
    arm64-darwin  => "{rubygems}/libexec/lefthook-darwin-arm64/lefthook",
    arm64-freebsd => "{rubygems}/libexec/lefthook-freebsd-arm64/lefthook",
    arm64-openbsd => "{rubygems}/libexec/lefthook-openbsd-arm64/lefthook",
  );

  submethod kind returns Registry::Kind { Registry::Kind::<rubygem> }

  method clean {
    $!sys.rm("{rubygems}/libexec/".IO.dir.grep(*.d));
    $!sys.rm("{rubygems}/pkg/");
  }

  method set-version {
    $!sys.replace(
      file => "{rubygems}/lefthook.gemspec",
      regex => /$<spec-version>=(spec '.' version \s* '=') .* $/,
      replacement => -> $/ { qq[$<spec-version> "{VERSION}"] },
    );
  }

  method prepare {
    die "rubygems/ setup is not complete" unless %RUBYGEM-DISTS.keys.Set == %DISTS.keys.Set;

    $!sys.cp(.value, %RUBYGEM-DISTS{.key}) for %DISTS.pairs;
  }

  method publish {
    say "Publish lefthook gem";

    $!sys.cd(rubygems);
    run("rake", "build");
    my $last-pkg = "{rubygems}/pkg/".IO.dir.tail;
    $!sys.run("gem push $last-pkg");
  }
}
