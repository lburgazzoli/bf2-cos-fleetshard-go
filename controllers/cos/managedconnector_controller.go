package cos

import (
	"context"
	"github.com/go-logr/logr"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/controller"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/resources"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sort"
	"time"

	cos "gitub.com/lburgazzoli/bf2-cos-fleetshard-go/apis/cos/v2"
	cosmeta "gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/cos/fleetshard/meta"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// ManagedConnectorReconciler reconciles a ManagedConnector object
type ManagedConnectorReconciler struct {
	client.Client
	Scheme  *runtime.Scheme
	mgr     manager.Manager
	options controller.Options
	l       logr.Logger
}

func NewManagedConnectorReconciler(mgr manager.Manager, options controller.Options) (*ManagedConnectorReconciler, error) {
	r := &ManagedConnectorReconciler{
		Client:  mgr.GetClient(),
		Scheme:  mgr.GetScheme(),
		mgr:     mgr,
		options: options,
		l:       log.Log.WithName("connector-reconciler"),
	}

	return r, r.initialize(mgr)
}

//+kubebuilder:rbac:groups=cos.bf2.dev,resources=managedconnectors,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cos.bf2.dev,resources=managedconnectors/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cos.bf2.dev,resources=managedconnectors/finalizers,verbs=update
//+kubebuilder:rbac:groups=camel.apache.org,resources=kameletbindings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete

func (r *ManagedConnectorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)
	l.Info("Reconciling")

	rc := controller.ReconciliationContext{
		C:      ctx,
		M:      r.mgr,
		Client: r.Client,
		NamespacedName: types.NamespacedName{
			Name:      req.Name,
			Namespace: req.Namespace,
		},
		Connector: &cos.ManagedConnector{
			ObjectMeta: metav1.ObjectMeta{
				Name:      req.Name,
				Namespace: req.Namespace,
			},
		},
		Secret: &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      req.Name + "-deploy",
				Namespace: req.Namespace,
			},
		},
		ConfigMap: &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      req.Name + "-deploy",
				Namespace: req.Namespace,
			},
		},
	}

	if err := resources.Get(ctx, r, rc.Connector); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	if err := resources.Get(ctx, r, rc.Secret); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	if err := resources.Get(ctx, r, rc.ConfigMap); err != nil {
		if k8serrors.IsNotFound(err) {
			if err := r.Create(ctx, rc.ConfigMap); err != nil {
				return ctrl.Result{}, err
			}
		} else {
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
	}

	// ignore notification for resources not with the same UOW.
	// schedule a new reconcile event after a second.

	if rc.Connector.Annotations[cosmeta.MetaUnitOfWork] != rc.Secret.Annotations[cosmeta.MetaUnitOfWork] {
		return ctrl.Result{
			RequeueAfter: 1 * time.Second,
		}, nil
	}

	//
	// Reconcile
	//

	// copy for patch
	//orig := rc.Connector.DeepCopy()

	rc.Connector = rc.Connector.DeepCopy()
	rc.Secret = rc.Secret.DeepCopy()
	rc.ConfigMap = rc.ConfigMap.DeepCopy()

	if update, err := r.options.Reconciler.ApplyFunc(rc); err != nil && !update {
		return ctrl.Result{}, err
	}

	//
	// Update connector
	//

	// TODO: must be properly computed or removed
	rc.Connector.Status.Phase = "Unknown"

	//if err := resources.PatchStatus(ctx, r.Client, &connector, rc.Connector); err != nil {
	// resources.PatchStatus(ctx, r.Client, &orig, rc.Connector)

	sort.SliceStable(rc.Connector.Status.Conditions, func(i, j int) bool {
		return rc.Connector.Status.Conditions[i].Type < rc.Connector.Status.Conditions[j].Type
	})

	if err := r.Status().Update(ctx, rc.Connector); err != nil {
		if k8serrors.IsConflict(err) {
			l.Info(err.Error())
			return ctrl.Result{Requeue: true}, nil
		}

		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}
