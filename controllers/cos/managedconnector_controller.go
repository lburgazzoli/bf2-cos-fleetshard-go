package cos

import (
	"context"
	"github.com/go-logr/logr"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/controller"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/resources"
	"sort"
	"time"

	cos "gitub.com/lburgazzoli/bf2-cos-fleetshard-go/apis/cos/v2"
	cosmeta "gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/cos/fleetshard/meta"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/internal/camel"
)

// ManagedConnectorReconciler reconciles a ManagedConnector object
type ManagedConnectorReconciler struct {
	client.Client
	mgr     manager.Manager
	options controller.Options
	l       logr.Logger
}

func SetupManagedConnectorReconciler(mgr manager.Manager, options controller.Options) error {
	r := &ManagedConnectorReconciler{
		Client:  mgr.GetClient(),
		mgr:     mgr,
		options: options,
		l:       log.Log.WithName("connector-reconciler"),
	}

	return r.initialize(mgr)
}

func (r *ManagedConnectorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)
	l.Info("Reconciling")

	rc := controller.ReconciliationContext{
		C:      ctx,
		Client: r.Client,
		NamespacedName: types.NamespacedName{
			Name:      req.Name,
			Namespace: req.Namespace,
		},
		Connector: &cos.ManagedConnector{
			ObjectMeta: resources.AsObjectMeta(req.NamespacedName),
		},
		Secret: &corev1.Secret{
			ObjectMeta: resources.AsObjectMeta(req.NamespacedName),
		},
		ConfigMap: &corev1.ConfigMap{
			ObjectMeta: resources.AsObjectMeta(req.NamespacedName),
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
			return ctrl.Result{}, err
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
	// orig := rc.Connector.DeepCopy()

	rc.Connector = rc.Connector.DeepCopy()
	rc.Secret = rc.Secret.DeepCopy()
	rc.ConfigMap = rc.ConfigMap.DeepCopy()

	if update, err := camel.Apply(rc); err != nil && !update {
		return ctrl.Result{}, err
	}

	//
	// Update connector
	//

	// TODO: must be properly computed or removed
	rc.Connector.Status.Phase = "Unknown"

	sort.SliceStable(rc.Connector.Status.Conditions, func(i, j int) bool {
		return rc.Connector.Status.Conditions[i].Type < rc.Connector.Status.Conditions[j].Type
	})

	err := r.Status().Update(ctx, rc.Connector)
	if err != nil && k8serrors.IsConflict(err) {
		l.Info(err.Error())
		return ctrl.Result{Requeue: true}, nil
	}

	return ctrl.Result{}, err
}
