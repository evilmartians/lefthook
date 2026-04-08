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

    chdir $path unless $!dry-run;
    LEAVE { say "cd $old"; chdir $old unless $!dry-run; } # like defer in Go

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
    say "replace in $file\n\t{$regex.gist} -> {$replacement.gist}";
    return if $!dry-run;

    die "$file does not exist" unless $file.f;

    spurt $file, $file.slurp.lines.map({ .subst($regex, $replacement) }).join("\n") ~ "\n";
  }

  # Runs the command.
  method run(*@argv --> Nil) {
    say "run {@argv.join(' ')}";
    return if $!dry-run;

    my $proc = run(|@argv, :out, :err);

    # Stream stdout and stderr concurrently, printing each line as it arrives.
    # Running both reads in parallel prevents a pipe deadlock: if the child
    # fills one pipe's kernel buffer (~64 KB) while the parent is blocked on
    # the other, neither side can make progress.  Streaming also gives
    # real-time output for long-running commands like `uv build` and `makepkg`.
    my $out-promise = start { for $proc.out.lines(:close) -> $line { say $line } };
    my $err-promise = start { for $proc.err.lines(:close) -> $line { note $line } };

    await $out-promise;
    await $err-promise;

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
