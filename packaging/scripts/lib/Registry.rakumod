unit module Registry;

enum Kind is export(:kinds) <
  all-registries

  npm
  rubygem
  pypi
  aur
  aur-bin
>;

role Package {
  submethod kind returns Kind { ... }

  method clean       { ... }
  method set-version { ... }
  method prepare     { ... }
  method publish     { ... }
}
