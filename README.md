# Github-Jira-Bot

A bot for a tighter integration of GitHub into Jira and vice versa.

## What the bot does

To link Jira tickets and GitHub pull request with one another, we at Liefery use [Tickety-Tick](https://github.com/bitcrowd/tickety-tick). The format for title on GitHub is: `[#${Jira ticket key}] ${Jira ticket summary}`

The bot does two things:

1. When a new Pull Request has been opened it adds a comment including description and URL of the associated Jira ticket if such a ticket exists.
2. When a comment has been added to an issue or pull request which includes `QA:` the bot does the following:
  * When no associated ticket exists, the bot creates a ticket in the current sprint and updates the title of the Pull Request accordingly.
  * The bot adds everything in the GitHub comment that comes after `QA:` as a comment to the associated Jira ticket. The magic QA keyword can be [configured](#configuration).

## Usage

### Setup the bot on Jenkins

First you need a place for the bot to run. We at Liefery use Jenkins for this.

* Create a new item from the Jenkins dashboard and choose `Freestyle project`
* Select `Git` for Source Code Management:
  * Repository URL: `git@github.com:liefery/github-jira-bot.git`
  * Branch Specifier: `*/production`
* Select `Trigger builds remotely (e.g., from scripts)` for Build Triggers:
  * Choose an `Authentication Token`. Remember your token as you'll need it later when [adding the bot to repositories on GitHub](#adding-the-bot-to-new-repos).
* Select `Generic Webhook Trigger` for Build Triggers:
  * Add some `Post content parameters` (choose `JSONPath` as expression format):
    * `ACTION` => `$.action`
    * `REPO` => `$.repository.full_name`
    * `PR_TITLE` => `$.pull_request.title`
    * `PR_TITLE` => `$.issue.title`
    * `AUTHOR` => `$.comment.user.login`
    * `PR_NUMBER` => `$.pull_request.number`
    * `PR_NUMBER` => `$.issue.number`
    * `COMMENT_ID` => `$.comment.id`
    * `COMMENT_BODY` => `$.comment.body`
  * Select `Print post content`
  * Select `Print contributed variables`
* Add an `Execute shell` build step for Build:
  * Command: Paste your [configuration](#configuration).

**Note:** `PR_TITLE` and `PR_NUMBER` will be set in different ways depending on the webhook used to trigger the bot.

### Adding the bot to a repository

In a GitHub repository navigate to settings/webhooks and add a new webhook using the following configuration:

* Payload URL: `https:/JENKINS_URL/generic-webhook-trigger/invoke?token=XXX` (use the same token as configured in the `Authentication Token` setting of the Jenkins build it should trigger)
* Content type: `application/json`
* Choose `Let me select individual events`
  * select `Issue comments`
  * select `Pull requests`
  * **deselect** `Pushes`

### Setup bot accounts

It is recommended that you don't use your personal GitHub and Jira accounts for the bot.

You can create accounts on GitHub and Jira with permission to access and contribute to the repositories/projects you want to use the bot for.

## Configuration

Here is a sample configuration:

```
export JIRA_USER = "XXX"
export JIRA_PASSWORD = "XXX"
export JIRA_URL = "https://XXX.atlassian.net"
export JIRA_PROJECT_KEY = "XXX"
export JIRA_ISSUE_TYPE = "Story"
export JIRA_FIX_VERSION_ID = "XXX"
export JIRA_SPRINT_FIELD = "customfield_XXX"
export JIRA_NEW_TICKET_TRANSITION_ID = 711
export GITHUB_USERNAME = "XXX"
export GITHUB_PASSWORD = "XXX"

export MAGIC_QA_KEYWORD = "QA:"
export MAX_DESCRIPTION_CHARS = 600
export COMPONENT_MAP = '{"repo":"component"}'


bundle install
ruby lib/run.rb
```

* `JIRA_USER`: Username of a Jira account.
* `JIRA_PASSWORD`: Password of a Jira account.
* `JIRA_URL`: URL of your Jira workspace.
* `JIRA_PROJECT_KEY`: Key of a Jira project. This looks something like `BT`.
* `JIRA_ISSUE_TYPE`: Can be `Story`, `Improvement` or similar. This will only be used if the bot creates tickets.
* `JIRA_FIX_VERSION_ID`: Id of the fix version to use when creating a Jira ticket.
* `JIRA_SPRINT_FIELD`: The identifier of the sprint new tickets should be attached to upon creation.
* `JIRA_NEW_TICKET_TRANSITION_ID`: The database ID of the transition to the state where newly created Jira tickets should be moved to.
* `GITHUB_USERNAME`: Username of a GitHub account.
* `GITHUB_PASSWORD`: Password of a GitHub account.
* `MAGIC_QA_KEYWORD`: Keyword that detects information in comments that should be added to the Jira ticket.
* `MAX_DESCRIPTION_CHARS`: The maximum number of chars of a Jira ticket description that will be added to a pull request on GitHub. Omit this environment variable if you want to add the entire ticket description to GitHub.
* `COMPONENT_MAP`: Mapping of GitHub repositories and Jira components used when creating new Jira tickets. Has to be a stringified JSON hash.

## Local Development

1. Copy `.env.sample` to `.env`.
2. Edit `.env` and setup accounts.
3. Run `ruby lib/run.rb`.

## Running tests

This project uses RSpec for testing:

1. `bundle install`
2. `bundle exec rspec`

## Deployment

```
git fetch
git checkout production
git pull
git reset --hard origin/master
git push
```
