package fleetshard

import (
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/controllers/cos"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/controller"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/logger"
	"net/http"
	"os"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	// enable pprof
	_ "net/http/pprof"

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

func Start(options controller.Options) error {
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&logger.Options)))

	ctx := ctrl.SetupSignalHandler()

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                        Scheme,
		MetricsBindAddress:            options.MetricsAddr,
		HealthProbeBindAddress:        options.ProbeAddr,
		LeaderElection:                options.EnableLeaderElection,
		LeaderElectionID:              options.ID + "." + options.Group,
		LeaderElectionReleaseOnCancel: options.ReleaseLeaderElectionOnCancel,
	})
	if err != nil {
		Log.Error(err, "unable to create manager")
		os.Exit(1)
	}

	if _, err := cos.NewManagedConnectorReconciler(mgr, options); err != nil {
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

	if options.ProofAddr != "" {
		Log.Info("starting pprof")

		go func() {
			err := http.ListenAndServe(options.ProofAddr, nil)
			Log.Error(err, "pprof")
		}()
	}

	Log.Info("starting manager")

	if err := mgr.Start(ctx); err != nil {
		Log.Error(err, "problem running manager")
		return err
	}

	return nil
}
