use System;

# Mocks work with filesystem.
class TestSystem does SystemAPI {
  has @.removed;
  has @.replaced;
  has %.copied;
  has @.run-calls = ();
  has $!cwd;

  multi method rm(@paths) {
    @.removed.append(@paths);
  }

  method cd(IO() $path) {
    $!cwd = $path;
  }

  method cp(IO() $source, IO() $dest) {
    %.copied{$source} //= SetHash.new;
    %.copied{$source}.set($dest.Str);
  }

  method replace(IO() :$file, Regex :$regex, :$replacement) {
    ...
  }

  method run(Str:D $cmd) {
    @.run-calls.push(($cmd, $!cwd.clone));
  }
}
