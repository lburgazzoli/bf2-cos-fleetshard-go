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
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"time"
)

// ManagedConnectorClusterReconciler reconciles a ManagedConnector object
type ManagedConnectorClusterReconciler struct {
	client.Client
	mgr     manager.Manager
	options controller.Options
	clients map[types.NamespacedName]*Cluster
	l       logr.Logger
}

func SetupManagedConnectorClusterReconciler(mgr manager.Manager, options controller.Options) error {
	r := &ManagedConnectorClusterReconciler{
		Client:  mgr.GetClient(),
		mgr:     mgr,
		options: options,
		clients: make(map[types.NamespacedName]*Cluster),
		l:       log.Log.WithName("cluster-reconciler"),
	}

	return r.initialize(mgr)
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
				if k8serrors.IsNotFound(err) {
					return ctrl.Result{}, nil
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
				if k8serrors.IsNotFound(err) {
					return ctrl.Result{}, nil
				}

				return ctrl.Result{}, errors.Wrapf(err, "failure removing finalizer from connector cluster %s", req.NamespacedName)
			}
		}

		// TODO: delete all resources associated to the cluster

		delete(r.clients, req.NamespacedName)

		return ctrl.Result{}, nil
	}

	mcc.Status.Phase = "Running"

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
		mcc.Status.ClusterID = c.Parameters.ClusterID

		if c.Parameters.BaseURL != nil {
			mcc.Status.ControlPlaneURL = c.Parameters.BaseURL.String()
		}

		if err := r.updateClusterStatus(ctx, mcc); err != nil {
			meta.SetStatusCondition(&mcc.Status.Conditions, metav1.Condition{
				Type:    "SyncClusterStatus",
				Status:  metav1.ConditionFalse,
				Reason:  "Error",
				Message: err.Error(),
			})
		} else {
			meta.SetStatusCondition(&mcc.Status.Conditions, metav1.Condition{
				Type:    "SyncClusterStatus",
				Status:  metav1.ConditionTrue,
				Reason:  "Synced",
				Message: "Synced",
			})
		}

		if err := r.updateConnectorsStatus(ctx, mcc); err != nil {
			meta.SetStatusCondition(&mcc.Status.Conditions, metav1.Condition{
				Type:    "SyncConnectorsStatus",
				Status:  metav1.ConditionFalse,
				Reason:  "Error",
				Message: err.Error(),
			})
		} else {
			meta.SetStatusCondition(&mcc.Status.Conditions, metav1.Condition{
				Type:    "SyncConnectorsStatus",
				Status:  metav1.ConditionTrue,
				Reason:  "Synced",
				Message: "Synced",
			})
		}

		if err := r.pollAndApply(ctx, c); err != nil {
			meta.SetStatusCondition(&mcc.Status.Conditions, metav1.Condition{
				Type:    "Poll",
				Status:  metav1.ConditionFalse,
				Reason:  "Error",
				Message: err.Error(),
			})
		} else {
			meta.SetStatusCondition(&mcc.Status.Conditions, metav1.Condition{
				Type:    "Poll",
				Status:  metav1.ConditionTrue,
				Reason:  "Executed",
				Message: "Executed",
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

func (r *ManagedConnectorClusterReconciler) pollAndApply(ctx context.Context, c *Cluster) error {

	now := time.Now()
	resync := now.After(c.ResyncAt)

	r.l.Info("poll", "resyncIn", c.ResyncAt.Sub(now).String(), "resync", resync)

	//
	// Namespaces
	//

	var gvNamespace int64

	if !resync {
		namespaces, err := r.namespaces(ctx, c)
		if err != nil {
			return err
		}

		rev, err := computeMaxRevision(namespaces, cosmeta.MetaNamespaceRevision)
		if err != nil {
			return err
		}

		gvNamespace = rev
	}

	if err := r.deployNamespaces(ctx, c, gvNamespace); err != nil {
		return errors.Wrapf(err, "failure handling namespaces")
	}

	//
	// Connectors
	//

	var gvConnector int64

	if !resync {
		connectors, err := r.connectors(ctx, c)
		if err != nil {
			return err
		}

		rev, err := computeMaxRevision(connectors, cosmeta.MetaDeploymentRevision)
		if err != nil {
			return err
		}

		gvConnector = rev
	}

	if err := r.deployConnectors(ctx, c, gvConnector); err != nil {
		return errors.Wrapf(err, "failure handling connectors deployment")
	}

	if resync {
		c.ResyncAt = time.Now().Add(c.ResyncDelay)
	}

	return nil
}

func (r *ManagedConnectorClusterReconciler) clusterById(id string) *Cluster {
	for _, v := range r.clients {
		cluster := v

		if cluster.Parameters.ClusterID == id {
			return cluster
		}
	}

	return nil
}

func (r *ManagedConnectorClusterReconciler) cluster(ctx context.Context, mcc *cosv2.ManagedConnectorCluster) (*Cluster, error) {
	named := resources.AsNamespacedName(mcc)

	if v, ok := r.clients[named]; ok {
		cluster := v
		return cluster, nil
	}

	secret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mcc.Spec.Secret,
			Namespace: mcc.Namespace,
		},
	}

	if err := resources.Get(ctx, r, &secret); err != nil {
		return nil, err
	}

	params, err := secrets.Decode[AddonParameters](secret)
	if err != nil {
		return nil, err
	}

	c, err := fleetmanager.NewClient(ctx, fleetmanager.Config{
		ApiURL:       params.BaseURL,
		AuthURL:      params.AuthURL,
		AuthTokenURL: params.AuthURL.JoinPath("auth", "realms", params.AuthRealm, "protocol", "openid-connect", "token"),
		ClientID:     params.ClientID,
		ClientSecret: params.ClientSecret,
		ClusterID:    params.ClusterID,
		Debug:        false,
	})

	if err != nil {
		return nil, err
	}

	answer := &Cluster{
		MCC:         *mcc,
		Client:      c,
		Parameters:  params,
		ResyncDelay: mcc.Spec.ResyncDelay.Duration,
		ResyncAt:    time.Time{},
	}

	r.clients[named] = answer

	return answer, nil

}

func (r *ManagedConnectorClusterReconciler) namespaces(ctx context.Context, c *Cluster) ([]corev1.Namespace, error) {
	options := client.MatchingLabels{
		cosmeta.MetaClusterID: c.Parameters.ClusterID,
	}

	namespaces := corev1.NamespaceList{}
	if err := r.List(ctx, &namespaces, options); err != nil {
		return nil, err
	}

	return namespaces.Items, nil
}

func (r *ManagedConnectorClusterReconciler) connectors(ctx context.Context, c *Cluster) ([]cosv2.ManagedConnector, error) {
	options := client.MatchingLabels{
		cosmeta.MetaClusterID: c.Parameters.ClusterID,
	}

	connectors := cosv2.ManagedConnectorList{}
	if err := r.List(ctx, &connectors, options); err != nil {
		return nil, err
	}

	return connectors.Items, nil
}
