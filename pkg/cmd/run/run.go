package run

import (
	"github.com/Benbentwo/github-jira-bot/pkg/cmd/common"
	"github.com/Benbentwo/github-jira-bot/pkg/cmd/create"
	"github.com/Benbentwo/utils/log"
	"github.com/Benbentwo/utils/util"
	"github.com/go-errors/errors"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

// options for the command
type RunOptions struct {
	*common.CommonOptions
	FromFile string
	EnvFile  string
}

func NewCmdRun(commonOpts *common.CommonOptions) *cobra.Command {
	options := &RunOptions{
		CommonOptions: commonOpts,
	}

	cmd := &cobra.Command{
		Use:     "run",
		Short:   "execute the main function of this bot",
		Long:    `Similar to ruby lib/run.rb on the repo https://github.com/liefery/github-jira-bot, this command is the main runner`,
		Example: `github-jira-bot run`,
		Run: func(cmd *cobra.Command, args []string) {
			options.Cmd = cmd
			options.Args = args
			err := options.Run()
			common.CheckErr(err)
		},
	}

	cmd.Flags().StringVarP(&options.FromFile, "from-file", "f", "", "Load values from a file.")
	cmd.Flags().StringVarP(&options.EnvFile, "env-file", "e", "", "Load mock env values from a file, mocks webhook variables.")

	return cmd
}

// Run implements this command
func (o *RunOptions) Run() error {

	newConfig, err := create.GetConfigFromFile(o.FromFile)
	if err != nil {
		return errors.Errorf("Error loading Run Config From file %s: %s", o.FromFile, err)
	}

	robot := newConfig.CreateBotFromRunConfig()

	if o.EnvFile != "" {
		ex, err := util.FileExists(util.HomeReplace(o.EnvFile))
		if err != nil {
			return errors.Errorf("Error finding file %s, does it exist? %s", o.EnvFile, err)
		}
		if ex {
			err = godotenv.Load(o.EnvFile)
			if err != nil {
				return errors.Errorf("Error loading env file %s: %s", o.EnvFile, err)
			}
		}
	}

	robot.SetKeyValuePairsFromEnv()
	if robot.GithubConfig.Username == "" {
		robot.GithubConfig.Username = robot.GithubLogin
	}

	err = robot.CreateGithubClient(robot.GithubConfig.Token, robot.GithubConfig.Enterprise, robot.GithubConfig.Url)
	if err != nil {
		return err
	}
	err = robot.CreateJiraClient(robot.JiraConfig.JiraUrl, newConfig.JiraConfig.JiraUser, []byte(newConfig.JiraConfig.JiraToken))
	if err != nil {
		return err
	}

	err = robot.ValidateGithubClient()
	if err != nil {
		return err
	}
	err = robot.ValidateJiraClient()
	if err != nil {
		return err
	}

	err = robot.ValidateJiraConfig()
	if err != nil {
		return err
	}

	log.Logger().Debugf("Debugging Robot Values:\n\tAction\t%s\n\tTitle\t%s\n\tComment\t\t%s\n\tPR Number\t%d\n\tAuthor\t%s\n\tComment Id\t%d", robot.Action, robot.PrTitle, robot.Comment, robot.PrNumber, robot.PrAuthor, robot.CommentId)

	if issueComment(robot.Action, robot.IssueTitle, robot.Comment, robot.CommentId, robot.CommentAuthor, robot.IssueNumber) {
		util.Logger().Debug("Running Issue Comment")
		robot.RunComment()
	} else if pullRequest(robot.Action, robot.PrTitle, robot.PrNumber) {
		util.Logger().Debug("Running Pull Request")
		robot.RunPullRequest()
	} else {
		log.Logger().Fatal("Robot is not sure what to do " + util.ColorInfo("¯\\_(ツ)_/¯"))
	}

	return nil
}

// def issue_comment?(action, title, comment, pr_number, author, comment_id)
// !action.empty? && !title.empty? && !comment.empty? && !pr_number.empty? && !author.empty? && !comment_id.empty?
// end
//
func issueComment(action string, title string, comment string, prNumber int, author string, commentId int) bool {
	//util.Logger().Debugf("Issue Comment: %s\t%s\t%s\t%d\t%s\t%d", action, title, comment, prNumber, author, commentId)
	return action != "" &&
		title != "" &&
		comment != "" &&
		prNumber > 0 &&
		author != "" &&
		commentId > 0
}

// def pull_request?(action, title, pr_number)
// !action.empty? && !title.empty? && !pr_number.empty?
// end
func pullRequest(action string, title string, prNumber int) bool {
	return action != "" && title != "" && prNumber > 0
}
