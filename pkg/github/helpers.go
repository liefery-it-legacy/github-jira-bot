package github

import (
	"context"
	"github.com/Benbentwo/utils/util"
	"github.com/go-errors/errors"
	"github.com/google/go-github/github"
	"log"
	"os"
	"strings"
)

func UpdateTitle(client *github.Client, owner string, repo string, issueNumber int, title string) (*github.Issue, error) {
	input := &github.IssueRequest{Title: &title}
	issue, _, err := client.Issues.Edit(context.Background(), owner, repo, issueNumber, input)
	if err != nil {
		util.Logger().Errorf("Edit Issue Failed an issue: %s", err)
		return nil, err
	}

	return issue, nil
}

func GetToken(envVar string) (string, error) {
	value := ""
	if os.Getenv(envVar) == "" {
		log.Printf("Enviornment Variable %s not found, checking default `GITHUB_AUTH_TOKEN`", envVar)
		value = os.Getenv("GITHUB_AUTH_TOKEN")
		if value == "" {
			return value, errors.Errorf("Default env variable %s for token not set", "GITHUB_AUTH_TOKEN")
		}
		return value, nil
	}
	return os.Getenv(envVar), nil
}

func SeparateOrgAndRepo(togetherStr string) (string, string, error) {
	arr := strings.Split(togetherStr, "/")
	if len(arr) != 2 {
		return "", "", errors.Errorf("Error splitting the string %s into two strings by `/`", togetherStr)
	}

	return arr[0], arr[1], nil
}
