#! /usr/bin/env raku

# Software development should not just bring profit, it also has to be fun.
# Of course Raku isn't the perfect tool for scripting, but it is quite expressive,
# it has types, and it feels like real magic.
#
# I hope that reading through these scripts will show you
# a lot of interesting concepts... and definitely fun!

use v6;

use lib $*PROGRAM.dirname ~ "/raku-lib";

use Constants;
use System;

use Package;

use Dists::NPM;
use Dists::RubyGem;
use Dists::PyPI;
use Dists::AUR;
use Dists::AUR-Bin;

#| Available steps:
#|   clean
#|   set-version
#|   prepare
#|   publish
sub MAIN(*@steps, Package::Kind :$package = Package::Kind::<all>, Bool :$dry-run = False) {
  my $sys = System.new(dry-run => $dry-run);

  my Package::Dist @packages = (
    Dists::NPM,
    Dists::RubyGem,
    Dists::PyPI,
    Dists::AUR,
    Dists::AUR-Bin,
  );

  @packages .= grep(*.kind == $package) unless $package == Package::Kind::<all>;
  @packages .= map(*.new(sys => $sys));

  for @steps -> $step {
    ."$step"() for @packages;
  }
}
