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
    before do
      FileStructure.make_config
      FileStructure.make_scripts_preset
      _, @status = Open3.capture2(command)
    end

    include_examples 'hook examples'
    it 'exit with 0 status' do
      expect(@status.success?).to be_truthy
    end

    it 'dont rename existed lefthook hook with .old extension' do
      expect(File.exist?(pre_push_hook + '.old')).to be_falsy
    end
  end
end
