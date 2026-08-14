unit role SystemAPI;

multi method rm(*@paths --> Nil) { self.rm(@paths) }
multi method rm(@paths --> Nil) { ... }

method in-dir(IO() $path, &block --> Nil) { ... }

method cp(IO() $source, IO() $dest --> Nil) { ... }

method replace(IO() :$file, Regex :$regex, :$replacement --> Nil) { ... }

method run(*@argv --> Nil) {... }
