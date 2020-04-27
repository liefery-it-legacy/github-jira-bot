package github

import (
	"context"
	"github.com/go-errors/errors"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"log"
	"regexp"
)

const enterpriseEndpoint = "/api/v3/"

func CreateClient(token string, enterprise bool, enterpriseUrl string) (*github.Client, error) {
	if enterprise {
		re := regexp.MustCompile(`github.([a-z]+).com`)
		if !re.MatchString(enterpriseUrl) {
			return nil, errors.New("Enterprise Url is not valid, regex to check `github.([a-z]+).com`")
		}
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	var client *github.Client
	var err error
	if enterprise {
		//baseURL := "https://custom-url/api/v3/"
		baseUrl := "https://" + enterpriseUrl + enterpriseEndpoint
		//uploadURL := "https://custom-upload-url/api/v3/"
		client, err = github.NewEnterpriseClient(baseUrl, baseUrl, tc)
		//github.NewEnterpriseClient()
		if err != nil {
			log.Fatal(err)
		}

	} else {
		client = github.NewClient(tc)
	}

	return client, nil

}
