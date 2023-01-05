package camel

import (
	camelv1 "github.com/apache/camel-k/pkg/apis/camel/v1"
	camelv1alpha1 "github.com/apache/camel-k/pkg/apis/camel/v1alpha1"
	"github.com/spf13/cobra"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/cmd/fleetshard/camel/run"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/cos/fleetshard"

	cosv2 "gitub.com/lburgazzoli/bf2-cos-fleetshard-go/apis/cos/v2"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

func init() {
	utilruntime.Must(cosv2.AddToScheme(fleetshard.Scheme))
	utilruntime.Must(camelv1alpha1.AddToScheme(fleetshard.Scheme))
	utilruntime.Must(camelv1.AddToScheme(fleetshard.Scheme))
}

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
