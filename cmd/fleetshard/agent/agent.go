package agent

import (
	"github.com/spf13/cobra"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/cmd/fleetshard/agent/run"
)

func NewCamelCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:   "agent",
		Short: "agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	cmd.AddCommand(run.NewRunCmd())

	return &cmd
}
