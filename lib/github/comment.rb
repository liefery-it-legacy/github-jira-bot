# frozen_string_literal: true

require "octokit"

module Github
  class Comment
    def self.create(repo, issue_number, content)
      Octokit.client.add_comment(repo, issue_number, content)
    end
  end
end
