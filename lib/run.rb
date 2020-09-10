# frozen_string_literal: true

$LOAD_PATH.unshift(File.dirname(__FILE__))
require "dotenv"
Dotenv.load

require "json"
require "bot"
require "configuration/jira"

def get_array_from_env(key)
  JSON.parse(ENV[key]) rescue []
end

repo                  = ENV.fetch("REPO", "")
action                = ENV.fetch("ACTION", "")
title                 = ENV.fetch("PR_TITLE", "")
pr_number             = ENV.fetch("PR_NUMBER", "")
pr_labels             = get_array_from_env("PR_LABELS")
author                = ENV.fetch("AUTHOR", "")

comment               = ENV.fetch("COMMENT_BODY", "")
comment_id            = ENV.fetch("COMMENT_ID", "")

magic_qa_keyword      = ENV.fetch("MAGIC_QA_KEYWORD")
max_description_chars = ENV.fetch("MAX_DESCRIPTION_CHARS", nil)
component_map         = JSON.parse(ENV.fetch("COMPONENT_MAP", "{}"))
bot_github_login      = ENV.fetch("GITHUB_USERNAME")
jira_project_key      = ENV.fetch("JIRA_PROJECT_KEY")
jira_issue_type       = ENV.fetch("JIRA_ISSUE_TYPE")
jira_fix_version_id   = ENV.fetch("JIRA_FIX_VERSION_ID")
jira_transition_id    = ENV.fetch("JIRA_NEW_TICKET_TRANSITION_ID")

Octokit.configure do |c|
  c.login    = bot_github_login
  c.password = ENV.fetch("GITHUB_PASSWORD")
end

def pull_request_comment?(action, title, comment, pr_number, author, comment_id)
  pull_request?(action, title, pr_number) && comment.present? && author.present? && comment_id.present?
end

def pull_request?(action, title, pr_number)
  action.present? && title.present? && pr_number.present?
end

jira_configuration = Configuration::Jira.new(
  project_key:    jira_project_key,
  issue_type:     jira_issue_type,
  fix_version_id: jira_fix_version_id,
  transition_id:  jira_transition_id
)
bot = Bot.new(
  repo:                  repo,
  magic_qa_keyword:      magic_qa_keyword,
  max_description_chars: max_description_chars,
  component_map:         component_map,
  bot_github_login:      bot_github_login,
  jira_configuration:    jira_configuration
)

begin
  if pull_request_comment?(action, title, comment, pr_number, author, comment_id)
    bot.handle_comment(action: action, title: title, comment: comment, pr_number: pr_number, author: author, comment_id: comment_id)
  elsif pull_request?(action, title, pr_number)
    bot.handle_pull_request(action: action, title: title, pr_number: pr_number, pr_labels: pr_labels)
  end
rescue JIRA::HTTPError => e
  puts "JIRA responded with #{e.response.code}: #{e.response.body}"
  raise
end
