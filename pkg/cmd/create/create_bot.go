package create

import (
	"fmt"
	"github.com/Benbentwo/github-jira-bot/pkg/bot"
	"github.com/Benbentwo/github-jira-bot/pkg/cmd/common"
	"github.com/Benbentwo/github-jira-bot/pkg/github"

	"github.com/Benbentwo/utils/util"
	"github.com/go-errors/errors"
	"github.com/spf13/cobra"
	"gopkg.in/AlecAivazis/survey.v1"
	"io/ioutil"
	"os"
	"regexp"
	"sigs.k8s.io/yaml"
	"strconv"
)

const (
	DefaultMagicWord  = "QA:"
	DefaultConfigDir  = "config"
	DefaultConfigFile = "config.yaml"

	EnterpriseGithubServerRegex = `github.(.\w+.).com`
)

var (
	levels = util.GetLevels()
	_      = util.SetLevel(levels[0])
	logger = util.Logger().Logger
)

// options for the command
type CreateBotOptions struct {
	*common.CommonOptions
	batch    bool
	FileName string

	runConfig RunConfig

	AskEverything bool
	FromFile      string
	// These are for sample env - they mock the payload of the webhook
	// Repo		string
	// PrTitle		string
	// PrNumber	int
	// PrAuthor		string
}

type RunJiraConfig struct {
	JiraUser            string `json:"user,omitempty"`
	JiraToken           string `json:"token,omitempty"`
	JiraUrl             string `json:"url,omitempty"`
	ProjectKey          string `json:"projectKey,omitempty"`
	DefaultIssueType    string `json:"defaultIssueType,omitempty"`
	FixVersionId        string `json:"fixVersionId,omitempty"`
	SprintField         string `json:"sprintField,omitempty"`
	NewTicketTransition int    `json:"newTicketTransition,omitempty"`
}

type RunGithubConfig struct {
	Enterprise bool   `json:"enterprise"`
	Url        string `json:"url"` // should default to github.com if not enterprise
	Username   string `json:"username,omitempty"`
	Token      string `json:"token,omitempty"`
}

type RunConfig struct {
	JiraConfig   RunJiraConfig     `json:"jira,omitempty"`
	GithubConfig RunGithubConfig   `json:"github,omitempty"`
	MagicQAWord  string            `json:"magicQaWord,omitempty"`
	MaxLength    int               `json:"maxLength,omitempty"`
	ComponentMap map[string]string `json:"componentMap,omitempty"`
}

type FileSaver struct {
	FileName string
}

func (s *FileSaver) SaveConfig(config *RunConfig) error {
	fileName := s.FileName
	if fileName == "" {
		return fmt.Errorf("no filename defined")
	}
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	util.Debug("Data to save: %s", data)
	return ioutil.WriteFile(fileName, data, util.DefaultWritePermissions)
}

func (s *FileSaver) LoadConfig() (*RunConfig, error) {
	config := &RunConfig{}
	fileName := s.FileName
	if fileName != "" {
		exists, err := util.FileExists(fileName)
		if err != nil {
			return config, fmt.Errorf("Could not check if file exists %s due to %s", fileName, err)
		}
		if exists {
			data, err := ioutil.ReadFile(fileName)
			if err != nil {
				return config, fmt.Errorf("Failed to load file %s due to %s", fileName, err)
			}
			err = yaml.Unmarshal(data, config)
			if err != nil {
				return config, fmt.Errorf("Failed to unmarshal YAML file %s due to %s", fileName, err)
			}
		}
	}
	return config, nil
}

func (runConfig *RunConfig) ClearConfig() {
	runConfig.MaxLength = -1
	runConfig.MagicQAWord = ""
	runConfig.GithubConfig.Url = ""
	runConfig.GithubConfig.Token = ""
	runConfig.GithubConfig.Username = ""
	runConfig.GithubConfig.Enterprise = false
	runConfig.JiraConfig.NewTicketTransition = -1
	runConfig.JiraConfig.SprintField = ""
	runConfig.JiraConfig.FixVersionId = ""
	runConfig.JiraConfig.DefaultIssueType = ""
	runConfig.JiraConfig.ProjectKey = ""
	runConfig.JiraConfig.JiraUser = ""
	runConfig.JiraConfig.JiraUrl = ""
	runConfig.JiraConfig.JiraToken = ""
}

func NewCmdCreateBot(commonOpts *common.CommonOptions) *cobra.Command {
	options := &CreateBotOptions{
		CommonOptions: commonOpts,
	}

	cmd := &cobra.Command{
		Use:     "bot",
		Short:   "create a bot config for quick loading",
		Long:    "create a bot config, similar to https://github.com/liefery/github-jira-bot#configuration",
		Example: "github-jira-bot create bot",
		Run: func(cmd *cobra.Command, args []string) {
			options.Cmd = cmd
			options.Args = args
			err := options.Run()
			common.CheckErr(err)
		},
	}

	cmd.Flags().StringVarP(&options.FromFile, "from-file", "", "", "Load values from a file. Overridden by other options on this command.")

	cmd.Flags().StringVarP(&options.runConfig.JiraConfig.JiraUrl, "jira-url", "", "", "The URL of your Jira workspace.")
	cmd.Flags().StringVarP(&options.runConfig.JiraConfig.JiraUser, "jira-user", "", "", "The Username of a Jira account.")
	cmd.Flags().StringVarP(&options.runConfig.JiraConfig.JiraToken, "jira-token", "", "", "The Token of a Jira account.")
	cmd.Flags().StringVarP(&options.runConfig.JiraConfig.ProjectKey, "jira-project-key", "", "", "Key of a Jira project. This looks something like BT.")
	cmd.Flags().StringVarP(&options.runConfig.JiraConfig.DefaultIssueType, "jira-issue-type", "", "Story", "Can be Story, Improvement or similar. This will only be used if the bot creates tickets.")
	cmd.Flags().StringVarP(&options.runConfig.JiraConfig.FixVersionId, "jira-fix-version-id", "", "", "Id of the fix version to use when creating a Jira ticket.")
	cmd.Flags().StringVarP(&options.runConfig.JiraConfig.SprintField, "jira-sprint-field", "", "", "The identifier of the sprint new tickets should be attached to upon creation.")
	cmd.Flags().IntVarP(&options.runConfig.JiraConfig.NewTicketTransition, "ticket-transition", "", -1, "The database ID of the transition to the state where newly created Jira tickets should be moved to.")

	cmd.Flags().BoolVarP(&options.runConfig.GithubConfig.Enterprise, "enterprise", "e", false, "Is the git server a github enterprise server?")
	cmd.Flags().StringVarP(&options.runConfig.GithubConfig.Url, "git-url", "s", "github.com", "Git server url. Either github.com (default) or github.myCompany.com")
	// TODO add validation for if enterprise that the url is github.X.com
	// 	add validation if enterprise is false that the url is github.com
	cmd.Flags().StringVarP(&options.runConfig.GithubConfig.Username, "git-username", "u", "", "Git user name")
	cmd.Flags().StringVarP(&options.runConfig.GithubConfig.Token, "git-api-token", "p", "", "Git apiToken")

	cmd.Flags().StringVarP(&options.runConfig.MagicQAWord, "magic-word", "m", DefaultMagicWord, "Default Magic Word")
	cmd.Flags().IntVarP(&options.runConfig.MaxLength, "max-length", "l", 600, "The maximum number of chars of a Jira ticket description that will be added to a pull request on GitHub. Omit this environment variable if you want to add the entire ticket description to GitHub.")

	cmd.Flags().StringVarP(&options.FileName, "file", "f", "", "File Name to save the config to. Places in a ./config/<fileName>")

	cmd.Flags().BoolVarP(&options.AskEverything, "ask-everything", "a", false, "Prompt the User for everything? (Clears Defaults)")

	options.runConfig.ComponentMap = make(map[string]string, 0)
	options.runConfig.ComponentMap["repo"] = "component"

	return cmd
}

// Run implements this command
func (o *CreateBotOptions) Run() error {

	if o.AskEverything {
		o.runConfig.ClearConfig()
	}

	if o.FileName != "" {
		newConfig, err := GetConfigFromFile(o.FileName)
		if err != nil {
			return errors.Errorf("Error loading Run Config From file %s: %s", o.FileName, err)
		}
		o.runConfig = *newConfig
	}

	path, _ := os.Getwd()
	path = util.StripTrailingSlash(path) + "/" + util.StripTrailingSlash(DefaultConfigDir)
	exists, err := util.DirExists(path)
	if err != nil {
		return errors.Errorf("Error checking if dir exists %s", err)
	}

	if !exists {
		err = os.MkdirAll(path, util.DefaultWritePermissions)
		if err != nil {
			return errors.Errorf("Error making path %s: %s", path, err)
		}
	}

	if o.FileName == "" {
		util.Logger().Warn("File Name is empty... Changing to default (config.yaml)")
		o.FileName = DefaultConfigFile
	}
	fs := FileSaver{
		FileName: util.StripTrailingSlash(path) + "/" + o.FileName,
	}
	util.Debug("Saving to file %s", util.ColorInfo(fs.FileName))

	if o.batch {
		err = fs.SaveConfig(&o.runConfig)
		if err != nil {
			return errors.Errorf("Error saving config %d to file %s", o.runConfig, fs.FileName)
		}
	}

	err = util.PromptForMissingString(&o.runConfig.JiraConfig.JiraUrl, "Jira Url", "What is the jira server url?", false)
	if err != nil {
		return errors.Errorf("Error Getting %s: %s", "Jira Url", err)
	}

	err = util.PromptForMissingString(&o.runConfig.JiraConfig.JiraUser, "Jira Username", "What is the jira username for the robot?", false)
	if err != nil {
		return errors.Errorf("Error Getting %s: %s", "Jira User", err)
	}

	o.runConfig.JiraConfig.JiraToken, err = util.PromptValuePassword("What is your JIRA API Token?", "Navigate here: \nhttps://id.atlassian.com/manage-profile/security/api-tokens\nAnd create a new token, pasting the key below.")
	if err != nil {
		return errors.Errorf("Error Getting %s: %s", "Jira Api Token", err)
	}

	// TODO these should be pickers so you can select from existing (using above information to fetch)
	err = util.PromptForMissingString(&o.runConfig.JiraConfig.ProjectKey, "Jira Project Key:", "What is the Jira Project key for issues?", false)
	if err != nil {
		return errors.Errorf("Error Getting %s: %s", "Jira Project Key", err)
	}

	err = util.PromptForMissingString(&o.runConfig.JiraConfig.DefaultIssueType, "Default Jira Issue Type", "The bot will create JIRA issues if not found, what type should they be: Story? Improvement?...", false)
	if err != nil {
		return errors.Errorf("Error Getting %s: %s", "Jira Issue Type", err)
	}

	err = util.PromptForMissingString(&o.runConfig.JiraConfig.FixVersionId, "Default Fix Version ID", "Id of the fix version to use when creating a Jira ticket.", false)
	if err != nil {
		return errors.Errorf("Error Getting %s: %s", "Default Fix Version ID", err)
	}

	err = util.PromptForMissingString(&o.runConfig.JiraConfig.SprintField, "Default Sprint Field", "The identifier of the sprint new tickets should be attached to upon creation.", false)
	if err != nil {
		return errors.Errorf("Error Getting %s: %s", "Default Sprint Field", err)
	}

	if o.runConfig.JiraConfig.NewTicketTransition < 0 {
		o.runConfig.JiraConfig.NewTicketTransition, err = promptInt("New Ticket Transition (number)", "", "The database ID of the transition to the state where newly created Jira tickets should be moved to.")
		if err != nil {
			return errors.Errorf("Error Getting %s: %s", "New Ticket Transition", err)
		}
	}

	// --------- Github
	o.runConfig.GithubConfig.Enterprise, err = promptBool("Is this a Github Enterprise Server", false, "Is this a Github Enterprise server, like github.myCompany.com")
	if err != nil {
		return errors.Errorf("Error Getting %s: %s", "Github Enterprise Server Check", err)
	}

	re := regexp.MustCompile(EnterpriseGithubServerRegex)
	util.Debug("%t", !re.MatchString(o.runConfig.GithubConfig.Url))
	util.Debug("%t", o.runConfig.GithubConfig.Enterprise, o.runConfig.GithubConfig.Url == "", re.MatchString(o.runConfig.GithubConfig.Url))
	if o.runConfig.GithubConfig.Enterprise && !re.MatchString(o.runConfig.GithubConfig.Url) {
		question := []*survey.Question{
			{
				Name: "url",
				Prompt: &survey.Input{
					Message: "What is the Git Server Url?",
				},
				Validate: func(val interface{}) error {
					// if the input matches the expectation
					if str := val.(string); !re.MatchString(str) {
						return fmt.Errorf("must match %s%s%s", util.ColorInfo("github."), util.ColorDebug("something"), util.ColorInfo(".com"))
					}
					// nothing was wrong
					return nil
				},
			},
		}
		answers := struct {
			Url string
		}{}
		err := survey.Ask(question, &answers)
		// o.runConfig.GithubConfig.Url, err = survey.
		// err = util.PromptForMissingString(&o.runConfig.GithubConfig.Url, "Git Server Url", "Git", false)
		if err != nil {
			return errors.Errorf("Error Getting %s: %s", "Git Server Url", err)
		}
		o.runConfig.GithubConfig.Url = answers.Url
	}

	// log.Debugf("\t%s:\t%s", "Jira User", o.runConfig.JiraConfig.JiraUser)
	// err = promptForMissingString(&o.runConfig.JiraConfig.JiraUser, "Jira Username", "What is the jira username for the robot?", false)
	// if err != nil { return errors.Errorf("Error Getting %s: %s", "JiraUser", err)}
	//
	err = util.PromptForMissingString(&o.runConfig.GithubConfig.Username, "github Username", "What is the github username for the robot?", false)
	if err != nil {
		return errors.Errorf("Error Getting %s: %s", "Github User", err)
	}
	//

	o.runConfig.GithubConfig.Token, err = PromptValuePassword("What is your Github API Token?", "Navigate here: \n "+util.ColorInfo("https://"+o.runConfig.GithubConfig.Url+"/settings/tokens/new?scopes=repo,read:user,read:org,user:email,write:repo_hook,delete_repo")+"\nAnd create a new token, pasting the key below.", validateGHApiToken)
	if err != nil {
		return errors.Errorf("Error Getting %s: %s", "Github Api Token", err)
	}

	// ----------------- Run Config
	err = util.PromptForMissingString(&o.runConfig.MagicQAWord, "Magic Qa Word:", "What do you wish to be your magic QA Word?", false)
	if err != nil {
		return errors.Errorf("Error Getting %s: %s", "Magic Qa Word", err)
	}

	if o.runConfig.MaxLength < 0 {
		o.runConfig.MaxLength, err = promptInt("Max Character Length on comments or Tickets? (number)", "", "The maximum number of chars of a Jira ticket description that will be added to a pull request on GitHub. Omit this environment variable if you want to add the entire ticket description to GitHub.")
		if err != nil {
			return errors.Errorf("Error Getting %s: %s", "New Ticket Transition", err)
		}
	}

	err = fs.SaveConfig(&o.runConfig)
	if err != nil {
		return errors.Errorf("Errors saving config to file %s, data: %d, error: %s", util.ColorInfo(fs.FileName), util.ColorDebug(&o.runConfig), util.ColorError(err))
	}

	return nil
}

func promptBool(message string, defaultChoice bool, help string) (bool, error) {
	name := defaultChoice
	prompt := &survey.Confirm{
		Message: message,
		Default: defaultChoice,
		Help:    help,
	}

	surveyOpts := survey.WithStdio(os.Stdin, os.Stdout, os.Stderr)
	err := survey.AskOne(prompt, &name, nil, surveyOpts)
	return name, err
}

func promptInt(message string, defaultChoice string, help string) (int, error) {
	name := defaultChoice
	prompt := &survey.Input{
		Message: message,
		Default: defaultChoice,
		Help:    help,
	}

	surveyOpts := survey.WithStdio(os.Stdin, os.Stdout, os.Stderr)
	err := survey.AskOne(prompt, &name, validateInt, surveyOpts)
	value, err := strconv.Atoi(name)
	return value, err
}

var validateInt = func(val interface{}) error {
	_, err := strconv.Atoi(val.(string))
	if err != nil {
		return errors.Errorf("Input was not a number format")
	}
	return nil
}

var validateGHApiToken = func(val interface{}) error {
	if val.(string) == "" {
		return nil
	}
	matched, err := regexp.MatchString(`^[a-fA-F0-9]{40}$`, val.(string))
	if err != nil {
		return errors.Errorf("Input was not a number format")
	}
	if !matched { // allow empty or a token
		return errors.Errorf("Invalid API Token format, regex to validate `^[a-fA-F0-9]{40}$`")
	}
	return nil
}

func PromptValuePassword(message string, help string, validator func(val interface{}) error) (string, error) {
	name := ""
	prompt := &survey.Password{
		Message: message,
		Help:    help,
	}

	surveyOpts := survey.WithStdio(os.Stdin, os.Stdout, os.Stderr)
	err := survey.AskOne(prompt, &name, validator, surveyOpts)
	return name, err
}

func GetConfigFromFile(filePath string) (*RunConfig, error) {
	if filePath == "" {
		return nil, errors.Errorf("Empty File Path.")
	}

	exists, err := util.FileExists(util.HomeReplace(filePath))
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.Errorf("Couldn't find the file you specified to load from, are you sure it exists?")
	}

	fs := FileSaver{
		FileName: filePath,
	}
	newConfig, err := fs.LoadConfig()
	if err != nil {
		return nil, errors.Errorf("Error loading Run Config From file %s: %s", fs.FileName, err)
	}

	return newConfig, nil
}

func (runConfig *RunConfig) CreateBotFromRunConfig() *bot.Bot {
	robot := &bot.Bot{
		MagicQAWord:  runConfig.MagicQAWord,
		MaxDesc:      runConfig.MaxLength,
		ComponentMap: runConfig.ComponentMap,
		GithubLogin:  runConfig.GithubConfig.Username,
		// jiraProjectKey string, jiraIssueType string, jiraFixVersionId string, jiraTransitionId string
		JiraConfig: *bot.NewJiraConfigFromStrings(
			runConfig.JiraConfig.ProjectKey,
			runConfig.JiraConfig.DefaultIssueType,
			runConfig.JiraConfig.FixVersionId,
			strconv.Itoa(runConfig.JiraConfig.NewTicketTransition)),
		GithubConfig: *runConfig.GithubConfig.CreateGHConfigFromRunGHConfig(),
	}
	robot.JiraConfig.JiraUser = runConfig.JiraConfig.JiraUser
	//util.Logger().Debugf("Setting JiraUrl %s=%s", robot.JiraConfig.JiraUrl, runConfig.JiraConfig.JiraUrl)
	robot.JiraConfig.JiraUrl = runConfig.JiraConfig.JiraUrl
	robot.JiraConfig.JiraToken = runConfig.JiraConfig.JiraToken
	robot.JiraConfig.SprintField = runConfig.JiraConfig.SprintField

	return robot

}
func (ghConfig *RunGithubConfig) CreateGHConfigFromRunGHConfig() *github.GHConfig {
	return &github.GHConfig{
		Token:      ghConfig.Token,
		Enterprise: ghConfig.Enterprise,
		Url:        ghConfig.Url,
		Username:   ghConfig.Username,
	}
}
