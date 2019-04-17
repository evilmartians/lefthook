# frozen_string_literal: true

require 'spec_helper'
require 'open3'

RSpec.describe 'add command' do
  let(:pre_commit_hook) do
    File.join(FileStructure.tmp, '.git', 'hooks', 'pre-commit')
  end
  let(:lefthook_pre_commit_group) do
    File.join(FileStructure.tmp, '.lefthook', 'pre-commit')
  end
  let(:expected_pre_commit_hook) { FileStructure.pre_commit_hook_path }

  before do
    FileStructure.make_config
    _, @status = Open3.capture2(command)
  end

  describe 'with -d flag' do
    let(:command) { 'lefthook add -d pre-commit' }

    it 'exit with 0 status' do
      expect(@status.success?).to be_truthy
    end

    it 'create pre-commit git hook' do
      expect(File.exist?(pre_commit_hook)).to be_truthy
      expect(
        FileUtils.compare_file(expected_pre_commit_hook, pre_commit_hook)
      ).to be_truthy
    end

    it 'create lefthook pre-commit group' do
      expect(Dir.exist?(lefthook_pre_commit_group)).to be_truthy
    end
  end

  describe 'without -d flag' do
    let(:command) { 'lefthook add pre-commit' }

    it 'exit with 0 status' do
      expect(@status.success?).to be_truthy
    end

    it 'create pre-commit git hook' do
      expect(File.exist?(pre_commit_hook)).to be_truthy
      expect(
        FileUtils.compare_file(expected_pre_commit_hook, pre_commit_hook)
      ).to be_truthy
    end

    it 'skip lefthook pre-commit group' do
      expect(Dir.exist?(lefthook_pre_commit_group)).to be_falsy
    end
  end
end
