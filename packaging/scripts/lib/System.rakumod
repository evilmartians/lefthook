use SystemAPI;

# Provides wrappers for interaction with file system.
class System does SystemAPI {
  has Bool $.dry-run is required;

  # Removes file or dir recursively.
  multi method rm(@paths --> Nil) {
    for @paths -> $path {
      next unless $path.IO.e;

      say "rm " ~ $path;
      next if $!dry-run;

      self!rm-r($path);
    };
  }

  # Changes current dir and execute the &block.
  method in-dir(IO() $path, &block --> Nil) {
    my $old = $*CWD;

    say "cd $path";
    chdir $path;
    LEAVE chdir $old;

    block();
  }

  # Copies a file. Creates parent dirs for $dest if needed.
  method cp(IO() $source, IO() $dest --> Nil) {
    say "cp $source -> $dest";
    return if $!dry-run;

    mkdir $dest.dirname unless $dest.IO.parent.e;
    $source.IO.copy($dest) unless $!dry-run;
  }

  # Replaces text in a $file line-by-line.
  method replace(IO() :$file, Regex :$regex, :$replacement --> Nil) {
    die "$file does not exist" unless $file.f;

    say "replace in $file\n\t{$regex.gist} -> {$replacement.gist}";
    return if $!dry-run;

    spurt $file, $file.slurp.lines.map({ .subst($regex, $replacement) }).join("\n") ~ "\n";
  }

  # Runs the command.
  method run(*@argv --> Nil) {
    say "run {@argv.join(' ')}";
    return if $!dry-run;

    my $proc = run(|@argv, :out, :err);
    my $out = $proc.out.slurp(:close);
    my $err = $proc.err.slurp(:close);

    print $out if $out.chars;
    note $err if $err.chars;

    die "failed: {@argv.join(' ')} --> {$proc.exitcode}" if $proc.exitcode != 0;
  }

  method !rm-r(IO() $path --> Nil) {
    return unless $path.e;

    if $path.f {
      $path.unlink;
      return;
    }

    die "not a file/dir: $path" unless $path.d;

    for $path.dir -> $entry {
      self!rm-r($entry);
    }

    $path.rmdir;
  }
}
