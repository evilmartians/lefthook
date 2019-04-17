# frozen_string_literal: true

require 'fileutils'

class FileStructure
  class << self
    attr_accessor :root

    def have_git
      FileUtils.mkdir_p(File.join(tmp, '.git', 'hooks'))
    end

    def make_scripts_preset
      FileUtils.mkdir_p(File.join(tmp, '.lefthook', 'pre-commit'))
      FileUtils.cp(
        [ok_script_path, fail_script_path],
        File.join(tmp, '.lefthook', 'pre-commit')
      )

      FileUtils.mkdir_p(File.join(tmp, '.lefthook', 'pre-push'))
      FileUtils.cp(ok_script_path, File.join(tmp, '.lefthook', 'pre-push'))

      FileUtils.cp(pre_push_hook_path, File.join(tmp, '.git', 'hooks'))

      FileUtils.chmod_R 0o777, tmp
    end

    def clean
      FileUtils.remove_dir(tmp)
    end

    def make_config
      FileUtils.cp(config_yaml_path, tmp)
    end

    def tmp
      @tmp ||= File.join(root, 'tmp')
    end

    def config_yaml_path
      File.join(fixtures, 'lefthook.yml')
    end

    def ok_script_path
      File.join(fixtures, 'ok_script')
    end

    def fail_script_path
      File.join(fixtures, 'fail_script')
    end

    def pre_commit_hook_path
      File.join(fixtures, 'pre-commit')
    end

    def pre_push_hook_path
      File.join(fixtures, 'pre-push')
    end

    private

    def fixtures
      File.join(root, 'fixtures')
    end
  end
end
