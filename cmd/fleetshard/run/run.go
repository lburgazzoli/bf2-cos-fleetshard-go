package run

import (
	camelv1 "github.com/apache/camel-k/pkg/apis/camel/v1"
	camelv1alpha1 "github.com/apache/camel-k/pkg/apis/camel/v1alpha1"
	"github.com/spf13/cobra"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/controllers/cos"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	cosv2 "gitub.com/lburgazzoli/bf2-cos-fleetshard-go/apis/cos/v2"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/controller"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

func init() {
	// change the scheme defined for camel from camel.apache.org to cos.bf2.dev
	camelv1alpha1.SchemeGroupVersion = schema.GroupVersion{Group: "cos.bf2.dev", Version: "v2"}
	camelv1.SchemeGroupVersion = schema.GroupVersion{Group: "cos.bf2.dev", Version: "v2"}

	// register the schemes
	utilruntime.Must(cosv2.AddToScheme(controller.Scheme))
	utilruntime.Must(camelv1alpha1.AddToScheme(controller.Scheme))
	utilruntime.Must(camelv1.AddToScheme(controller.Scheme))
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
		Type:                          "",
	}

	cmd := cobra.Command{
		Use:   "run",
		Short: "run",
		RunE: func(cmd *cobra.Command, args []string) error {
			return controller.Start(options, func(manager manager.Manager, opts controller.Options) error {
				// camel-k
				if err := cos.SetupCamelControllers(manager, options); err != nil {
					ctrl.Log.WithName("camel-k").Error(
						err,
						"unable to set-up camel-k controllers")

					return err
				}

				// cos
				if err := cos.SetupManagedConnectorClusterReconciler(manager, opts); err != nil {
					ctrl.Log.WithName("managed-connector-cluster").Error(
						err,
						"unable to create controller",
						"gvk", cosv2.GroupVersion.String()+":ManagedConnectorCluster")

					return err
				}
				if err := cos.SetupManagedConnectorReconciler(manager, opts); err != nil {
					ctrl.Log.WithName("managed-connector").Error(
						err,
						"unable to create controller",
						"gvk", cosv2.GroupVersion.String()+":ManagedConnector")

					return err
				}

				return nil
			})
		},
	}

	cmd.Flags().StringVar(&options.ID, "operator-id", options.ID, "The ID of the operator.")
	cmd.Flags().StringVar(&options.Group, "operator-group", options.Group, "The group of the operator.")
	cmd.Flags().StringVar(&options.Type, "operator-type", options.Type, "The type of the operator.")
	cmd.Flags().StringVar(&options.MetricsAddr, "metrics-bind-address", options.MetricsAddr, "The address the metric endpoint binds to.")
	cmd.Flags().StringVar(&options.ProbeAddr, "health-probe-bind-address", options.ProbeAddr, "The address the probe endpoint binds to.")
	cmd.Flags().StringVar(&options.ProofAddr, "pprof-bind-address", options.ProofAddr, "The address the pprof endpoint binds to.")
	cmd.Flags().BoolVar(&options.EnableLeaderElection, "leader-election", options.EnableLeaderElection, "Enable leader election for controller manager.")
	cmd.Flags().BoolVar(&options.ReleaseLeaderElectionOnCancel, "leader-election-release", options.ReleaseLeaderElectionOnCancel, "If the leader should step down voluntarily.")

	_ = cmd.MarkFlagRequired("operator-id")
	_ = cmd.MarkFlagRequired("operator-group")
	_ = cmd.MarkFlagRequired("operator-type")
	_ = cmd.MarkFlagRequired("operator-version")

	return &cmd
}
