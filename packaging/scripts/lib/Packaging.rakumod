# Software development should not just bring profit, it also has to be fun.
# Of course Raku isn't the perfect tool for scripting, but it is quite expressive,
# it has types, and it feels like real magic.
#
# I hope that reading through these scripts will show you
# a lot of interesting concepts... and definitely fun!

unit class Packaging;

use System;
use Registry;

use Registries::NPM;
use Registries::RubyGems;
use Registries::PyPI;
use Registries::AUR;
use Registries::AUR-Bin;

my constant @PACKAGE-TYPES = (
  Registries::NPM,
  Registries::RubyGems,
  Registries::PyPI,
  Registries::AUR,
  Registries::AUR-Bin,
);

has Bool           $.dry-run is required;
has Registry::Kind $.target  is required;

method clean(--> Nil)       { .clean       for self!packages }
method set-version(--> Nil) { .set-version for self!packages }
method prepare(--> Nil)     { .prepare     for self!packages }
method publish(--> Nil)     { .publish     for self!packages }

method !packages(--> Seq) {
  my $sys = System.new(dry-run => $!dry-run);
  my @packages = @PACKAGE-TYPES.map({ .new(sys => $sys) });

  return @packages.Seq if $!target == Registry::Kind::all-registries;

  @packages.grep(*.kind == $!target);
}
