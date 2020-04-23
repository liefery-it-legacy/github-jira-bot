package run

import (
	"fmt"
	"github.com/Benbentwo/github-jira-bot/pkg/bot"
	"github.com/Benbentwo/github-jira-bot/pkg/github"
	"github.com/Benbentwo/github-jira-bot/pkg/jira"
	"github.com/Benbentwo/utils/util"
	"io/ioutil"
	"sigs.k8s.io/yaml"
)

type BotConfig struct {
	//NOTE this represents <Org/Repo>
	Repo     string `json:"repo,omitempty"`
	Action   string `json:"action,omitempty"`
	Title    string `json:"title,omitempty"`
	Author   string `json:"author,omitempty"`
	PrNumber int    `json:"prNumber,omitempty"`

	Comment   string `json:"comment,omitempty"`
	CommentId int    `json:"commentId,omitempty"`

	MagicQAWord  string            `json:"magicQaWord,omitempty"`
	MaxDesc      int               `json:"maxDescriptionLength,omitempty"`
	ComponentMap map[string]string `json:"componentMap,omitempty"`
	GithubLogin  string            `json:"githubLogin,omitempty"`

	JiraConfig   jira.JiraConfig `json:"jiraConfig,omitempty"`
	GithubConfig github.GHConfig `json:"githubConfig,omitempty"`
}

type FileSaver struct {
	FileName string
}

func (s *FileSaver) SaveConfig(config *BotConfig) error {
	fileName := s.FileName
	if fileName == "" {
		return fmt.Errorf("no filename defined")
	}
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(fileName, data, util.DefaultWritePermissions)
}

func (s *FileSaver) LoadConfig() (*BotConfig, error) {
	config := &BotConfig{}
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

func (botConfig *BotConfig) CreateBotFromConfig() *bot.Bot {
	bot := &bot.Bot{
		Repo:         botConfig.Repo,
		Action:       botConfig.Action,
		PrTitle:      botConfig.Title,
		PrNumber:     botConfig.PrNumber,
		PrAuthor:     botConfig.Author,
		Comment:      botConfig.Comment,
		CommentId:    botConfig.CommentId,
		MagicQAWord:  botConfig.MagicQAWord,
		MaxDesc:      botConfig.MaxDesc,
		ComponentMap: botConfig.ComponentMap,
		GithubLogin:  botConfig.GithubLogin,
		JiraConfig:   botConfig.JiraConfig,
		GithubConfig: botConfig.GithubConfig,
	}
	return bot
}
