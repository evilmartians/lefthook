# frozen_string_literal: true

require 'spec_helper'
require 'open3'

RSpec.describe 'add command' do
  let(:command) { 'hookah add pre-commit' }
  let(:pre_commit_hook) do
    File.join(FileStructure.tmp, '.git', 'hooks', 'pre-commit')
  end
  let(:hookah_pre_commit_group) do
    File.join(FileStructure.tmp, '.hookah', 'pre-commit')
  end
  let(:expected_pre_commit_hook) { FileStructure.pre_commit_hook_path }

  before do
    FileStructure.make_config
    _, @status = Open3.capture2(command)
  end

  it 'exit with 0 status' do
    expect(@status.success?).to be_truthy
  end

  it 'create pre-commit git hook' do
    expect(File.exist?(pre_commit_hook)).to be_truthy
    expect(
      FileUtils.compare_file(expected_pre_commit_hook, pre_commit_hook)
    ).to be_truthy
  end

  it 'create hookah pre-commit group' do
    expect(Dir.exist?(hookah_pre_commit_group)).to be_truthy
  end
end
