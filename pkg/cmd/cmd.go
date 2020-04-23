
package cmd

import (
	"github.com/Benbentwo/github-jira-bot/pkg/cmd/common"
	"github.com/Benbentwo/github-jira-bot/pkg/cmd/create"
	"github.com/Benbentwo/github-jira-bot/pkg/cmd/run"
	"github.com/spf13/viper"
	"io"
	"strings"

	"github.com/Benbentwo/utils/log"
	"github.com/spf13/cobra"
	"gopkg.in/AlecAivazis/survey.v1/terminal"
)



func NewJiraGithubBotCommand(in terminal.FileReader, out terminal.FileWriter, err io.Writer, args []string) *cobra.Command {

	replacer := strings.NewReplacer("-", "_")
	viper.SetEnvKeyReplacer(replacer)
	cmd := &cobra.Command{
		Use:              "github-jira-bot",
		Short:            "CLI tool to help with Github Jira Connections of issues",
		Long:             "This CLI tool is a rewrite of https://github.com/liefery/github-jira-bot and is intended to be placed on a build pod or attached to a jenkins script in order to help provide the same functionality as the ruby version.",
		PersistentPreRun: common.SetLoggingLevel,
		Run:              runHelp,
	}
	commonOpts := &common.CommonOptions{
		In:  in,
		Out: out,
		Err: err,
	}
	commonOpts.AddBaseFlags(cmd)

	// Section to add commands to:
	cmd.AddCommand(run.NewCmdRun(commonOpts))
	cmd.AddCommand(create.NewCmdCreate(commonOpts))

	return cmd
}

func runHelp(cmd *cobra.Command, args []string) {
	err := cmd.Help()
	if err != nil {
		log.Logger().Errorf("Error running help: %s", err)
	}
}

