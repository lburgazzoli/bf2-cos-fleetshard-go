package run

import (
	camelv1 "github.com/apache/camel-k/pkg/apis/camel/v1"
	camelv1alpha1 "github.com/apache/camel-k/pkg/apis/camel/v1alpha1"
	"github.com/spf13/cobra"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/cos/agent"

	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/controller"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/cos/fleetshard"

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
		ProbeAddr:                     ":8081",
		ProofAddr:                     "",
		EnableLeaderElection:          false,
		ReleaseLeaderElectionOnCancel: true,
		Reconciler: controller.Reconciler{
			Owned:     []client.Object{&camelv1alpha1.KameletBinding{}},
			ApplyFunc: nil,
		},
	}

	cmd := cobra.Command{
		Use:   "run",
		Short: "run",
		RunE: func(cmd *cobra.Command, args []string) error {
			return agent.Start(options)
		},
	}

	cmd.Flags().StringVar(&options.MetricsAddr, "metrics-bind-address", options.MetricsAddr, "The address the metric endpoint binds to.")
	cmd.Flags().StringVar(&options.ProbeAddr, "health-probe-bind-address", options.ProbeAddr, "The address the probe endpoint binds to.")
	cmd.Flags().StringVar(&options.ProofAddr, "pprof-bind-address", options.ProofAddr, "The address the pprof endpoint binds to.")
	cmd.Flags().BoolVar(&options.EnableLeaderElection, "leader-election", options.EnableLeaderElection, "Enable leader election for controller manager.")
	cmd.Flags().BoolVar(&options.ReleaseLeaderElectionOnCancel, "leader-election-release", options.ReleaseLeaderElectionOnCancel, "If the leader should step down voluntarily.")

	return &cmd
}
