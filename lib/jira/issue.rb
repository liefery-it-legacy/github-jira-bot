# frozen_string_literal: true

require "jira/jira_client"
require "json"

module Jira
  class Issue
    include Jira::JiraClient

    def self.find(issue_id)
      issue = jira_client.Issue.find(issue_id)
      issue.attrs["url"] = "#{@jira_client.options[:site]}/browse/#{issue_id}"
      issue
    end

    def self.create(project, issue_type, component, title)
      issue = jira_client.Issue.build
      issue.save!(
        "fields" => field_attributes(project, issue_type, component, title)
      )
      issue
    end

    def self.transition(issue, transition_id)
      transition = issue.transitions.build
      transition.save!("transition" => { "id" => transition_id })
    end

    class << self
      private

      def field_attributes(project, issue_type, component, title)
        {
          "summary" => title,
          "project" => { "key" => project },
          "issuetype" => { "name" => issue_type },
          "components" => [{ "name" => component }],
          ENV.fetch("JIRA_SPRINT_FIELD") => current_sprint_id
        }
      end

      def current_sprint_id
        http_response = jira_client.get("/rest/greenhopper/1.0/sprint/picker")
        json_response = JSON.parse(http_response.body)
        active_sprints = json_response.fetch("allMatches").select { |sprint| sprint.fetch("stateKey") == "ACTIVE" }
        active_sprints.first.fetch("id")
      end
    end
  end
end
