# frozen_string_literal: true

require 'spec_helper'
require 'open3'

RSpec.describe 'run command' do
  before do
    FileStructure.make_config
    FileStructure.make_scripts_preset
    _, @stderr, @status = Open3.capture3(command)
  end

  describe 'fail chain' do
    let(:command) { 'hookah run pre-commit' }
    let(:expected_output) do
      "\n[  \e[32mOK\e[0m  ] ok_script\n[ \e[31mFAIL\e[0m ] fail_script\n"
    end

    it 'exit with 1 status' do
      expect(@status.success?).to be_falsy
    end

    it 'contain expected output' do
      expect(@stderr).to include(expected_output)
    end
  end

  describe 'ok chain' do
    let(:command) { 'hookah run pre-push' }
    let(:expected_output) { "[  \e[32mOK\e[0m  ] ok_script\n" }

    it 'exit with 0 status' do
      expect(@status.success?).to be_truthy
    end

    it 'contain expected output' do
      expect(@stderr).to include(expected_output)
    end
  end
end
