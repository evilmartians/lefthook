unit module Registry;

# Supported regitstries.
enum Kind <
  all-registries

  npm
  rubygem
  pypi
  aur
  aur-bin
>;

# Abstract interface for a registry class to implement.
role Package {
  method kind(--> Kind:D)     { ... }
  method clean(--> Nil)       { ... }
  method set-version(--> Nil) { ... }
  method prepare(--> Nil)     { ... }
  method publish(--> Nil)     { ... }
}
