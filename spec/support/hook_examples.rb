require 'spec_helper'

RSpec.shared_examples "hook examples" do 
  let(:pre_push_hook) do
    File.join(FileStructure.tmp, '.git', 'hooks', 'pre-push')
  end

  let(:pre_commit_hook) do
    File.join(FileStructure.tmp, '.git', 'hooks', 'pre-commit')
  end

  let(:expected_pre_push_hook) { FileStructure.pre_push_hook_path }
  let(:expected_pre_commit_hook) { FileStructure.pre_commit_hook_path }

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
end