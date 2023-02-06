package run

import (
	"github.com/spf13/cobra"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/controllers/cos"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	cosv2 "gitub.com/lburgazzoli/bf2-cos-fleetshard-go/apis/cos/v2"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/controller"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

var (
	Log = ctrl.Log.WithName("controller-cluster")
)

func init() {
	utilruntime.Must(cosv2.AddToScheme(controller.Scheme))
}

func NewRunCmd() *cobra.Command {
	options := controller.Options{
		MetricsAddr:                   ":8080",
		ProbeAddr:                     ":8081",
		ProofAddr:                     "",
		EnableLeaderElection:          true,
		ReleaseLeaderElectionOnCancel: true,
		Group:                         "cos.bf2.dev",
		ID:                            "",
		Version:                       "",
		Type:                          "",
	}

	cmd := cobra.Command{
		Use:   "run",
		Short: "run",
		RunE: func(cmd *cobra.Command, args []string) error {
			return controller.Start(options, func(manager manager.Manager, opts controller.Options) error {
				if _, err := cos.NewManagedConnectorClusterReconciler(manager, opts); err != nil {
					ctrl.Log.WithName("controller-camel").Error(
						err,
						"unable to create controller",
						"controller", cosv2.GroupVersion.String()+":ManagedConnectorCluster")

					return err
				}

				return nil
			})
		},
	}

	cmd.Flags().StringVar(&options.ID, "operator-id", options.ID, "The ID of the operator.")
	cmd.Flags().StringVar(&options.Group, "operator-group", options.Group, "The group of the operator.")
	cmd.Flags().StringVar(&options.MetricsAddr, "metrics-bind-address", options.MetricsAddr, "The address the metric endpoint binds to.")
	cmd.Flags().StringVar(&options.ProbeAddr, "health-probe-bind-address", options.ProbeAddr, "The address the probe endpoint binds to.")
	cmd.Flags().StringVar(&options.ProofAddr, "pprof-bind-address", options.ProofAddr, "The address the pprof endpoint binds to.")
	cmd.Flags().BoolVar(&options.EnableLeaderElection, "leader-election", options.EnableLeaderElection, "Enable leader election for controller manager.")
	cmd.Flags().BoolVar(&options.ReleaseLeaderElectionOnCancel, "leader-election-release", options.ReleaseLeaderElectionOnCancel, "If the leader should step down voluntarily.")

	_ = cmd.MarkFlagRequired("operator-id")
	_ = cmd.MarkFlagRequired("operator-group")

	return &cmd
}
