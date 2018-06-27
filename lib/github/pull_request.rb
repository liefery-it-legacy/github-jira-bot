# frozen_string_literal: true

require "octokit"

module Github
  class PullRequest
    def self.update_title(repo, issue_number, title)
      Octokit.client.update_issue(repo, issue_number, title: title)
    end
  end
end
