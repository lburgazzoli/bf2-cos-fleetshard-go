package camel

import (
	"github.com/spf13/cobra"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/cmd/fleetshard/camel/run"
)

func NewCamelCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:   "camel",
		Short: "camel",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	cmd.AddCommand(run.NewRunCmd())

	return &cmd
}
