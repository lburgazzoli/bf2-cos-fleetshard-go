package cos

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	cosv2 "gitub.com/lburgazzoli/bf2-cos-fleetshard-go/apis/cos/v2"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/controller"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/cos/fleetmanager"
	cosmeta "gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/cos/fleetshard/meta"
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
	"strconv"
	"time"
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

	c, err := r.cluster(ctx, mcc)
	if err != nil {
		meta.SetStatusCondition(&mcc.Status.Conditions, metav1.Condition{
			Type:    "Triggered",
			Status:  metav1.ConditionFalse,
			Reason:  "Error",
			Message: err.Error(),
		})
	} else {
		mcc.Status.ObservedGeneration = mcc.Generation
		mcc.Status.Phase = "Running"
		mcc.Status.ClusterID = c.Parameters.ClusterID

		if c.Parameters.BaseURL != nil {
			mcc.Status.ControlPlaneURL = c.Parameters.BaseURL.String()
		}

		if err := r.updateClusterStatus(ctx, mcc); err != nil {
			return ctrl.Result{}, err
		}
		if err := r.updateConnectorsStatus(ctx, mcc); err != nil {
			return ctrl.Result{}, err
		}

		if err := r.pollAndApply(ctx, c); err != nil {
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

func (r *ManagedConnectorClusterReconciler) pollAndApply(ctx context.Context, c Cluster) error {

	resync := time.Now().After(c.ResyncAt)
	options := client.MatchingLabels{cosmeta.MetaClusterID: c.Parameters.ClusterID}

	//
	// Namespaces
	//

	var gvNamespace int64

	if !resync {
		namespaces := corev1.NamespaceList{}
		if err := r.List(ctx, &namespaces, options); err != nil {
			return err
		}

		for i := range namespaces.Items {
			namespace := namespaces.Items[i]

			if len(namespace.Annotations) == 0 {
				continue
			}

			if r, ok := namespace.Annotations[cosmeta.MetaNamespaceRevision]; ok {
				rev, err := strconv.ParseInt(r, 10, 64)
				if err != nil {
					return err
				}

				if rev > gvNamespace {
					gvNamespace = rev
				}
			}
		}
	}

	if err := r.deployNamespaces(ctx, c, gvNamespace); err != nil {
		return errors.Wrapf(err, "failure handling for namespaces")
	}

	//
	// Connectors
	//

	connectors := cosv2.ManagedConnectorList{}
	if err := r.List(ctx, &connectors, options); err != nil {
		return err
	}

	var gvConnector int64

	if !resync {
		for i := range connectors.Items {
			connector := connectors.Items[i]

			if len(connector.Annotations) == 0 {
				continue
			}

			if r, ok := connector.Annotations[cosmeta.MetaConnectorRevision]; ok {
				rev, err := strconv.ParseInt(r, 10, 64)
				if err != nil {
					return err
				}

				if rev > gvConnector {
					gvConnector = rev
				}
			}
		}
	}

	if err := r.deployConnectors(ctx, c, gvConnector); err != nil {
		return errors.Wrapf(err, "failure handling for namespaces")
	}

	c.ResyncAt = time.Now().Add(c.ResyncDelay)

	return nil
}

func (r *ManagedConnectorClusterReconciler) clusterById(id string) *Cluster {
	for _, v := range r.clients {
		cluster := v

		if cluster.Parameters.ClusterID == id {
			return &cluster
		}
	}

	return nil
}

func (r *ManagedConnectorClusterReconciler) cluster(ctx context.Context, mcc *cosv2.ManagedConnectorCluster) (Cluster, error) {
	named := resources.AsNamespacedName(mcc)

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
		Client:      c,
		Parameters:  params,
		ResyncDelay: mcc.Spec.ResyncDelay.Duration,
		ResyncAt:    time.Now().Add(mcc.Spec.ResyncDelay.Duration),
	}

	r.clients[named] = answer

	return answer, nil

}
