unit module Registry;

# Supported regitstries.
enum Target is export(:Target) <
  all-registries

  npm
  rubygems
  pypi
  aur
  aur-bin
>;

# Abstract interface for a registry class to implement.
role Package {
  method target(--> Target:D)     { ... }
  method clean(--> Nil)       { ... }
  method set-version(--> Nil) { ... }
  method prepare(--> Nil)     { ... }
  method publish(--> Nil)     { ... }
}
