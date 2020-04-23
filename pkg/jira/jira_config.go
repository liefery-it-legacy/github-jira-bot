package jira

import (
	"fmt"
	"github.com/Benbentwo/utils/util"
	"github.com/andygrunwald/go-jira"
	"io/ioutil"
	"sigs.k8s.io/yaml"
)

type JiraConfig struct {
	ProjectKey string          `json:"projectKey,omitempty"`
	IssueType  jira.IssueType  `json:"issueType,omitempty"`
	FixVersion jira.FixVersion `json:"fixVersion,omitempty"`
	Transition jira.Transition `json:"transition,omitempty"`

	JiraUser            string `json:"user,omitempty"`
	JiraToken           string `json:"token,omitempty"`
	JiraUrl             string `json:"url,omitempty"`
	FixVersionId        string `json:"fixVersionId,omitempty"`
	SprintField         string `json:"sprintField,omitempty"`
	NewTicketTransition int    `json:"newTicketTransition,omitempty"`
}

type FileSaver struct {
	FileName string
}

func (s *FileSaver) SaveConfig(config *JiraConfig) error {
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

func (s *FileSaver) LoadConfig() (*JiraConfig, error) {
	config := &JiraConfig{}
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
