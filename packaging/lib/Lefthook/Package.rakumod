use Lefthook::PackageKind;

role Lefthook::Package {
  submethod kind returns Lefthook::PackageKind { ... }

  method clean       { ... }
  method set-version { ... }
  method prepare     { ... }
  method publish     { ... }
}
