#! /usr/bin/env raku

use v6;

use lib $?FILE.IO.parent.child("lib");
use Packager;
use Registry :Target;

sub MAIN(
  Registry::Target :$target = all-registries,
  Bool :$dry-run = False,
) {
  Packager.new(
    target  => $target,
    dry-run => $dry-run,
  ).clean;
}
