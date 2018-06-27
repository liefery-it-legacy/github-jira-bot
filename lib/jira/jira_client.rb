# frozen_string_literal: true

require "jira-ruby"

module Jira
  module JiraClient
    def self.included(base)
      base.extend ClassMethods
    end

    attr_accessor :username, :site
    attr_writer   :password

    module ClassMethods
      def jira_client
        @jira_client ||= JIRA::Client.new(
          username:     ENV.fetch("JIRA_USER"),
          password:     ENV.fetch("JIRA_PASSWORD"),
          site:         ENV.fetch("JIRA_URL"),
          auth_type:    :basic,
          context_path: ""
        )
      end
    end
  end
end
