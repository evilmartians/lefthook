#!/usr/bin/env ruby

require "fileutils"

VERSION = "1.7.14"

ROOT = File.join(__dir__, "..")
DIST = File.join(ROOT, "dist")

module Pack
  extend FileUtils

  module_function

  def prepare
    clean
    set_version
    put_readme
    put_binaries
  end

  def clean
    cd(__dir__)
    puts "Cleaning... "
    rm(Dir["npm/**/README.md"])
    rm(Dir["npm/**/lefthook*"].filter(&File.method(:file?)))
    system("git clean -fdX npm-installer/ npm-bundled/ npm-bundled/bin/ rubygems/libexec/ rubygems/pkg/", exception: true)
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
  end

  def put_readme
    cd(__dir__)
    puts "Putting READMEs... "
    Dir["npm/*"].each do |npm_dir|
      cp(File.join(ROOT, "README.md"), File.join(npm_dir, "README.md"), verbose: true)
    end
    cp(File.join(ROOT, "README.md"), "npm-bundled/", verbose: true)
    cp(File.join(ROOT, "README.md"), "npm-installer/", verbose: true)
    puts "done"
  end

  def put_binaries
    cd(__dir__)
    puts "Putting binaries to packages..."
    {
      "#{DIST}/no_self_update_linux_amd64_v1/lefthook"        =>  "npm/lefthook-linux-x64/bin/lefthook",
      "#{DIST}/no_self_update_linux_arm64/lefthook"           =>  "npm/lefthook-linux-arm64/bin/lefthook",
      "#{DIST}/no_self_update_freebsd_amd64_v1/lefthook"      =>  "npm/lefthook-freebsd-x64/bin/lefthook",
      "#{DIST}/no_self_update_freebsd_arm64/lefthook"         =>  "npm/lefthook-freebsd-arm64/bin/lefthook",
      "#{DIST}/no_self_update_openbsd_amd64_v1/lefthook"      =>  "npm/lefthook-openbsd-x64/bin/lefthook",
      "#{DIST}/no_self_update_openbsd_arm64/lefthook"         =>  "npm/lefthook-openbsd-arm64/bin/lefthook",
      "#{DIST}/no_self_update_windows_amd64_v1/lefthook.exe"  =>  "npm/lefthook-windows-x64/bin/lefthook.exe",
      "#{DIST}/no_self_update_windows_arm64/lefthook.exe"     =>  "npm/lefthook-windows-arm64/bin/lefthook.exe",
      "#{DIST}/no_self_update_darwin_amd64_v1/lefthook"       =>  "npm/lefthook-darwin-x64/bin/lefthook",
      "#{DIST}/no_self_update_darwin_arm64/lefthook"          =>  "npm/lefthook-darwin-arm64/bin/lefthook",
    }.each do |(source, dest)|
      mkdir_p(File.dirname(dest))
      cp(source, dest, verbose: true)
    end

    {
      "#{DIST}/no_self_update_linux_amd64_v1/lefthook"        =>  "npm-bundled/bin/lefthook-linux-x64/lefthook",
      "#{DIST}/no_self_update_linux_arm64/lefthook"           =>  "npm-bundled/bin/lefthook-linux-arm64/lefthook",
      "#{DIST}/no_self_update_freebsd_amd64_v1/lefthook"      =>  "npm-bundled/bin/lefthook-freebsd-x64/lefthook",
      "#{DIST}/no_self_update_freebsd_arm64/lefthook"         =>  "npm-bundled/bin/lefthook-freebsd-arm64/lefthook",
      "#{DIST}/no_self_update_openbsd_amd64_v1/lefthook"      =>  "npm-bundled/bin/lefthook-openbsd-x64/lefthook",
      "#{DIST}/no_self_update_openbsd_arm64/lefthook"         =>  "npm-bundled/bin/lefthook-openbsd-arm64/lefthook",
      "#{DIST}/no_self_update_windows_amd64_v1/lefthook.exe"  =>  "npm-bundled/bin/lefthook-windows-x64/lefthook.exe",
      "#{DIST}/no_self_update_windows_arm64/lefthook.exe"     =>  "npm-bundled/bin/lefthook-windows-arm64/lefthook.exe",
      "#{DIST}/no_self_update_darwin_amd64_v1/lefthook"       =>  "npm-bundled/bin/lefthook-darwin-x64/lefthook",
      "#{DIST}/no_self_update_darwin_arm64/lefthook"          =>  "npm-bundled/bin/lefthook-darwin-arm64/lefthook",
    }.each do |(source, dest)|
      mkdir_p(File.dirname(dest))
      cp(source, dest, verbose: true)
    end

    {
      "#{DIST}/no_self_update_linux_amd64_v1/lefthook"        =>  "rubygems/libexec/lefthook-linux-x64/lefthook",
      "#{DIST}/no_self_update_linux_arm64/lefthook"           =>  "rubygems/libexec/lefthook-linux-arm64/lefthook",
      "#{DIST}/no_self_update_freebsd_amd64_v1/lefthook"      =>  "rubygems/libexec/lefthook-freebsd-x64/lefthook",
      "#{DIST}/no_self_update_freebsd_arm64/lefthook"         =>  "rubygems/libexec/lefthook-freebsd-arm64/lefthook",
      "#{DIST}/no_self_update_openbsd_amd64_v1/lefthook"      =>  "rubygems/libexec/lefthook-openbsd-x64/lefthook",
      "#{DIST}/no_self_update_openbsd_arm64/lefthook"         =>  "rubygems/libexec/lefthook-openbsd-arm64/lefthook",
      "#{DIST}/no_self_update_windows_amd64_v1/lefthook.exe"  =>  "rubygems/libexec/lefthook-windows-x64/lefthook.exe",
      "#{DIST}/no_self_update_windows_arm64/lefthook.exe"     =>  "rubygems/libexec/lefthook-windows-arm64/lefthook.exe",
      "#{DIST}/no_self_update_darwin_amd64_v1/lefthook"       =>  "rubygems/libexec/lefthook-darwin-x64/lefthook",
      "#{DIST}/no_self_update_darwin_arm64/lefthook"          =>  "rubygems/libexec/lefthook-darwin-arm64/lefthook",
    }.each do |(source, dest)|
      mkdir_p(File.dirname(dest))
      cp(source, dest, verbose: true)
    end

    puts "done"
  end

  def publish
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

    puts "Publishing lefthook gem..."
    cd(File.join(__dir__, "rubygems"))
    system("rake build", exception: true)
    system("gem push pkg/*.gem", exception: true)

    puts "done"
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
