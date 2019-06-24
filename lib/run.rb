# frozen_string_literal: true

$LOAD_PATH.unshift(File.dirname(__FILE__))
require "dotenv"
Dotenv.load

require "json"
require "bot"
require "configuration/jira"

repo                  = ENV.fetch("REPO", "")
action                = ENV.fetch("ACTION", "")
title                 = ENV.fetch("PR_TITLE", "")
pr_number             = ENV.fetch("PR_NUMBER", "")
author                = ENV.fetch("AUTHOR", "")

comment               = ENV.fetch("COMMENT_BODY", "")
comment_id            = ENV.fetch("COMMENT_ID", "")

magic_qa_keyword      = ENV.fetch("MAGIC_QA_KEYWORD")
max_description_chars = ENV.fetch("MAX_DESCRIPTION_CHARS", nil)
component_map         = JSON.parse(ENV.fetch("COMPONENT_MAP", "{}"))
bot_github_login      = ENV.fetch("GITHUB_USERNAME")
jira_project_key      = ENV.fetch("JIRA_PROJECT_KEY")
jira_issue_type       = ENV.fetch("JIRA_ISSUE_TYPE")
jira_transition_id    = ENV.fetch("JIRA_NEW_TICKET_TRANSITION_ID")

Octokit.configure do |c|
  c.login    = bot_github_login
  c.password = ENV.fetch("GITHUB_PASSWORD")
end

def issue_comment?(action, title, comment, pr_number, author, comment_id)
  !action.empty? && !title.empty? && !comment.empty? && !pr_number.empty? && !author.empty? && !comment_id.empty?
end

def pull_request?(action, title, pr_number)
  !action.empty? && !title.empty? && !pr_number.empty?
end

jira_configuration = Configuration::Jira.new(
  project_key:   jira_project_key,
  issue_type:    jira_issue_type,
  transition_id: jira_transition_id
)
bot = Bot.new(
  repo:                  repo,
  magic_qa_keyword:      magic_qa_keyword,
  max_description_chars: max_description_chars,
  component_map:         component_map,
  bot_github_login:      bot_github_login,
  jira_configuration:    jira_configuration
)

if issue_comment?(action, title, comment, pr_number, author, comment_id)
  bot.handle_comment(action: action, title: title, comment: comment, pr_number: pr_number, author: author, comment_id: comment_id)
elsif pull_request?(action, title, pr_number)
  bot.handle_pull_request(action: action, title: title, pr_number: pr_number)
end
