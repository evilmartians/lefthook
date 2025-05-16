#!/usr/bin/env ruby

require "fileutils"
require "digest"
require "open-uri"

VERSION = "1.11.13"

ROOT = File.join(__dir__, "..")
DIST = File.join(ROOT, "dist")

module Pack
  extend FileUtils

  module_function

  def prepare
    clean
    set_version
    put_additional_files
    put_binaries
  end

  def clean
    cd(__dir__)
    puts "Cleaning... "
    rm(Dir["npm/**/README.md"])
    rm(Dir["npm/**/lefthook*"].filter(&File.method(:file?)))
    system("git clean -fdX npm-installer/ npm-bundled/ npm-bundled/bin/ rubygems/libexec/ rubygems/pkg/ pypi pypi/lefthook/", exception: true)
    puts "done"
  end

  def set_version
    cd(__dir__)
    puts "Replacing version to #{VERSION} in packages"
    (Dir["npm/**/package.json"] + ["npm-bundled/package.json", "npm-installer/package.json"]).each do |package_json|
      replace_in_file(package_json, /"version": "[\d.]+"/, %{"version": "#{VERSION}"})
    end

    replace_in_file("npm/lefthook/package.json", /"(lefthook-.+)": "[\d.]+"/, %{"\\1": "#{VERSION}"})
    replace_in_file("rubygems/lefthook.gemspec", /(spec\.version\s+= ).*/, %{\\1"#{VERSION}"})
    replace_in_file("pypi/setup.py", /(version+=).*/, %{\\1'#{VERSION}',})
    replace_in_file("aur/lefthook/PKGBUILD", /(pkgver+=).*/, %{\\1#{VERSION}})
    replace_in_file("aur/lefthook-bin/PKGBUILD", /(pkgver+=).*/, %{\\1#{VERSION}})
  end

  def put_additional_files
    cd(__dir__)
    puts "Putting README... "
    Dir["npm/*"].each do |npm_dir|
      cp(File.join(ROOT, "README.md"), File.join(npm_dir, "README.md"), verbose: true)
    end
    cp(File.join(ROOT, "README.md"), "npm-bundled/", verbose: true)
    cp(File.join(ROOT, "README.md"), "npm-installer/", verbose: true)

    puts "Putting schema.json..."
    cp(File.join(ROOT, "schema.json"), "npm/lefthook/", verbose: true)
    cp(File.join(ROOT, "schema.json"), "npm-bundled/", verbose: true)
    cp(File.join(ROOT, "schema.json"), "npm-installer/", verbose: true)
    puts "done"
  end

  def put_binaries
    cd(__dir__)
    puts "Putting binaries to packages..."
    {
      "#{DIST}/no_self_update_linux_amd64_v1/lefthook"        =>  "npm/lefthook-linux-x64/bin/lefthook",
      "#{DIST}/no_self_update_linux_arm64_v8.0/lefthook"           =>  "npm/lefthook-linux-arm64/bin/lefthook",
      "#{DIST}/no_self_update_freebsd_amd64_v1/lefthook"      =>  "npm/lefthook-freebsd-x64/bin/lefthook",
      "#{DIST}/no_self_update_freebsd_arm64_v8.0/lefthook"         =>  "npm/lefthook-freebsd-arm64/bin/lefthook",
      "#{DIST}/no_self_update_openbsd_amd64_v1/lefthook"      =>  "npm/lefthook-openbsd-x64/bin/lefthook",
      "#{DIST}/no_self_update_openbsd_arm64_v8.0/lefthook"         =>  "npm/lefthook-openbsd-arm64/bin/lefthook",
      "#{DIST}/no_self_update_windows_amd64_v1/lefthook.exe"  =>  "npm/lefthook-windows-x64/bin/lefthook.exe",
      "#{DIST}/no_self_update_windows_arm64_v8.0/lefthook.exe"     =>  "npm/lefthook-windows-arm64/bin/lefthook.exe",
      "#{DIST}/no_self_update_darwin_amd64_v1/lefthook"       =>  "npm/lefthook-darwin-x64/bin/lefthook",
      "#{DIST}/no_self_update_darwin_arm64_v8.0/lefthook"          =>  "npm/lefthook-darwin-arm64/bin/lefthook",
    }.each do |(source, dest)|
      mkdir_p(File.dirname(dest))
      cp(source, dest, verbose: true)
    end

    {
      "#{DIST}/no_self_update_linux_amd64_v1/lefthook"         =>  "npm-bundled/bin/lefthook-linux-x64/lefthook",
      "#{DIST}/no_self_update_linux_arm64_v8.0/lefthook"       =>  "npm-bundled/bin/lefthook-linux-arm64/lefthook",
      "#{DIST}/no_self_update_freebsd_amd64_v1/lefthook"       =>  "npm-bundled/bin/lefthook-freebsd-x64/lefthook",
      "#{DIST}/no_self_update_freebsd_arm64_v8.0/lefthook"     =>  "npm-bundled/bin/lefthook-freebsd-arm64/lefthook",
      "#{DIST}/no_self_update_openbsd_amd64_v1/lefthook"       =>  "npm-bundled/bin/lefthook-openbsd-x64/lefthook",
      "#{DIST}/no_self_update_openbsd_arm64_v8.0/lefthook"     =>  "npm-bundled/bin/lefthook-openbsd-arm64/lefthook",
      "#{DIST}/no_self_update_windows_amd64_v1/lefthook.exe"   =>  "npm-bundled/bin/lefthook-windows-x64/lefthook.exe",
      "#{DIST}/no_self_update_windows_arm64_v8.0/lefthook.exe" =>  "npm-bundled/bin/lefthook-windows-arm64/lefthook.exe",
      "#{DIST}/no_self_update_darwin_amd64_v1/lefthook"        =>  "npm-bundled/bin/lefthook-darwin-x64/lefthook",
      "#{DIST}/no_self_update_darwin_arm64_v8.0/lefthook"      =>  "npm-bundled/bin/lefthook-darwin-arm64/lefthook",
    }.each do |(source, dest)|
      mkdir_p(File.dirname(dest))
      cp(source, dest, verbose: true)
    end

    {
      "#{DIST}/no_self_update_linux_amd64_v1/lefthook"         =>  "rubygems/libexec/lefthook-linux-x64/lefthook",
      "#{DIST}/no_self_update_linux_arm64_v8.0/lefthook"       =>  "rubygems/libexec/lefthook-linux-arm64/lefthook",
      "#{DIST}/no_self_update_freebsd_amd64_v1/lefthook"       =>  "rubygems/libexec/lefthook-freebsd-x64/lefthook",
      "#{DIST}/no_self_update_freebsd_arm64_v8.0/lefthook"     =>  "rubygems/libexec/lefthook-freebsd-arm64/lefthook",
      "#{DIST}/no_self_update_openbsd_amd64_v1/lefthook"       =>  "rubygems/libexec/lefthook-openbsd-x64/lefthook",
      "#{DIST}/no_self_update_openbsd_arm64_v8.0/lefthook"     =>  "rubygems/libexec/lefthook-openbsd-arm64/lefthook",
      "#{DIST}/no_self_update_windows_amd64_v1/lefthook.exe"   =>  "rubygems/libexec/lefthook-windows-x64/lefthook.exe",
      "#{DIST}/no_self_update_windows_arm64_v8.0/lefthook.exe" =>  "rubygems/libexec/lefthook-windows-arm64/lefthook.exe",
      "#{DIST}/no_self_update_darwin_amd64_v1/lefthook"        =>  "rubygems/libexec/lefthook-darwin-x64/lefthook",
      "#{DIST}/no_self_update_darwin_arm64_v8.0/lefthook"      =>  "rubygems/libexec/lefthook-darwin-arm64/lefthook",
    }.each do |(source, dest)|
      mkdir_p(File.dirname(dest))
      cp(source, dest, verbose: true)
    end

    {
      "#{DIST}/no_self_update_linux_amd64_v1/lefthook"         =>  "pypi/lefthook/bin/lefthook-linux-x86_64/lefthook",
      "#{DIST}/no_self_update_linux_arm64_v8.0/lefthook"       =>  "pypi/lefthook/bin/lefthook-linux-arm64/lefthook",
      "#{DIST}/no_self_update_freebsd_amd64_v1/lefthook"       =>  "pypi/lefthook/bin/lefthook-freebsd-x86_64/lefthook",
      "#{DIST}/no_self_update_freebsd_arm64_v8.0/lefthook"     =>  "pypi/lefthook/bin/lefthook-freebsd-arm64/lefthook",
      "#{DIST}/no_self_update_openbsd_amd64_v1/lefthook"       =>  "pypi/lefthook/bin/lefthook-openbsd-x86_64/lefthook",
      "#{DIST}/no_self_update_openbsd_arm64_v8.0/lefthook"     =>  "pypi/lefthook/bin/lefthook-openbsd-arm64/lefthook",
      "#{DIST}/no_self_update_windows_amd64_v1/lefthook.exe"   =>  "pypi/lefthook/bin/lefthook-windows-x86_64/lefthook.exe",
      "#{DIST}/no_self_update_windows_arm64_v8.0/lefthook.exe" =>  "pypi/lefthook/bin/lefthook-windows-arm64/lefthook.exe",
      "#{DIST}/no_self_update_darwin_amd64_v1/lefthook"        =>  "pypi/lefthook/bin/lefthook-darwin-x86_64/lefthook",
      "#{DIST}/no_self_update_darwin_arm64_v8.0/lefthook"      =>  "pypi/lefthook/bin/lefthook-darwin-arm64/lefthook",
    }.each do |(source, dest)|
      mkdir_p(File.dirname(dest))
      cp(source, dest, verbose: true)
    end

    puts "done"
  end

  def publish
    publish_pypi
    publish_npm
    publish_gem
  end

  def publish_npm
    puts "Publishing lefthook npm..."
    cd(File.join(__dir__, "npm"))
    Dir["lefthook*"].each do |package|
      puts "publishing #{package}"
      cd(File.join(__dir__, "npm", package))
      system("npm publish --access public", exception: true)
      cd(File.join(__dir__, "npm"))
    end

    puts "Publishing @evilmartians/lefthook npm..."
    cd(File.join(__dir__, "npm-bundled"))
    system("npm publish --access public", exception: true)

    puts "Publishing @evilmartians/lefthook-installer npm..."
    cd(File.join(__dir__, "npm-installer"))
    system("npm publish --access public", exception: true)
  end

  def publish_gem
    puts "Publishing to Rubygems..."
    cd(File.join(__dir__, "rubygems"))
    system("rake build", exception: true)
    system("gem push pkg/*.gem", exception: true)
  end

  def publish_pypi
    puts "Publishing to PyPI..."
    cd(File.join(__dir__, "pypi"))
    system("python setup.py sdist bdist_wheel", exception: true)
    system("python -m twine upload --verbose --repository lefthook dist/*", exception: true)
  end

  def publish_aur_lefthook
    publish_aur("lefthook", {
      sha256sum: "https://github.com/evilmartians/lefthook/archive/v#{VERSION}.tar.gz"
    })
  end

  def publish_aur_lefthook_bin
    publish_aur("lefthook-bin", {
      sha256sum_linux_x86_64: "https://github.com/evilmartians/lefthook/releases/download/v#{VERSION}/lefthook_#{VERSION}_Linux_x86_64.gz",
      sha256sum_linux_aarch64: "https://github.com/evilmartians/lefthook/releases/download/v#{VERSION}/lefthook_#{VERSION}_Linux_aarch64.gz"
    })
  end

  def publish_aur(package_name, sha256urls = {})
    aur_repo = File.join(__dir__, "#{package_name}-aur")
    system("git clone ssh://aur@aur.archlinux.org/#{package_name}.git #{aur_repo}", exception: true)
    pkgbuild_source = File.join(__dir__, "aur", package_name, "PKGBUILD")
    pkgbuild_dest = File.join(aur_repo, "PKGBUILD")
    cp(pkgbuild_source, pkgbuild_dest, verbose: true)

    sha256sums = {}
    sha256urls.each do |name, url|
      sha256 = Digest::SHA256.new
      URI.open(url) do |file|
        while chunk = file.read(1024)  # Read the file in chunks
          sha256.update(chunk)
        end
      end

      sha256sums[name] = sha256.hexdigest
    end

    sha256sums.each do |name, sha256sum|
      replace_in_file(pkgbuild_dest, /{{ #{name} }}/, sha256sum)
    end

    cd(aur_repo)
    system("makepkg --printsrcinfo > .SRCINFO", exception: true)
    system("makepkg --noconfirm", exception: true)
    system("makepkg --install --noconfirm", exception: true)

    system("git config user.name 'github-actions[bot]'", exception: true)
    system("git config user.email 'github-actions[bot]@users.noreply.github.com'", exception: true)
    system("git add PKGBUILD .SRCINFO", exception: true)
    system("git commit -m 'release v#{VERSION}'", exception: true)
    system("git push origin master", exception: true)
  end

  def replace_in_file(filepath, regexp, value)
    text = File.open(filepath, "r") do |f|
      f.read
    end
    text.gsub!(regexp, value)
    File.open(filepath, "w") do |f|
      f.write(text)
    end
  end
end

ARGV.each do |cmd|
  Pack.public_send(cmd)
end
