package bot

import (
	"context"
	localGitHub "github.com/Benbentwo/github-jira-bot/pkg/github"
	"github.com/Benbentwo/github-jira-bot/pkg/helpers"
	localJira "github.com/Benbentwo/github-jira-bot/pkg/jira"
	"github.com/Benbentwo/utils/util"
	"github.com/go-errors/errors"
	"github.com/google/go-github/github"
	"strings"

	// "gopkg.in/andygrunwald/go-jira.v1"
	"github.com/andygrunwald/go-jira"

	"log"
	"os"
	"regexp"
	"strconv"
)

// # frozen_string_literal: true
//
// require "jira/comment"
// require "jira/issue"
// require "github/comment"
// require "github/pull_request"
// require "github/reaction"
// require "parser/github_to_jira/heading"
// require "parser/github_to_jira/image"
// require "parser/jira_to_github/heading"
//
// # rubocop:disable Metrics/ClassLength

const (
	ActionCreatedStatus               = "created"
	ActionOpenedStatus                = "opened"
	GoLangCaseInsensitiveRegexAndWord = "(?mi)"
	GoLangCaseInsensitiveRegex        = "(?i)"
)

var DefaultComponentMap = map[string]string{
	"repo": "component",
}

func NewJiraConfigFromJira(project jira.Project, issueType jira.IssueType, version jira.FixVersion, transition jira.Transition) (*localJira.JiraConfig, error) {
	return &localJira.JiraConfig{
		ProjectKey: project.Key,
		IssueType:  issueType,
		FixVersion: version,
		Transition: transition,
	}, nil
}
func NewJiraConfigFromEnv() *localJira.JiraConfig {
	var (
		jiraProjectKey   = os.Getenv("JIRA_PROJECT_KEY")
		jiraIssueType    = os.Getenv("JIRA_ISSUE_TYPE")
		jiraFixVersionId = os.Getenv("JIRA_FIX_VERSION_ID")
		jiraTransitionId = os.Getenv("JIRA_NEW_TICKET_TRANSITION_ID")
	)

	return NewJiraConfigFromStrings(jiraProjectKey, jiraIssueType, jiraFixVersionId, jiraTransitionId)

}
func NewJiraConfigFromStrings(jiraProjectKey string, jiraIssueType string, jiraFixVersionId string, jiraTransitionId string) *localJira.JiraConfig {
	issueType := jira.IssueType{
		Name: jiraIssueType,
	}
	jiraFixVersion := jira.FixVersion{
		Name: jiraFixVersionId,
	}
	jiraTransition := jira.Transition{
		ID: jiraTransitionId,
	}

	jiraConfig, err := NewJiraConfigFromJira(jira.Project{Key: jiraProjectKey}, issueType, jiraFixVersion, jiraTransition)
	if err != nil {
		log.Fatalf("Error on New Jira Config from Jira %s", err)
	}
	return jiraConfig
}

// def initialize(repo:, magic_qa_keyword:, max_description_chars:, component_map:, bot_github_login:, jira_configuration:)
// @repo                  = repo
// @magic_qa_keyword      = magic_qa_keyword
// @max_description_chars = max_description_chars
// @component_map         = component_map
// @bot_github_login      = bot_github_login
// @jira_configuration    = jira_configuration
// end
//

type Bot struct {
	Repo   string //NOTE this represents <Org>/<Repo>
	Action string

	PrTitle       string
	PrNumber      int
	PrAuthor      string
	PrDescription string

	IssueNum int

	Comment       string
	CommentAuthor string
	CommentId     int
	IssueTitle    string
	IssueNumber   int

	MagicQAWord  string
	QAComment    string
	MaxDesc      int
	ComponentMap map[string]string
	Component    string
	GithubLogin  string

	//JiraUrl			string
	JiraDesc  string
	JiraTitle string

	JiraConfig   localJira.JiraConfig
	GithubConfig localGitHub.GHConfig

	JiraClient   *jira.Client
	GithubClient *github.Client
}

//
//// TODO delete this
//func NewBot(repo string, magic string, max int, componentMap map[string]string, githubLogin string, config localJira.JiraConfig,
//	action string, title string, prNumber int, author string, comment string, commentAuthor string, commentId int) (*Bot, error) {
//	return &Bot{
//		Repo:     repo,
//		Action:   action,
//
//		PrTitle:  title,
//		PrNumber: prNumber,
//		PrAuthor: author,
//
//		Comment:   comment,
//		CommentAuthor:	commentAuthor,
//		CommentId: commentId,
//
//		MagicQAWord:  magic,
//		MaxDesc:      max,
//		ComponentMap: componentMap,
//		GithubLogin:  githubLogin,
//		JiraConfig:   config,
//	}, nil
//}
//
//// TODO delete this
//func NewBotFromEnv() (*Bot, error) {
//	var (
//		repo     = os.Getenv("REPO")
//		action   = os.Getenv("ACTION")
//
//		title    = os.Getenv("PR_TITLE")
//		prNumber = os.Getenv("PR_NUMBER")
//		author   = os.Getenv("AUTHOR")
//
//		comment      		= os.Getenv("COMMENT_BODY")
//		commentAuthor 		= os.Getenv("COMMENT_AUTHOR")
//		commentIdStr 		= os.Getenv("COMMENT_ID")
//
//		//magicQaKeyword      = os.Getenv("MAGIC_QA_KEYWORD")
//		//maxDescriptionChars = os.Getenv("MAX_DESCRIPTION_CHARS")
//		//componentMap        = os.Getenv("COMPONENT_MAP")
//		//botGithubLogin      = os.Getenv("GITHUB_USERNAME")
//	)
//
//	//var data map[string]string
//	//err := json.Unmarshal([]byte(componentMap), data)
//	//if err != nil {
//	//	data = DefaultComponentMap
//	//}
//
//	//max := convertStrToInt(maxDescriptionChars)
//	prNum := convertStrToInt(prNumber)
//	commentId := convertStrToInt(commentIdStr)
//
//	return NewBot(repo, magicQaKeyword, max, data, botGithubLogin, *NewJiraConfigFromEnv(),
//		action, title, prNum, author, comment, commentAuthor, commentId)
//}

func (bot *Bot) SetKeyValuePairsFromEnv() {
	var (
		repo   = os.Getenv("REPO")   // $.repository.full_name
		action = os.Getenv("ACTION") // $.action

		prNumber = os.Getenv("PR_NUMBER") // $.pull_request.number
		prTitle  = os.Getenv("PR_TITLE")  // $.pull_request.title
		prAuthor = os.Getenv("PR_AUTHOR") // $.pull_request.user.login
		prDesc   = os.Getenv("PR_DESC")   // $.pull_request.body

		commentAuthor = os.Getenv("COMMENT_AUTHOR") // $.comment.user.login
		comment       = os.Getenv("COMMENT_BODY")   // $.comment.body
		commentIdStr  = os.Getenv("COMMENT_ID")     // $.comment.id
		issueTitle    = os.Getenv("ISSUE_TITLE")    // $.issue.title
		issueNumStr   = os.Getenv("ISSUE_NUMBER")   // $.issue.number
	)

	prNum := convertStrToInt(prNumber)
	issueNumber := convertStrToInt(issueNumStr)
	commentId := convertStrToInt(commentIdStr)

	bot.Repo = repo
	bot.Action = action

	bot.PrTitle = prTitle
	bot.PrNumber = prNum
	bot.PrAuthor = prAuthor
	bot.PrDescription = prDesc

	bot.Comment = comment
	bot.CommentAuthor = commentAuthor
	bot.CommentId = commentId
	bot.IssueTitle = issueTitle
	bot.IssueNumber = issueNumber

	// bot.ComponentMap = componentMap
	//bot.GithubLogin = botGithubLogin
}

func convertStrToInt(str string) int {
	if str == "" {
		return 0
	}
	variable, err := strconv.Atoi(str)
	if err != nil {
		util.Logger().Errorf("Error converting max to int: %s", err)
	}
	return variable
}

func (bot *Bot) RunComment() {
	bot.QAComment = bot.ExtractQaComment() //Comment = <Words>, QAComment = <magicWord> + <Words>
	if bot.QAComment == "" {
		return
	}
	bot.Component = bot.ComponentMap[bot.Repo]

	if bot.QAComment != "" && bot.Action == ActionCreatedStatus && bot.PrAuthor != bot.GithubLogin {
		bot.HandleCommentCreated()
	}
}

//   def handle_comment_created
//    issue = find_or_create_issue(extract_issue_id(@title))
//    @qa_comment = Parser::GithubToJira::Image.new.call(@qa_comment)
//    @qa_comment = Parser::GithubToJira::Heading.new.call(@qa_comment)
//    Jira::Comment.create(issue.key, @qa_comment)
//    Github::Reaction.create(@repo, @comment_id, "+1")
//  end
//
func (bot *Bot) HandleCommentCreated() {
	issueId := bot.ExtractIssueId("")
	if issueId == "" {
		util.Warn("No Issue Found")
		// The below code will create a new issue if the title or branch doesn't have a reference, omitting for now
		// As I think most people would not want jira issues made for each pull request.

		//bot.PrNumber
		_, _, err := bot.CreateIssueAndUpdateGithubPrTitle()
		if err != nil {
		}
		return
	}

	issue := bot.FindOrCreateIssue(strconv.Itoa(bot.IssueNum))
	//bot.QAComment = helpers.ConvertMarkdownToHtml(bot.QAComment)
	util.Logger().Debugf("Comment: %s", bot.Comment)
	util.Logger().Debugf("QAComment: %s", bot.QAComment)
	comment := &jira.Comment{
		Body: bot.Comment,
	}
	comment, _, err := bot.JiraClient.Issue.AddComment(issue.ID, comment)
	if err != nil {
		log.Fatalf("HandleCommentCreated errored with: %s", err)
	}

}

// def handle_pull_request(action:, title:, pr_number:)
// @action    = action
// @title     = title
// jira_issue = find_issue_by_title(@title)
// return if jira_issue.nil?
//
// @jira_url         = jira_issue&.attrs&.dig("url")
// @jira_description = jira_issue&.attrs&.dig("fields", "description")
// @jira_title       = jira_issue&.attrs&.dig("fields", "summary")
// @pr_number        = pr_number
//
// handle_pull_request_opened if @jira_url.present? && @action == "opened"
// end
func (bot *Bot) HandlePullRequest(action string, title string, prNumber int) bool {
	bot.Action = action
	bot.PrTitle = title
	bot.PrNumber = prNumber

	issue, err := bot.FindIssueByTitle()
	if err != nil {
		log.Fatalf("Handle Pull Request failed: %s", err)
	}

	bot.JiraConfig.JiraUrl = issue.Self
	bot.JiraDesc = issue.Fields.Description
	bot.JiraTitle = issue.Fields.Summary

	if bot.JiraConfig.JiraUrl != "" && bot.Action == ActionOpenedStatus {
		bot.HandlePullRequestOpened()
	}
	return true
}
func (bot *Bot) RunPullRequest() bool {

	issue, err := bot.FindIssueByTitle()
	if err != nil {
		log.Fatalf("Handle Pull Request failed: %s", err)
	}

	bot.JiraConfig.JiraUrl = "https://" + strings.TrimPrefix(
		strings.TrimSuffix(bot.JiraConfig.JiraUrl, "/"), "https://") + "/browse/" +
		bot.JiraConfig.ProjectKey + "-" + strconv.Itoa(bot.IssueNum)

	bot.JiraDesc = issue.Fields.Description
	bot.JiraTitle = issue.Fields.Summary

	if bot.JiraConfig.JiraUrl != "" && bot.Action == ActionOpenedStatus {
		_ = bot.HandlePullRequestOpened()
	}
	return true
}

//
// def extract_issue_id(title)
// match_data = title.match(pr_name_ticket_id_regex) || title.match(branch_name_ticket_id_regex)
// return unless match_data
//
// "#{@jira_configuration.project_key}-" + match_data[1].strip
// end
//
func (bot *Bot) ExtractIssueId(title string) string {
	if title == "" {
		if bot.PrTitle != "" {
			title = bot.PrTitle
		} else if bot.IssueTitle != "" {
			title = bot.IssueTitle
		}
	}
	if title == "" {
		log.Fatal("Title is empty")
	}
	util.Logger().Debugf("Extracting Issue ID from %s", title)

	re := regexp.MustCompile(bot.PrNameTicketIdRegex())
	arr := re.FindStringSubmatch(title)

	util.Logger().Debugf("Match Length: %d\n", len(arr))
	if len(arr) > 0 {
		util.Logger().Debugf("Match 0: %s\n", arr[0])
		util.Logger().Debugf("Match 1: %s\n", arr[1])
		issueNum, err := strconv.Atoi(strings.TrimSpace(arr[1]))
		if err != nil {
			util.Logger().Errorf("Error converting issueNum to an int: %s", strings.TrimSpace(arr[1]))
		}
		bot.IssueNum = issueNum
		util.Logger().Debugf("IssueNum: %d\n", bot.IssueNum)
		return "#" + bot.JiraConfig.ProjectKey + "-" + strings.TrimSpace(arr[1])
	}

	util.Logger().Debug("Checking for Branch Name match")

	re = regexp.MustCompile(bot.BranchNameTicketIdRegex())
	arr = re.FindStringSubmatch(title)

	util.Logger().Debugf("Match Length: %d\n", len(arr))
	if len(arr) > 1 {
		util.Logger().Debugf("Branch Match: %s\n", arr[1])
		issueNum, err := strconv.Atoi(strings.TrimSpace(arr[1]))
		if err != nil {
			util.Logger().Errorf("Error converting issueNum to an int: %s", strings.TrimSpace(arr[1]))
		}
		bot.IssueNum = issueNum
		util.Logger().Debugf("IssueNum: %d\n", bot.IssueNum)
		return "#" + bot.JiraConfig.ProjectKey + "-" + strings.TrimSpace(arr[1])
	}

	util.Logger().Errorf("Error finding title or branch match: %q", arr)
	// TODO fix this
	//os.Exit(1)
	return ""
}

// private
//
// def find_issue_by_title(title)
// jira_issue_id = extract_issue_id(title)
// return if jira_issue_id.nil?
//
// Jira::Issue.find(jira_issue_id)
// end
func (bot *Bot) FindIssueByTitle() (*jira.Issue, error) {
	util.Logger().Debug("Function Call FindIssueByTitle")
	jiraIssueId := bot.ExtractIssueId("")
	if bot.JiraClient == nil {
		return nil, errors.Errorf("FindIssueByTitle error: bot's JiraClient is nil")
	}
	util.Logger().Debugf("Fetching issue %s", jiraIssueId)
	return GetJiraIssue(bot.JiraClient, strings.TrimPrefix(jiraIssueId, "#"), &jira.GetQueryOptions{})
}
func (bot *Bot) FindIssueById(id int) (*jira.Issue, error) {
	util.Logger().Debug("Function Call FindIssueById")
	if bot.JiraClient == nil {
		return nil, errors.Errorf("FindIssueByTitle error: bot's JiraClient is nil")
	}
	return GetJiraIssue(bot.JiraClient, bot.JiraConfig.ProjectKey+"-"+strconv.Itoa(id), &jira.GetQueryOptions{})
}

func GetJiraIssue(jiraClient *jira.Client, issueId string, queryOptions *jira.GetQueryOptions) (*jira.Issue, error) {
	issue, response, err := jiraClient.Issue.Get(issueId, queryOptions)
	defer response.Body.Close()
	if err != nil {
		// This might be able to be safer, check for something besides 404.
		// 	Ideally we're not comparing error message though, because if they change the api message then were horked
		if response.StatusCode == 404 {
			return nil, nil // no issue found but also no error
		}
		return nil, errors.Errorf("Fetching Jira Issue Failed: %s", err)
	}

	return issue, nil
}

func parseJiraHeader(desc string) string {
	return helpers.ConvertHtmlToMarkdown(desc)
}
func (bot *Bot) AddCommentToIssue(issueNumber int, content string) error {
	org, repo, err := localGitHub.SeparateOrgAndRepo(bot.Repo)
	if err != nil {
		return err
	}
	comment := &github.IssueComment{
		Body: &content,
	}
	_, _, err = bot.GithubClient.Issues.CreateComment(context.Background(), org, repo, issueNumber, comment)
	if err != nil {
		return err
	}
	return nil
}

func (bot *Bot) AddCommentToPR(prNumber int, content string) error {
	org, repo, err := localGitHub.SeparateOrgAndRepo(bot.Repo)
	if err != nil {
		return err
	}
	comment := &github.PullRequestComment{
		Body: &content,
	}
	_, _, err = bot.GithubClient.PullRequests.CreateComment(context.Background(), org, repo, prNumber, comment)
	if err != nil {
		return err
	}
	return nil
}

// def handle_pull_request_opened
// @jira_description = Parser::JiraToGithub::Heading.new.call(@jira_description)
// Github::Comment.create(@repo, @pr_number, pull_request_comment_content)
// fix_pr_title
// end
//
func (bot *Bot) HandlePullRequestOpened() error {
	//bot.JiraDesc = parseJiraHeader(bot.JiraDesc)
	err := bot.AddCommentToIssue(bot.PrNumber, bot.PullRequestCommentContent())
	if err != nil {
		return err
	}

	return bot.FixPrTitle()
}

// def pull_request_comment_content
// return @jira_url unless @jira_description
//
// if @max_description_chars
// "#{@jira_description.truncate(@max_description_chars.to_i)}\n\n#{@jira_url}"
// else
// "<details><summary>Ticket description</summary>#{@jira_description}</details>\n\n#{@jira_url}"
// end
// end
func (bot *Bot) PullRequestCommentContent() string {
	if bot.JiraDesc == "" {
		return bot.JiraConfig.JiraUrl
	}

	if bot.MaxDesc > 0 {
		return truncateString(bot.JiraDesc, bot.MaxDesc) + "\n\n[Jira Ticket](" + bot.JiraConfig.JiraUrl + ")"
	} else {
		return bot.JiraDesc + "\n\n[Jira Ticket](" + bot.JiraConfig.JiraUrl + ")"
	}
}

//
// def extract_qa_comment
// qa_comment = @comment[/#{@magic_qa_keyword}(.*\w+.*)/im, 1]
// return unless qa_comment
//
// "#{@magic_qa_keyword}#{qa_comment}"
// end
func (bot *Bot) ExtractQaComment() string {
	re, err := regexp.Compile(bot.MagicQAWord + `(.*\w+.*)`)
	if err != nil {
		log.Fatal("Error compiling magic word as regex.")
	}
	util.Logger().Debugf("BotComment %s", bot.Comment)
	if strings.Contains(bot.Comment, bot.MagicQAWord) {
		arr := re.FindAllStringSubmatch(bot.Comment, -1)
		bot.Comment = arr[0][1]
		util.Logger().Debugf("QA COMMENT: %s", bot.Comment)

		return bot.MagicQAWord + " " + bot.Comment
	}
	return ""

}

//
// def find_or_create_issue(issue_id)
// (issue_id && Jira::Issue.find(issue_id)) ||
// create_issue_and_update_github_pr_title
// end
func (bot *Bot) FindOrCreateIssue(issueId string) *jira.Issue {
	// client, _ := jira.NewClient()
	var id = -1
	var err error
	util.Logger().Debugf("Issue Id: %s", issueId)
	if issueId == "" {
		util.Logger().Error("No Issue passed to Find Issue Function, using bot default")
		id = bot.IssueNum
	} else {

		id, err = strconv.Atoi(issueId)
		if err != nil {
			log.Fatalf("error converting string to int %s", err)
		}

	}
	issue, err := bot.FindIssueById(id)
	if err != nil {
		// couldn't get the issue and there was an error
		util.Logger().Errorf("Could not get issue, %d: %s", id, err)
	}
	if issue != nil {
		// issue found.
		util.Logger().Debugf("Issue Found: %s", issue.Self)
		return issue
	} else {
		util.Logger().Debug("Issue Not Found, Creating")
		// there wasn't an error (404 okay) but no issue was found. so lets create one
		jiraIssue, _, err := bot.CreateIssueAndUpdateGithubPrTitle()
		if err != nil {
			util.Logger().Errorf("Issue on creating the issue and updating github Pr PrTitle")
		}
		return jiraIssue
	}

}

//
// def create_issue_and_update_github_pr_title
// new_issue = create_issue
// update_github_pr_title(new_issue)
// new_issue
// end
//
func (bot *Bot) CreateIssueAndUpdateGithubPrTitle() (*jira.Issue, *github.Issue, error) {
	jiraIssue := bot.CreateJiraIssue()
	if jiraIssue == nil {
		return nil, nil, errors.Errorf("Error Creating Jira Issue: %s")
	}
	u, err := bot.GetJiraTicketUrl(jiraIssue)
	util.Logger().Debugln("Created Jira Issue", u)
	ghIssue, err := bot.UpdateGithubPrTitle(jiraIssue)
	if err != nil {
		return nil, nil, errors.Errorf("Create Issue and Update Pr failed with error: %s", err)
	}
	return jiraIssue, ghIssue, nil
}

func (bot *Bot) GetJiraTicketUrl(issue *jira.Issue) (string, error) {
	if issue == nil {
		return "", nil
	}
	return bot.JiraConfig.JiraUrl + "browse/" + issue.Key, nil
}

func (bot *Bot) GetGithubIssueUrl(issue *github.Issue) (string, error) {
	if issue == nil {
		return "", errors.Errorf("Cannot get URL of empty github issue!")
	}
	return "https://" + strings.TrimSuffix(bot.GithubConfig.Url, "/") + "/" + bot.Repo + "/issues/" + strconv.Itoa(*issue.Number), nil
}

func (bot *Bot) MustGetGithubIssueUrl(issue *github.Issue) string {
	str, err := bot.GetGithubIssueUrl(issue)
	if err != nil {
		util.Logger().Fatalf("Must Get Github Issue URL failed: %s", err)
	}
	return str
}

// def create_issue
// new_issue = Jira::Issue.create(
// @jira_configuration.project_key,
// @jira_configuration.issue_type,
// @jira_configuration.fix_version_id,
// @component,
// @title
// )
// Jira::Issue.transition(new_issue, @jira_configuration.transition_id) if @jira_configuration.transition_id
// new_issue
// end
//
func (bot *Bot) CreateJiraIssue() *jira.Issue {

	fixVersions := make([]*jira.FixVersion, 0)
	fixVersions = append(fixVersions, &bot.JiraConfig.FixVersion)

	//components := make([]*jira.Component, 0)
	//components = append(components, &jira.Component{Name: bot.Repo})

	newIssue := jira.Issue{
		Fields: &jira.IssueFields{
			Summary:     bot.Comment,
			Project:     jira.Project{Key: bot.JiraConfig.ProjectKey},
			Type:        bot.JiraConfig.IssueType,
			FixVersions: fixVersions,
			//Components:		components,
		},
	}

	if bot.JiraConfig.Transition.ID != "" {
		newIssue.Transitions = append(newIssue.Transitions, jira.Transition{
			ID: bot.JiraConfig.Transition.ID,
		})
	}
	issue, resp, err := bot.JiraClient.Issue.Create(&newIssue)
	if err != nil {
		util.Logger().Debugf("Jira New Issue Response: %d", resp.StatusCode)
		util.Logger().Errorf("Bot Client couldn't create Jira Issue: %s", err)
		helpers.PrintBody(resp.Body, err)
		return nil
	}
	return issue
}

// def update_github_pr_title(new_issue)
// prefixed_title = "[#{new_issue.key}] #{@title}"
// Github::PullRequest.update_title(@repo, @pr_number, prefixed_title)
// end

// TODO support custom formatting of title here
func (bot *Bot) UpdateGithubPrTitle(issue *jira.Issue) (*github.Issue, error) {
	var title string
	if bot.PrTitle != "" {
		title = bot.PrTitle
	} else if bot.IssueTitle != "" {
		title = bot.IssueTitle
	} else {
		return nil, errors.Errorf("Error determining which title to use")
	}
	prefixedTitle := "[#" + issue.Key + "] " + title

	org, repo, err := localGitHub.SeparateOrgAndRepo(bot.Repo)

	if err != nil {
		return nil, err
	}
	//id, err := strconv.Atoi(issue.ID) // Jira Issue Number
	//if err != nil { return nil, err}

	util.Logger().Debugf("PR: %d, Issue %d, IssueNumber: %d", bot.PrNumber, bot.IssueNum, bot.IssueNumber)
	var ghIssue *github.Issue

	if bot.PrNumber != 0 {
		ghIssue, err = localGitHub.UpdateTitle(bot.GithubClient, org, repo, bot.PrNumber, prefixedTitle)
		if err != nil {
			util.Logger().Fatalf("Error updating github title on Pr Number %d", bot.PrNumber)
		}
	} else if bot.IssueNumber != 0 {
		ghIssue, err = localGitHub.UpdateTitle(bot.GithubClient, org, repo, bot.IssueNumber, prefixedTitle)
		if err != nil {
			util.Logger().Fatalf("Error updating github title on Issue Number %d", bot.IssueNumber)
		}
	} else if bot.IssueNum != 0 {
		ghIssue, err = localGitHub.UpdateTitle(bot.GithubClient, org, repo, bot.IssueNum, prefixedTitle)
		if err != nil {
			util.Logger().Fatalf("Error updating github title on Issue Num %d", bot.IssueNum)
		}
	} else {
		util.Logger().Fatal("No number found to update a title upon")
	}

	if err != nil {
		return nil, errors.Errorf("Error updating github PR title with issue id: %s", issue.ID)
	}
	util.Logger().Debugf("Updated Github Issue %s", bot.MustGetGithubIssueUrl(ghIssue))
	return ghIssue, nil
}

//
// def fix_pr_title
// id = @title.match(branch_name_ticket_id_regex)
// return unless id
//
// Github::PullRequest.update_title(@repo, @pr_number, "[#{@jira_configuration.project_key}-#{id[1]}] #{@jira_title}")
// end

func (bot *Bot) FixPrTitle() error {
	re, err := regexp.Compile(bot.BranchNameTicketIdRegex())
	if err != nil {
		return err
	}
	match := re.MatchString(bot.PrTitle)
	if !match {
		return errors.Errorf("Error getting title from regex, title: %s, regex `%s`", bot.BranchNameTicketIdRegex(), bot.PrTitle)
	}
	id := re.FindAllString(bot.PrTitle, -1)
	org, repo, err := localGitHub.SeparateOrgAndRepo(bot.Repo)
	// err = bot.CreateClientIfDNE()
	if err != nil {
		return errors.Errorf("Error Creating Client: %s", err)
	}
	_, err = localGitHub.UpdateTitle(bot.GithubClient, org, repo, bot.PrNumber, "[#"+bot.JiraConfig.ProjectKey+"-#"+id[1]+" #"+bot.JiraTitle)
	if err != nil {
		return errors.Errorf("Error updating title: %s", err)
	}
	return nil
}

//
// def pr_name_ticket_id_regex
// /\A\[#{@jira_configuration.project_key}-(\d+)\]/i
// end

func (bot *Bot) PrNameTicketIdRegex() string {
	// If it finds [#ABC-123] referenced at the start of the string, this regex
	//   will match 123 as the first group. which should be the ticket number.
	return GoLangCaseInsensitiveRegexAndWord + `\A\[#` + bot.JiraConfig.ProjectKey + `-(\d+)\]`
}

//

// def branch_name_ticket_id_regex
// /\A\w+\/#{@jira_configuration.project_key} (\d+)/i
// end
// end
// # rubocop:enable Metrics/ClassLength
func (bot *Bot) BranchNameTicketIdRegex() string {
	// If it finds `words #ABC 123` referenced at the start of the string, this regex
	//   will match 123 as the first group. which should be the ticket number.
	// `\A` Start of string
	// `\w+` match words (branch name)
	// `\/?` allow `/` which could be used in branch names (feature/rewrite)
	// `\w+` more words
	// `\s+` allow whitespace
	// `\#` KEYWORD `#` Pound Symbol for matching ProjectKey
	// `\s+` allow whitespace
	// `(\d+) Group match digits - meaning this is returned as the group - so what were trying to fetch. should be ticket id.
	//(?i)\A[a-zA-Z0-9_\/-]+\s+
	REGEX := GoLangCaseInsensitiveRegexAndWord + GoLangCaseInsensitiveRegex + `\A[a-zA-Z0-9_\/-]+\s+\#` + bot.JiraConfig.ProjectKey + `\s+(\d+)`
	//util.Logger().Debugf("Regex: '%s'", REGEX)
	return REGEX
}

func (bot *Bot) CreateClientIfDNE(e bool, eUrl string) error {
	if bot.GithubClient != nil {
		log.Printf("GithubClient is not null, moving on")
		return nil
	}
	return bot.CreateGithubClient("", e, eUrl)
}
func (bot *Bot) CreateGithubClient(tokenEnv string, enterprise bool, enterpriseUrl string) error {
	//token, err := localGitHub.GetToken(tokenEnv)
	//if err != nil {
	//	return errors.Errorf("error getting token: %s", err)
	//}
	client, err := localGitHub.CreateClient(
		tokenEnv,
		enterprise,
		enterpriseUrl,
	)
	if err != nil {
		return errors.Errorf("Error creating client %s", err)
	}

	bot.GithubClient = client
	return nil
}

func (bot *Bot) ValidateGithubClient() error {
	util.Logger().Debugf("bot GH Validator:\n\t%s:\t %t\n\t%s:\t %s\n\t%s:\t %s", "Enterprise", bot.GithubConfig.Enterprise, "Url", bot.GithubConfig.Url, "Username", bot.GithubConfig.Username)
	if bot.GithubClient == nil {
		return errors.Errorf("Github Client is nil, try running bot.CreateGithubClient(token, false, `github.com`) first")
	}
	// github.com
	if !bot.GithubConfig.Enterprise {
		str, _, err := bot.GithubClient.Zen(context.Background())
		if err != nil {
			return errors.Errorf("Error getting github zen, github client might not be properly configured: %s", err)
		}
		util.Logger().Debugf("Zen String: `%s`", str)
		repos, _, err := bot.GithubClient.Repositories.List(context.Background(), "", &github.RepositoryListOptions{})
		if err != nil {
			return errors.Errorf("Error getting github user repositories, github client might not be properly configured: %s", err)
		}
		util.Logger().Debugf("Number of repos owned by user: `%d`", len(repos))
		return nil
	}

	//github.ABC.com
	repos, _, err := bot.GithubClient.Repositories.List(context.Background(), "", &github.RepositoryListOptions{})
	if err != nil {
		return errors.Errorf("Error getting github user repositories, github client might not be properly configured: %s", err)
	}
	util.Logger().Debugf("Number of repos owned by user: `%d`", len(repos))
	util.Logger().Debugln()
	return nil
}

func (bot *Bot) CreateJiraClient(jiraUrl string, username string, bytePassword []byte) error {
	util.Logger().Debugf("Jira config:\n\tUrl:\t%s\n\tUsername:%s", jiraUrl, username)
	client, err := localJira.GetJiraClient(jiraUrl, username, bytePassword)
	if err != nil {
		util.Logger().Errorf("Error creating jira client \n\tUrl:\t%s\n\tUsername:%s\nError: \t%s", jiraUrl, username, err)
	}
	bot.JiraClient = client
	return err

}

func (bot *Bot) ValidateJiraClient() error {
	if bot.JiraClient == nil {
		return errors.Errorf("Jira Client is nil, try running bot.CreateJiraClient(url, username, password) first")
	}

	user, _, err := bot.JiraClient.User.GetSelf()
	//user, _, err := bot.JiraClient.User.Get(bot.JiraConfig.JiraUser)
	if err != nil {
		return errors.Errorf("Error getting user from Jira, jira client might not be properly configured: %s", err)
	}
	util.Logger().Debugf("User Info: \n\tDisplay Name:\t%s\n\tAccount Id:\t%s\n\tEmail Addr:\t%s", user.DisplayName, user.AccountID, user.EmailAddress)

	return nil
}

func (bot *Bot) ValidateJiraConfig() error {
	if bot.JiraClient == nil || bot.JiraConfig.ProjectKey == "" {
		return errors.Errorf("Jira Client or Key is nil")
	}
	projects, _, err := bot.JiraClient.Project.GetList()
	if err != nil {
		return errors.Errorf("Error getting projects from jira client")
	}

	var proj *jira.Project
	for _, project := range *projects {
		if project.ID == bot.JiraConfig.ProjectKey || project.Key == bot.JiraConfig.ProjectKey {
			proj, _, err = bot.JiraClient.Project.Get(project.ID)
			if err != nil {
				return err
			}
		}
	}
	found := false
	available := make([]string, 0)
	for _, issueType := range proj.IssueTypes {
		available = append(available, issueType.Name)
		//util.Logger().Debugf("%s == %s? %s == %s", bot.JiraConfig.IssueType.Name, issueType.Name, bot.JiraConfig.IssueType.Name, issueType.ID)
		if bot.JiraConfig.IssueType.Name == issueType.Name || bot.JiraConfig.IssueType.Name == issueType.ID {
			found = true
			bot.JiraConfig.IssueType = issueType
		}
	}
	if !found {
		return errors.Errorf("Your Issue Type (%s) listed was not found in your project. Available are %q", bot.JiraConfig.IssueType.Name, available)
	}
	util.Logger().Infof("Successfully validated Jira Config")
	return nil
}
func truncateString(str string, num int) string {
	bnoden := str
	if len(str) > num {
		bnoden = str[0:num]
	}
	return bnoden
}
