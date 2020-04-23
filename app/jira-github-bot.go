// +build !windows

package app

import (
	"github.com/Benbentwo/github-jira-bot/pkg/cmd"
	"os"
)

// Run runs the command, if args are not nil they will be set on the command
func Run(args []string) error {
	cmd := cmd.NewJiraGithubBotCommand(os.Stdin, os.Stdout, os.Stderr, nil)
	if args != nil {
		args = args[1:]
		cmd.SetArgs(args)
	}
	return cmd.Execute()
}
