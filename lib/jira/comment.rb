# frozen_string_literal: true

require "jira/jira_client"

module Jira
  class Comment
    include Jira::JiraClient

    def self.create(issue_id, comment_body)
      issue = jira_client.Issue.find(issue_id)
      comment = issue.comments.build
      comment.save!(body: comment_body)
    end
  end
end
