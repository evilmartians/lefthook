# Software development should not just bring profit, it also has to be fun.
# Of course Raku isn't the perfect tool for scripting, but it is quite expressive,
# it has types, and it feels like real magic.
#
# I hope that reading through these scripts will show you
# a lot of interesting concepts... and definitely fun!

use System;
use Registry;

use Registries::NPM;
use Registries::RubyGem;
use Registries::PyPI;
use Registries::AUR;
use Registries::AUR-Bin;

my constant packages = (
  Registries::NPM,
  Registries::RubyGem,
  Registries::PyPI,
  Registries::AUR,
  Registries::AUR-Bin,
);

class Packaging {
  has Bool           $.dry-run is required;
  has Registry::Kind $.target  is required;

  method clean {
    .clean for self!packages;
  }

  method set-version {
    .set-version for self!packages;
  }

  method prepare {
    .prepare for self!packages;
  }

  method publish {
    .publish for self!packages;
  }

  method !packages(--> Seq) {
    my @packages = packages;

    unless $!target == Registry::Kind::<all-registries> {
      @packages .= grep(*.kind == $!target)
    }

    my $sys = System.new(dry-run => $!dry-run);
    @packages.map(*.new(sys => $sys));
  }
}
