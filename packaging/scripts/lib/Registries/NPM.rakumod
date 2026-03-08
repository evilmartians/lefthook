use Registry;

unit class Registries::NPM does Registry::Package;

use Constants;
use System;
use SystemAPI;

my constant NPM           = PKG-ROOT.child("npm");
my constant NPM-BUNDLED   = PKG-ROOT.child("npm-bundled");
my constant NPM-INSTALLER = PKG-ROOT.child("npm-installer");

my constant @READMES = qq:to/END/.lines.map(*.trim);
  {NPM}/lefthook/README.md
  {NPM}/lefthook-darwin-arm64/README.md
  {NPM}/lefthook-darwin-x64/README.md
  {NPM}/lefthook-linux-arm64/README.md
  {NPM}/lefthook-linux-x64/README.md
  {NPM}/lefthook-windows-arm64/README.md
  {NPM}/lefthook-windows-x64/README.md
  {NPM}/lefthook-freebsd-arm64/README.md
  {NPM}/lefthook-freebsd-x64/README.md
  {NPM}/lefthook-openbsd-arm64/README.md
  {NPM}/lefthook-openbsd-x64/README.md
  {NPM-BUNDLED}/README.md
  {NPM-INSTALLER}/README.md
END
my constant @PACKAGES = qq:to/END/.lines.map(*.trim);
  {NPM}/lefthook-darwin-arm64/
  {NPM}/lefthook-darwin-x64/
  {NPM}/lefthook-linux-arm64/
  {NPM}/lefthook-linux-x64/
  {NPM}/lefthook-windows-arm64/
  {NPM}/lefthook-windows-x64/
  {NPM}/lefthook-freebsd-arm64/
  {NPM}/lefthook-freebsd-x64/
  {NPM}/lefthook-openbsd-arm64/
  {NPM}/lefthook-openbsd-x64/
  {NPM}/lefthook/
  {NPM-BUNDLED}
  {NPM-INSTALLER}
END
my constant @PACKAGE-JSONS = @PACKAGES.map(*.IO.child("package.json"));
my constant @SCHEMAS = qq:to/END/.lines.map(*.trim);
  {NPM}/lefthook/schema.json
  {NPM-BUNDLED}/schema.json
  {NPM-INSTALLER}/schema.json
END

has SystemAPI $.sys is required;

my constant %NPM-DISTS = (
  amd64-linux   => "{NPM}/lefthook-linux-x64/bin/lefthook",
  amd64-windows => "{NPM}/lefthook-windows-x64/bin/lefthook.exe",
  amd64-darwin  => "{NPM}/lefthook-darwin-x64/bin/lefthook",
  amd64-freebsd => "{NPM}/lefthook-freebsd-x64/bin/lefthook",
  amd64-openbsd => "{NPM}/lefthook-openbsd-x64/bin/lefthook",

  arm64-linux   => "{NPM}/lefthook-linux-arm64/bin/lefthook",
  arm64-windows => "{NPM}/lefthook-windows-arm64/bin/lefthook.exe",
  arm64-darwin  => "{NPM}/lefthook-darwin-arm64/bin/lefthook",
  arm64-freebsd => "{NPM}/lefthook-freebsd-arm64/bin/lefthook",
  arm64-openbsd => "{NPM}/lefthook-openbsd-arm64/bin/lefthook",
);
my constant %NPM-BUNDLED-DISTS = (
  amd64-linux   => "{NPM-BUNDLED}/bin/lefthook-linux-x64/lefthook",
  amd64-windows => "{NPM-BUNDLED}/bin/lefthook-windows-x64/lefthook.exe",
  amd64-darwin  => "{NPM-BUNDLED}/bin/lefthook-darwin-x64/lefthook",
  amd64-freebsd => "{NPM-BUNDLED}/bin/lefthook-freebsd-x64/lefthook",
  amd64-openbsd => "{NPM-BUNDLED}/bin/lefthook-openbsd-x64/lefthook",

  arm64-linux   => "{NPM-BUNDLED}/bin/lefthook-linux-arm64/lefthook",
  arm64-windows => "{NPM-BUNDLED}/bin/lefthook-windows-arm64/lefthook.exe",
  arm64-darwin  => "{NPM-BUNDLED}/bin/lefthook-darwin-arm64/lefthook",
  arm64-freebsd => "{NPM-BUNDLED}/bin/lefthook-freebsd-arm64/lefthook",
  arm64-openbsd => "{NPM-BUNDLED}/bin/lefthook-openbsd-arm64/lefthook",
);

method kind(--> Registry::Kind:D) { Registry::Kind::npm }

method clean {
  $!sys.rm(
    |@READMES,
    |@SCHEMAS,
    |%NPM-DISTS.values,
    |%NPM-BUNDLED-DISTS.values,
  )
}

method set-version {
  for @PACKAGE-JSONS -> $path {
    $!sys.replace(
      file => $path,
      regex => /'"version":' \s* '"' <[\d\w.]>+ '"'/,
      replacement => qq["version": "{VERSION}"],
    );
  }

  # Update optional dependencies for the main lefthook package
  $!sys.replace(
    file => "{NPM}/lefthook/package.json",
    regex => /'"' $<package>=(lefthook '-' <[\d\w-]>+) '":' \s* '"' <[\d\w.]>+ '"'/,
    replacement => -> $/ { qq["$<package>": "{VERSION}"] },
  );
}

method prepare {
  $!sys.cp("{REPO-ROOT}/README.md", $_) for @READMES;
  $!sys.cp("{REPO-ROOT}/schema.json", $_) for @SCHEMAS;

  die "npm/ setup is not complete" unless %DISTS.keys.Set == %NPM-DISTS.keys.Set;
  die "NPM-BUNDLED/ setup is not complete" unless %DISTS.keys.Set == %NPM-BUNDLED-DISTS.keys.Set;

  for %DISTS.kv -> $kind, $source {
    $!sys.cp($source, %NPM-DISTS{$kind});
    $!sys.cp($source, %NPM-BUNDLED-DISTS{$kind});
  }
}

method publish {
  for @PACKAGES -> $package {
    say "Publish {$package.IO.basename}";

    $!sys.in-dir($package, {
      $!sys.run("npm", "publish", "--access", "public");
    });
  }
}
