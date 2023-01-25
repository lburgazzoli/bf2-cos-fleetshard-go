package cos

import (
	"context"
	"github.com/go-logr/logr"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/controller"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// ManagedConnectorClusterReconciler reconciles a ManagedConnector object
type ManagedConnectorClusterReconciler struct {
	client.Client
	Scheme  *runtime.Scheme
	mgr     manager.Manager
	options controller.Options
	l       logr.Logger
}

func NewManagedConnectorClusterReconciler(mgr manager.Manager, options controller.Options) (*ManagedConnectorClusterReconciler, error) {
	r := &ManagedConnectorClusterReconciler{
		Client:  mgr.GetClient(),
		Scheme:  mgr.GetScheme(),
		mgr:     mgr,
		options: options,
		l:       log.Log.WithName("cluster-reconciler"),
	}

	return r, r.initialize(mgr)
}

//+kubebuilder:rbac:groups=cos.bf2.dev,resources=managedconnectorclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cos.bf2.dev,resources=managedconnectorclusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cos.bf2.dev,resources=managedconnectorclusters/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete

func (r *ManagedConnectorClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)
	l.Info("Reconciling")

	return ctrl.Result{}, nil
}
