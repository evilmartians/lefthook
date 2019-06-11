# frozen_string_literal: true

require 'spec_helper'
require 'open3'

RSpec.describe 'install command' do
  let(:command) { 'lefthook install' }
  let(:config_name) { 'lefthook.yml' }
  let(:expected_config_path) { FileStructure.config_yaml_path }

  describe 'First time install' do
    before do
      _, @status = Open3.capture2(command)
    end

    it 'exit with 0 status' do
      expect(@status.success?).to be_truthy
    end

    it 'create config file' do
      expect(File.exist?(config_name)).to be_truthy
    end
  end

  describe 'install after cloning repo with existed lefthook structure' do
    let(:pre_push_hook) do
      File.join(FileStructure.tmp, '.git', 'hooks', 'pre-push')
    end
    let(:pre_commit_hook) do
      File.join(FileStructure.tmp, '.git', 'hooks', 'pre-commit')
    end
    let(:expected_pre_push_hook) { FileStructure.pre_push_hook_path }
    let(:expected_pre_commit_hook) { FileStructure.pre_commit_hook_path }

    before do
      FileStructure.make_config
      FileStructure.make_scripts_preset
      _, @status = Open3.capture2(command)
    end

    it 'exit with 0 status' do
      expect(@status.success?).to be_truthy
    end

    it 'create pre-push git hook' do
      expect(File.exist?(pre_push_hook)).to be_truthy
      expect(
        FileUtils.compare_file(expected_pre_push_hook, pre_push_hook)
      ).to be_truthy
    end

    it 'create pre-commit git hook' do
      expect(File.exist?(pre_commit_hook)).to be_truthy
      expect(
        FileUtils.compare_file(expected_pre_commit_hook, pre_commit_hook)
      ).to be_truthy
    end

    it 'rename existed hook with .old extension' do
      expect(File.exist?(pre_push_hook + '.old')).to be_truthy
    end

    it 'command can`t overwrite file with .old extension' do
      _, status = Open3.capture2(command)

      expect(status.success?).to be_falsy
    end
  end
end
