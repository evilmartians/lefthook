use Constants;
use System;
use Registry;

my constant npm           = PKG-ROOT.child("npm");
my constant npm-bundled   = PKG-ROOT.child("npm-bundled");
my constant npm-installer = PKG-ROOT.child("npm-installer");

my constant @READMEs = qq:to/END/.lines.map(*.trim);
  {npm}/lefthook/README.md
  {npm}/lefthook-darwin-arm64/README.md
  {npm}/lefthook-darwin-x64/README.md
  {npm}/lefthook-linux-arm64/README.md
  {npm}/lefthook-linux-x64/README.md
  {npm}/lefthook-windows-arm64/README.md
  {npm}/lefthook-windows-x64/README.md
  {npm}/lefthook-freebsd-arm64/README.md
  {npm}/lefthook-freebsd-x64/README.md
  {npm}/lefthook-openbsd-arm64/README.md
  {npm}/lefthook-openbsd-x64/README.md
  {npm-bundled}/README.md
  {npm-installer}/README.md
END
my constant @packages = qq:to/END/.lines.map(*.trim);
  {npm}/lefthook-darwin-arm64/
  {npm}/lefthook-darwin-x64/
  {npm}/lefthook-linux-arm64/
  {npm}/lefthook-linux-x64/
  {npm}/lefthook-windows-arm64/
  {npm}/lefthook-windows-x64/
  {npm}/lefthook-freebsd-arm64/
  {npm}/lefthook-freebsd-x64/
  {npm}/lefthook-openbsd-arm64/
  {npm}/lefthook-openbsd-x64/
  {npm}/lefthook/
  {npm-bundled}
  {npm-installer}
END
my constant @package-jsons = @packages.map(*.IO.child("package.json"));
my constant @schemas = qq:to/END/.lines.map(*.trim);
  {npm}/lefthook/schema.json
  {npm-bundled}/schema.json
  {npm-installer}/schema.json
END

class Registries::NPM does Registry::Package {
  has System $.sys is required;

  my constant %NPM-DISTS = (
    amd64-linux   => "{npm}/lefthook-linux-x64/bin/lefthook",
    amd64-windows => "{npm}/lefthook-windows-x64/bin/lefthook.exe",
    amd64-darwin  => "{npm}/lefthook-darwin-x64/bin/lefthook",
    amd64-freebsd => "{npm}/lefthook-freebsd-x64/bin/lefthook",
    amd64-openbsd => "{npm}/lefthook-openbsd-x64/bin/lefthook",

    arm64-linux   => "{npm}/lefthook-linux-arm64/bin/lefthook",
    arm64-windows => "{npm}/lefthook-windows-arm64/bin/lefthook.exe",
    arm64-darwin  => "{npm}/lefthook-darwin-arm64/bin/lefthook",
    arm64-freebsd => "{npm}/lefthook-freebsd-arm64/bin/lefthook",
    arm64-openbsd => "{npm}/lefthook-openbsd-arm64/bin/lefthook",
  );
  my constant %NPM-BUNDLED-DISTS = (
    amd64-linux   => "{npm-bundled}/bin/lefthook-linux-x64/lefthook",
    amd64-windows => "{npm-bundled}/bin/lefthook-windows-x64/lefthook.exe",
    amd64-darwin  => "{npm-bundled}/bin/lefthook-darwin-x64/lefthook",
    amd64-freebsd => "{npm-bundled}/bin/lefthook-freebsd-x64/lefthook",
    amd64-openbsd => "{npm-bundled}/bin/lefthook-openbsd-x64/lefthook",

    arm64-linux   => "{npm-bundled}/bin/lefthook-linux-arm64/lefthook",
    arm64-windows => "{npm-bundled}/bin/lefthook-windows-arm64/lefthook.exe",
    arm64-darwin  => "{npm-bundled}/bin/lefthook-darwin-arm64/lefthook",
    arm64-freebsd => "{npm-bundled}/bin/lefthook-freebsd-arm64/lefthook",
    arm64-openbsd => "{npm-bundled}/bin/lefthook-openbsd-arm64/lefthook",
  );

  submethod kind returns Registry::Kind { Registry::Kind::<npm> }

  method clean {
    $!sys.rm(
      |@READMEs,
      |@schemas,
      |%NPM-DISTS.values,
      |%NPM-BUNDLED-DISTS.values,
    )
  }

  method set-version {
    for @package-jsons -> $path {
      $!sys.replace(
        file => $path,
        regex => /'"version":' \s* '"' <[\d\w.]>+ '"'/,
        replacement => qq["version": "{VERSION}"],
      );
    }

    # Update optional dependencies for the main lefthook package
    $!sys.replace(
      file => "{npm}/lefthook/package.json",
      regex => /'"' $<package>=(lefthook '-' <[\d\w-]>+) '":' \s* '"' <[\d\w.]>+ '"'/,
      replacement => -> $/ { qq["$<package>": "{VERSION}"] },
    );
  }

  method prepare {
    $!sys.cp("{REPO-ROOT}/README.md", $_) for @READMEs;
    $!sys.cp("{REPO-ROOT}/schema.json", $_) for @schemas;

    die "npm/ setup is not complete" unless %DISTS.keys.Set == %NPM-DISTS.keys.Set;
    die "npm-bundled/ setup is not complete" unless %DISTS.keys.Set == %NPM-BUNDLED-DISTS.keys.Set;

    for %DISTS.kv -> $kind, $source {
      $!sys.cp($source, %NPM-DISTS{$kind});
      $!sys.cp($source, %NPM-BUNDLED-DISTS{$kind});
    }
  }

  method publish {
    for @packages -> $package {
      say "Publish {$package.IO.basename}";

      $!sys.cd($package);
      $!sys.run("npm publish --access public");
    }
  }
}
