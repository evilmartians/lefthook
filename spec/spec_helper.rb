# frozen_string_literal: true

Dir["#{File.dirname(__FILE__)}/support/**/*.rb"].each { |f| require f }

FileStructure.root = File.dirname(__FILE__)

RSpec.configure do |config|
  config.order = :random
  Kernel.srand config.seed

  config.around(:each) do |example|
    FileStructure.have_git
    Dir.chdir(FileStructure.tmp) { example.run }
    FileStructure.clean
  end
end
