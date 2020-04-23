package create

import (
	"github.com/Benbentwo/github-jira-bot/pkg/cmd/common"
	"github.com/spf13/cobra"
)

// options for the command
type CreateOptions struct {
	*common.CommonOptions
	batch bool
}

func NewCmdCreate(commonOpts *common.CommonOptions) *cobra.Command {
	options := &CreateOptions{
		CommonOptions: commonOpts,
	}

	cmd := &cobra.Command{
		Use:     "create",
		Short:   "Create Menu",
		Long:    `Allows user to create configs and other objects`,
		Example: `github-jira-bot create`,
		Run: func(cmd *cobra.Command, args []string) {
			options.Cmd = cmd
			options.Args = args
			err := options.Run()
			common.CheckErr(err)
		},
	}
	// commonOpts.AddBaseFlags(cmd)

	cmd.AddCommand(NewCmdCreateBot(commonOpts))
	return cmd
}

// Run implements this command
func (o *CreateOptions) Run() error {
	return o.Cmd.Help()
}
func (o *CreateOptions) AddCreateFlags(cmd *cobra.Command) {
	o.Cmd = cmd
}
