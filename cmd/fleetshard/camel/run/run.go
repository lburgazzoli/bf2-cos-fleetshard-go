package run

import (
	camelv1 "github.com/apache/camel-k/pkg/apis/camel/v1"
	camelv1alpha1 "github.com/apache/camel-k/pkg/apis/camel/v1alpha1"
	"github.com/spf13/cobra"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/internal/camel"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/cos/fleetshard"

	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/controller"

	"sigs.k8s.io/controller-runtime/pkg/client"

	cosv2 "gitub.com/lburgazzoli/bf2-cos-fleetshard-go/apis/cos/v2"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

func init() {
	utilruntime.Must(cosv2.AddToScheme(fleetshard.Scheme))
	utilruntime.Must(camelv1alpha1.AddToScheme(fleetshard.Scheme))
	utilruntime.Must(camelv1.AddToScheme(fleetshard.Scheme))
}

func NewRunCmd() *cobra.Command {
	options := controller.Options{
		Group:                         "cos.bf2.dev",
		ID:                            "",
		Type:                          "",
		Version:                       "",
		MetricsAddr:                   ":8080",
		EnableLeaderElection:          false,
		ReleaseLeaderElectionOnCancel: true,
		Reconciler: controller.Reconciler{
			Owned:     []client.Object{&camelv1alpha1.KameletBinding{}},
			ApplyFunc: camel.Apply,
		},
	}

	cmd := cobra.Command{
		Use:   "run",
		Short: "run",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fleetshard.Start(options)
		},
	}

	cmd.Flags().StringVar(&options.ID, "operator-id", options.ID, "The ID of the operator.")
	cmd.Flags().StringVar(&options.Group, "operator-group", options.Group, "The group of the operator.")
	cmd.Flags().StringVar(&options.Type, "operator-type", options.Type, "The type of the operator.")
	cmd.Flags().StringVar(&options.Version, "operator-version", options.Version, "The version of the operator.")
	cmd.Flags().StringVar(&options.MetricsAddr, "metrics-bind-address", options.MetricsAddr, "The address the metric endpoint binds to.")
	cmd.Flags().StringVar(&options.ProbeAddr, "health-probe-bind-address", options.ProbeAddr, "The address the probe endpoint binds to.")
	cmd.Flags().BoolVar(&options.EnableLeaderElection, "leader-election", options.EnableLeaderElection, "Enable leader election for controller manager.")
	cmd.Flags().BoolVar(&options.ReleaseLeaderElectionOnCancel, "leader-election-release", options.ReleaseLeaderElectionOnCancel, "If the leader should step down voluntarily.")

	_ = cmd.MarkFlagRequired("operator-id")
	_ = cmd.MarkFlagRequired("operator-group")
	_ = cmd.MarkFlagRequired("operator-type")
	_ = cmd.MarkFlagRequired("operator-version")

	return &cmd
}
