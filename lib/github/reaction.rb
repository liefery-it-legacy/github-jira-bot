# frozen_string_literal: true

require "octokit"

module Github
  class Reaction
    def self.create(repo, comment_id, reaction)
      Octokit.client.create_issue_comment_reaction(
        repo,
        comment_id,
        reaction,
        accept: "application/vnd.github.squirrel-girl-preview+json"
      )
    end
  end
end
