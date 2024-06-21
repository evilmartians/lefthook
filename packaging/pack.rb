#!/usr/bin/env ruby

require "fileutils"

VERSION = "1.6.17"

ROOT = File.join(__dir__, "..")
DIST = File.join(ROOT, "dist")

module Pack
  module_function

  def clean
    FileUtils.cd(__dir__)
    print "Cleaning... "
    FileUtils.rm(Dir["npm/**/README.md"])
    FileUtils.rm(Dir["npm/**/lefthook*"].filter(&File.method(:file?)))
    system("git clean -fdX npm-installer/ npm-bundled/ npm-bundled/bin/ rubygems/libexec/ rubygems/pkg/")
    puts "done"
  end

  def put_readme
    FileUtils.cd(__dir__)
    print "Putting READMEs... "
    Dir["npm/*"].each do |npm_dir|
      FileUtils.cp(File.join(ROOT, "README.md"), File.join(npm_dir, "README.md"), verbose: true)
    end
    FileUtils.cp(File.join(ROOT, "README.md"), "npm-bundled/", verbose: true)
    FileUtils.cp(File.join(ROOT, "README.md"), "npm-installer/", verbose: true)
    puts "done"
  end
end

Pack.public_send ARGV[0]
