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
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/resources/secrets"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// ManagedConnectorClusterReconciler reconciles a ManagedConnector object
type ManagedConnectorClusterReconciler struct {
	client.Client
	Scheme  *runtime.Scheme
	mgr     manager.Manager
	options controller.Options
	clients map[types.NamespacedName]Cluster
	l       logr.Logger
}

func NewManagedConnectorClusterReconciler(mgr manager.Manager, options controller.Options) (*ManagedConnectorClusterReconciler, error) {
	r := &ManagedConnectorClusterReconciler{
		Client:  mgr.GetClient(),
		Scheme:  mgr.GetScheme(),
		mgr:     mgr,
		options: options,
		clients: make(map[types.NamespacedName]Cluster),
		l:       log.Log.WithName("cluster-reconciler"),
	}

	return r, r.initialize(mgr)
}

func (r *ManagedConnectorClusterReconciler) initialize(mgr ctrl.Manager) error {
	c := ctrl.NewControllerManagedBy(mgr).
		Named("ManagedConnectorClusterController").
		For(&cosv2.ManagedConnectorCluster{}, builder.WithPredicates(
			predicate.Or(
				predicate.GenerationChangedPredicate{},
				predicate.AnnotationChangedPredicate{},
				predicate.LabelChangedPredicate{},
			)))

	return c.Complete(r)
}

// +kubebuilder:rbac:groups=cos.bf2.dev,resources=managedconnectorclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cos.bf2.dev,resources=managedconnectorclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cos.bf2.dev,resources=managedconnectorclusters/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete

func (r *ManagedConnectorClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)
	l.Info("Reconciling")

	mccRef := &cosv2.ManagedConnectorCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.NamespacedName.Name,
			Namespace: req.NamespacedName.Namespace,
		},
	}

	if err := resources.Get(ctx, r, mccRef); err != nil && !k8serrors.IsNotFound(err) {
		r.l.Error(err, "failure getting resource", "ManagedConnectorCluster", req.NamespacedName.String())
	}

	// safety copy
	mcc := mccRef.DeepCopy()

	if mcc.ObjectMeta.DeletionTimestamp.IsZero() {

		//
		// Add finalizer
		//

		if controllerutil.AddFinalizer(mcc, defaults.ConnectorClustersFinalizerName) {
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

		if controllerutil.RemoveFinalizer(mcc, defaults.ConnectorClustersFinalizerName) {
			if err := r.Update(ctx, mcc); err != nil {
				if k8serrors.IsConflict(err) {
					return ctrl.Result{}, err
				}

				return ctrl.Result{}, errors.Wrapf(err, "failure removing finalizer from connector cluster %s", req.NamespacedName)
			}
		}

		// TODO: delete all resources associated to the cluster

		delete(r.clients, req.NamespacedName)

		return ctrl.Result{}, nil
	}

	mcc.Status.ObservedGeneration = mcc.Generation
	mcc.Status.Phase = "Running"

	err := r.pollAndApply(ctx, req.NamespacedName, mcc)
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

	if _, err := resources.PatchStatus(ctx, r.Client, mccRef, mcc); err != nil {
		if k8serrors.IsConflict(err) {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, err
	}

	// Poor man scheduler
	return ctrl.Result{RequeueAfter: mcc.Spec.PollDelay.Duration}, nil
}

func (r *ManagedConnectorClusterReconciler) pollAndApply(ctx context.Context, named types.NamespacedName, mcc *cosv2.ManagedConnectorCluster) error {
	c, err := r.cluster(ctx, named, mcc)
	if err != nil {
		return err
	}

	if err := r.deployNamespaces(ctx, c, 0); err != nil {
		return errors.Wrapf(err, "failure handling for namespaces")
	}

	if err := r.deployConnectors(ctx, c, 0); err != nil {
		return errors.Wrapf(err, "failure handling for namespaces")
	}

	return nil
}

func (r *ManagedConnectorClusterReconciler) cluster(ctx context.Context, named types.NamespacedName, mcc *cosv2.ManagedConnectorCluster) (Cluster, error) {
	if c, ok := r.clients[named]; ok {
		return c, nil
	}

	secret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mcc.Spec.Secret,
			Namespace: mcc.Namespace,
		},
	}

	if err := resources.Get(ctx, r, &secret); err != nil {
		return Cluster{}, err
	}

	params, err := secrets.Decode[AddonParameters](secret)
	if err != nil {
		return Cluster{}, err
	}

	c, err := fleetmanager.NewClient(ctx, fleetmanager.Config{
		ApiURL:       params.BaseURL,
		AuthURL:      params.AuthURL,
		AuthTokenURL: params.AuthURL.JoinPath("auth", "realms", params.AuthRealm, "protocol", "openid-connect", "token"),
		ClientID:     params.ClientID,
		ClientSecret: params.ClientSecret,
		ClusterID:    params.ClusterID,
	})

	if err != nil {
		return Cluster{}, err
	}

	answer := Cluster{
		Client:     c,
		Parameters: params,
	}

	r.clients[named] = answer

	return answer, nil

}
