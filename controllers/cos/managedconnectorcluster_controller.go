package cos

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	cosv2 "gitub.com/lburgazzoli/bf2-cos-fleetshard-go/apis/cos/v2"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/controller"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/cos/fleetmanager"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/defaults"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/resources"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"net/url"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
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

// +kubebuilder:rbac:groups=cos.bf2.dev,resources=managedconnectorclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cos.bf2.dev,resources=managedconnectorclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cos.bf2.dev,resources=managedconnectorclusters/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete

func (r *ManagedConnectorClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)
	l.Info("Reconciling")

	mcc := &cosv2.ManagedConnectorCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.NamespacedName.Name,
			Namespace: req.NamespacedName.Namespace,
		},
	}

	if err := resources.Get(ctx, r, mcc); err != nil && !k8serrors.IsNotFound(err) {
		r.l.Error(err, "failure getting resource", "ManagedConnectorCluster", req.NamespacedName.String())
	}

	// safety copy
	mcc = mcc.DeepCopy()

	if mcc.ObjectMeta.DeletionTimestamp.IsZero() {

		//
		// Add finalizer
		//

		if !controllerutil.AddFinalizer(mcc, defaults.ConnectorClustersFinalizerName) {
			if err := r.Update(ctx, mcc); err != nil {
				if k8serrors.IsConflict(err) {
					return ctrl.Result{}, err
				}

				return ctrl.Result{}, errors.Wrapf(err, "failure adding finalizer to connector cluster %s", req.NamespacedName)
			}
		}
	} else {

		//
		// Handle deletion
		//

		if controllerutil.ContainsFinalizer(mcc, defaults.ConnectorClustersFinalizerName) {
			controllerutil.RemoveFinalizer(mcc, defaults.ConnectorClustersFinalizerName)

			if err := r.Update(ctx, mcc); err != nil {
				if k8serrors.IsConflict(err) {
					return ctrl.Result{}, err
				}

				return ctrl.Result{}, errors.Wrapf(err, "failure removing finalizer from connector cluster %s", req.NamespacedName)
			}
		}

		return ctrl.Result{}, nil
	}

	mcc.Status.ObservedGeneration = mcc.Generation
	// TODO: add conditions & co
	// TODO: deal with changes

	err := r.run(ctx, mcc)
	if err != nil {
		meta.SetStatusCondition(&mcc.Status.Conditions, metav1.Condition{
			Type:    "Triggered",
			Status:  metav1.ConditionFalse,
			Reason:  "Error",
			Message: err.Error(),
		})
	} else {
		meta.SetStatusCondition(&mcc.Status.Conditions, metav1.Condition{
			Type:    "Triggered",
			Status:  metav1.ConditionTrue,
			Reason:  "Scheduled",
			Message: "Scheduled",
		})
	}

	if err := r.Status().Update(ctx, mcc); err != nil {
		if k8serrors.IsConflict(err) {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, err
	}

	// Poor man scheduler
	return ctrl.Result{RequeueAfter: mcc.Spec.PollDelay.Duration}, nil
}

func (r *ManagedConnectorClusterReconciler) run(ctx context.Context, mcc *cosv2.ManagedConnectorCluster) error {
	cid := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mcc.Spec.Auth.ClientID.SecretKeyRef.Name,
			Namespace: mcc.Namespace,
		},
	}
	cs := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mcc.Spec.Auth.ClientSecret.SecretKeyRef.Name,
			Namespace: mcc.Namespace,
		},
	}

	if err := resources.Get(ctx, r, cid); err != nil {
		return err
	}
	if err := resources.Get(ctx, r, cs); err != nil {
		return err
	}

	cpUrl, err := url.Parse(mcc.Spec.ControlPlaneURL)
	if err != nil {
		return err
	}

	autUrl, err := url.Parse(mcc.Spec.Auth.AuthURL)
	if err != nil {
		return err
	}

	// TODO: cache
	c, err := fleetmanager.NewClient(ctx, fleetmanager.Config{
		ApiURL:       cpUrl,
		AuthURL:      autUrl,
		AuthTokenURL: autUrl.JoinPath("auth", "realms", mcc.Spec.Auth.AuthRealm, "protocol", "openid-connect", "token"),
		ClientID:     string(cid.Data[mcc.Spec.Auth.ClientID.SecretKeyRef.Key]),
		ClientSecret: string(cs.Data[mcc.Spec.Auth.ClientSecret.SecretKeyRef.Key]),
	})

	if err != nil {
		return err
	}

	namespaces, nsErr := c.GetNamespaces(ctx, mcc.Name, 0)
	if nsErr != nil {
		return errors.Wrapf(nsErr, "failure polling for namespaces")
	}

	for i := range namespaces {
		r.l.Info("namespace", "id", namespaces[i].Id, "revision", namespaces[i].ResourceVersion)
	}

	connectors, cnErr := c.GetConnectors(ctx, mcc.Name, 0)
	if cnErr != nil {
		return errors.Wrapf(nsErr, "failure polling for connectors")
	}

	for i := range connectors {
		r.l.Info("connector", "id", connectors[i].Id, "revision", connectors[i].Metadata.ResourceVersion)
	}

	return nil
}
