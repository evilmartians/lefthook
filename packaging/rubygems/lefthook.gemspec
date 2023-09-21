Gem::Specification.new do |spec|
  spec.name          = "lefthook"
  spec.version       = "1.5.0"
  spec.authors       = ["A.A.Abroskin", "Evil Martians"]
  spec.email         = ["lefthook@evilmartians.com"]

  spec.summary       = "A single dependency-free binary to manage all your git hooks that works with any language in any environment, and in all common team workflows."
  spec.homepage      = "https://github.com/evilmartians/lefthook"
  spec.post_install_message = "Lefthook installed! Run command in your project root directory 'lefthook install -f' to complete installation."

  spec.bindir        = "bin"
  spec.executables   << "lefthook"
  spec.require_paths = ["lib"]

  spec.files = %w(
    lib/lefthook.rb
    bin/lefthook
  ) + `find libexec/ -executable -type f -print0`.split("\x0")

  spec.licenses = ['MIT']
end
