package jira

import (
	"github.com/andygrunwald/go-jira"
	"log"
	"strings"
)

func GetJiraClient(jiraUrl string, username string, bytePassword []byte) (*jira.Client, error) {
	password := string(bytePassword)

	tp := jira.BasicAuthTransport{
		Username: strings.TrimSpace(username),
		Password: strings.TrimSpace(password),
	}

	client, err := jira.NewClient(tp.Client(), strings.TrimSpace(jiraUrl))
	if err != nil {
		log.Printf("Error Creating Jira Client %s", err)
		return nil, err
	}
	return client, nil
}
