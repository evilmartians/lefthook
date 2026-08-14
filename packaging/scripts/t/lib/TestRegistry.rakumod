unit module TestRegistry;

use FakeSystem;

use Registries::NPM;
use Registries::RubyGems;
use Registries::PyPI;
use Registries::AUR;
use Registries::AUR-Bin;

subset RegistryClass where * ~~ (
  | Registries::NPM
  | Registries::RubyGems
  | Registries::PyPI
  | Registries::AUR
  | Registries::AUR-Bin
);

sub new-registry(RegistryClass $class --> List) is export {
  my $sys = FakeSystem.new;
  my $npm = $class.new(:$sys);

  ($sys, $npm);
}
