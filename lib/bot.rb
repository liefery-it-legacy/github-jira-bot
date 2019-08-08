# frozen_string_literal: true

require "jira/comment"
require "jira/issue"
require "github/comment"
require "github/pull_request"
require "github/reaction"
require "parser/github_to_jira/heading"
require "parser/github_to_jira/image"
require "parser/jira_to_github/heading"

class Bot
  def initialize(repo:, magic_qa_keyword:, max_description_chars:, component_map:, bot_github_login:, jira_configuration:)
    @repo                  = repo
    @magic_qa_keyword      = magic_qa_keyword
    @max_description_chars = max_description_chars
    @component_map         = component_map
    @bot_github_login      = bot_github_login
    @jira_configuration    = jira_configuration
  end

  def handle_comment(action:, title:, comment:, pr_number:, author:, comment_id:)
    @action     = action
    @title      = title
    @comment    = comment
    @qa_comment = extract_qa_comment
    @pr_number  = pr_number
    @author     = author
    @comment_id = comment_id
    @component  = @component_map.fetch(@repo, nil)

    handle_comment_created if @qa_comment.present? && !@qa_comment.empty? && @action == "created" && @author != @bot_github_login
  end

  def handle_pull_request(action:, title:, pr_number:)
    @action = action
    @title  = title

    jira_issue_id = extract_issue_id(@title)
    return if jira_issue_id.nil?

    jira_issue = Jira::Issue.find(jira_issue_id)
    return if jira_issue.nil?

    @jira_url         = jira_issue&.attrs&.dig("url")
    @jira_description = jira_issue&.attrs&.dig("fields", "description")
    @jira_title       = jira_issue&.attrs&.dig("fields", "summary")
    @pr_number        = pr_number

    handle_pull_request_opened if @jira_url.present? && @action == "opened"
  end

  def extract_issue_id(title)
    match_data = title.match(pr_name_ticket_id_regex) || title.match(branch_name_ticket_id_regex)
    return unless match_data

    "#{@jira_configuration.project_key}-" + match_data[1].strip
  end

  private

  def handle_comment_created
    issue = find_or_create_issue(extract_issue_id(@title))
    @qa_comment = Parser::GithubToJira::Image.new.call(@qa_comment)
    @qa_comment = Parser::GithubToJira::Heading.new.call(@qa_comment)
    Jira::Comment.create(issue.key, @qa_comment)
    Github::Reaction.create(@repo, @comment_id, "+1")
  end

  def handle_pull_request_opened
    @jira_description = Parser::JiraToGithub::Heading.new.call(@jira_description)
    Github::Comment.create(@repo, @pr_number, pull_request_comment_content)
    fix_pr_title
  end

  def pull_request_comment_content
    return @jira_url unless @jira_description

    if @max_description_chars
      "#{@jira_description.truncate(@max_description_chars.to_i)}\n\n#{@jira_url}"
    else
      "<details><summary>Ticket description</summary>#{@jira_description}</details>\n\n#{@jira_url}"
    end
  end

  def extract_qa_comment
    qa_comment = @comment[/#{@magic_qa_keyword}(.*\w+.*)/im, 1]
    return unless qa_comment

    "#{@magic_qa_keyword}#{qa_comment}"
  end

  def find_or_create_issue(issue_id)
    (issue_id && Jira::Issue.find(issue_id)) ||
      create_issue_and_update_github_pr_title
  end

  def create_issue_and_update_github_pr_title
    new_issue = create_issue
    update_github_pr_title(new_issue)
    new_issue
  end

  def create_issue
    new_issue = Jira::Issue.create(
      @jira_configuration.project_key,
      @jira_configuration.issue_type,
      @jira_configuration.fix_version_id,
      @component,
      @title
    )
    Jira::Issue.transition(new_issue, @jira_configuration.transition_id) if @jira_configuration.transition_id
    new_issue
  end

  def update_github_pr_title(new_issue)
    prefixed_title = "[#{new_issue.key}] #{@title}"
    Github::PullRequest.update_title(@repo, @pr_number, prefixed_title)
  end

  def fix_pr_title
    id = @title.match(branch_name_ticket_id_regex)
    return unless id

    Github::PullRequest.update_title(@repo, @pr_number, "[#{@jira_configuration.project_key}-#{id[1]}] #{@jira_title}")
  end

  def pr_name_ticket_id_regex
    /\A\[#{@jira_configuration.project_key}-(\d+)\]/i
  end

  def branch_name_ticket_id_regex
    /\A\w+\/#{@jira_configuration.project_key} (\d+)/i
  end
end
