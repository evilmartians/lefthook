# Provides wrappers for interaction with file system.
class System {
  has Bool $.dry-run is required;

  # Removes file or dir recursively.
  #
  # `multi` allows calling the method in different ways:
  #   rm("path1", "path2", ...);
  #   rm(["path1", "path2", ...]);
  multi method rm(*@paths) { self.rm(@paths) }
  multi method rm(@paths) {
    for @paths -> $path {
      next unless $path.IO.e;

      say "rm " ~ $path;
      next if $!dry-run;

      self!rm-r($path);
    };
  }

  # Changes current dir.
  method cd(IO() $path) {
    say "cd $path";
    chdir $path;
  }

  # Copies a file. Creates parent dirs for $dest if needed.
  method cp(Str:D $source, Str:D $dest) {
    say "cp $source -> $dest";
    return if $!dry-run;

    mkdir $dest.dirname unless $dest.IO.parent.e;
    $source.IO.copy($dest) unless $!dry-run;
  }

  # Replaces text in a $file line-by-line.
  method replace(IO() :$file, Regex :$regex, :$replacement) {
    die "$file does not exist" unless $file.f;

    say "replace in $file\n\t{$regex.gist} -> {$replacement.gist}";
    return if $!dry-run;

    spurt $file, $file.slurp.lines.map({ .subst($regex, $replacement) }).join("\n") ~ "\n";
  }

  # Runs the command.
  method run(Str:D $cmd) {
    say "run $cmd";
    return if $!dry-run;

    run($cmd.words, :out).out.slurp(:close).chomp.say;
  }

  method !rm-r(IO() $path) {
    return unless $path.e;

    if $path.f {
      $path.unlink;
      return;
    }

    for $path.dir -> $entry {
      self!rm-r($entry);
    }
  }
}
