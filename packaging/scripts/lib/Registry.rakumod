unit module Registry;

# Supported regitstries.
enum Kind is export(:kinds) <
  all-registries

  npm
  rubygem
  pypi
  aur
  aur-bin
>;

# Abstract interface for a registry class to implement.
role Package {
  submethod kind returns Kind { ... }

  method clean       { ... }
  method set-version { ... }
  method prepare     { ... }
  method publish     { ... }
}
