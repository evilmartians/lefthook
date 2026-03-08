unit module TestRegistry;

use TestSystem;

use Registries::NPM;
use Registries::RubyGem;
use Registries::PyPI;
use Registries::AUR;
use Registries::AUR-Bin;

subset RegistryClass where * ~~ (
  | Registries::NPM
  | Registries::RubyGem
  | Registries::PyPI
  | Registries::AUR
  | Registries::AUR-Bin
);

sub new-registry(RegistryClass $class --> List) is export {
  my $sys = TestSystem.new;
  my $npm = $class.new(:$sys);

  ($sys, $npm);
}
