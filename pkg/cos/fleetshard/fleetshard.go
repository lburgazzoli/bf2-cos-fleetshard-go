package fleetshard

import (
	"flag"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/controllers/cos"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/controller"
	"k8s.io/klog/v2"
	"os"
	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	Scheme = runtime.NewScheme()
	Log    = ctrl.Log.WithName("controller")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(Scheme))
}

func Start(controller controller.Controller) error {
	metricsAddr := ":8080"
	enableLeaderElection := false
	releaseLeaderElectionOnCancel := true
	leaderElectionID := "7157fb2e.cos.bf2.dev"
	probeAddr := ":8081"

	flag.StringVar(&metricsAddr, "metrics-bind-address", metricsAddr, "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", probeAddr, "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-election", enableLeaderElection, "Enable leader election for controller manager.")
	flag.StringVar(&leaderElectionID, "leader-election-id", leaderElectionID, "The leader election id.")
	flag.BoolVar(&releaseLeaderElectionOnCancel, "leader-election-release", releaseLeaderElectionOnCancel, "If the leader should step down voluntarily.")

	opts := zap.Options{Development: true}

	opts.BindFlags(flag.CommandLine)
	klog.InitFlags(flag.CommandLine)

	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	ctx := ctrl.SetupSignalHandler()

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 Scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       leaderElectionID,
		// LeaderElectionReleaseOnCancel defines if the leader should step down voluntarily
		// when the Manager ends. This requires the binary to immediately end when the
		// Manager is stopped, otherwise, this setting is unsafe. Setting this significantly
		// speeds up voluntary leader transitions as the new leader don't have to wait
		// LeaseDuration time first.
		//
		// In the default scaffold provided, the program ends immediately after
		// the manager stops, so would be fine to enable this option. However,
		// if you are doing or is intended to do any operation such as perform cleanups
		// after the manager stops then its usage might be unsafe.
		LeaderElectionReleaseOnCancel: releaseLeaderElectionOnCancel,
	})
	if err != nil {
		Log.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if _, err := cos.NewManagedConnectorReconciler(mgr, controller); err != nil {
		Log.Error(err, "unable to create controller", "controller", "ManagedConnector")
		return err
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		Log.Error(err, "unable to set up health check")
		return err
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		Log.Error(err, "unable to set up ready check")
		return err
	}

	Log.Info("starting manager")

	if err := mgr.Start(ctx); err != nil {
		Log.Error(err, "problem running manager")
		return err
	}

	return nil
}
