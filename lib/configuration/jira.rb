# frozen_string_literal: true

module Configuration
  class Jira
    attr_reader :project_key, :issue_type, :fix_version_id, :transition_id

    def initialize(project_key:, issue_type:, fix_version_id:, transition_id: nil)
      @project_key    = project_key
      @issue_type     = issue_type
      @fix_version_id = fix_version_id
      @transition_id  = transition_id
    end
  end
end
