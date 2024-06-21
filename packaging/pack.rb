VERSION = 1.6.17

DIST_DIR = File.join(__dir__, "..", "dist")

def clean
  Dir["**/README.md"]
  Dir["**/lefthook*"]
  system("git clean -fdX npm-installer/ npm-bundled/ npm-bundled/bin/ rubygems/libexec/ rubygems/pkg/")
end

def put_readme
  Dir["npm/*"].each do |npm_dir|
    FileUtils.cp("../README.md", npm_dir)
  end
  FileUtils.cp("../README.md", "npm-bundled/")
  FileUtils.cp("../README.md", "npm-installer/")
end
