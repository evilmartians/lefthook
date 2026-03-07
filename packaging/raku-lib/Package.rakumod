unit module Package;

enum Kind is export(:kinds) <
  all-packages

  npm
  rubygem
  pypi
  aur
  aur-bin
>;

role Dist {
  submethod kind returns Kind { ... }

  method clean       { ... }
  method set-version { ... }
  method prepare     { ... }
  method publish     { ... }
}
