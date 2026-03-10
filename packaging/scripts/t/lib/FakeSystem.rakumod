use SystemAPI;

# Mocks work with filesystem.
class FakeSystem does SystemAPI {
  has Str @.removed;
  has %.copied;
  has @.run-calls = ();
  has $!cwd;

  multi method rm(@paths --> Nil) {
    @.removed.append(@paths.map(*.Str));
  }

  method in-dir(IO() $path, &block --> Nil) {
    $!cwd = $path;
    block();
  }

  method cp(IO() $source, IO() $dest --> Nil) {
    %.copied{$source} //= SetHash.new;
    %.copied{$source}.set($dest.Str);
  }

  method replace(IO() :$file, Regex :$regex, :$replacement --> Nil) {
    ...
  }

  method run(*@argv --> Nil) {
    @.run-calls.push((@argv.join(' '), $!cwd.clone));
  }
}
