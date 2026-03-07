#! /usr/bin/env raku

use v6;

use lib $*PROGRAM.dirname ~ "/lib";

use Constants;
use System;

use Lefthook::Package;
use Lefthook::PackageKind;

use NPM;
use RubyGem;
use PyPI;
use AUR;
use AUR-Bin;

#| Available steps:
#|   clean
#|   set-version
#|   prepare
#|   publish
sub MAIN(*@steps, Lefthook::PackageKind :$package = all, Bool :$dry-run = False) {
  my $sys = System.new(dry-run => $dry-run);

  my Lefthook::Package @packages = (
    NPM,
    RubyGem,
    PyPI,
    AUR,
    AUR-Bin,
  );

  @packages .= grep(*.kind == $package) unless $package == all;
  @packages .= map(*.new(sys => $sys));

  for @steps -> $step {
    ."$step"() for @packages;
  }
}
