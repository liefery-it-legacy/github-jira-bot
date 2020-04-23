package github

import (
	"fmt"
	"github.com/Benbentwo/utils/util"
	"io/ioutil"
	"sigs.k8s.io/yaml"
)

type GHConfig struct {
	Token      string `json:"token,omitempty"`
	Enterprise bool   `json:"enterprise"`
	Url        string `json:"url"`
	Username   string `json:"username"`
}

type FileSaver struct {
	FileName string
}

func (s *FileSaver) SaveConfig(config *GHConfig) error {
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

func (s *FileSaver) LoadConfig() (*GHConfig, error) {
	config := &GHConfig{}
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
