#! /usr/bin/env raku

use v6;

use lib $?FILE.IO.parent.child("lib");
use Packaging;
use Registry :kinds;

sub MAIN(
  Registry::Kind :$target = all-registries,
  Bool :$dry-run = False,
) {
  Packaging.new(
    target => $target,
    dry-run => $dry-run,
  ).clean;
}
