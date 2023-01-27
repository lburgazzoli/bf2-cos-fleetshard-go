package cos

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	cosv2 "gitub.com/lburgazzoli/bf2-cos-fleetshard-go/apis/cos/v2"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/internal/api/controlplane"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/controller"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/cos/fleetmanager"
	cosmeta "gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/cos/fleetshard/meta"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/defaults"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/resources"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
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

	err := r.poll(ctx, req.NamespacedName, mcc)
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

func (r *ManagedConnectorClusterReconciler) poll(ctx context.Context, named types.NamespacedName, mcc *cosv2.ManagedConnectorCluster) error {
	c, err := r.cluster(ctx, named, mcc)
	if err != nil {
		return err
	}

	namespaces, nsErr := c.GetNamespaces(ctx, 0)
	if nsErr != nil {
		return errors.Wrapf(nsErr, "failure polling for namespaces")
	}

	for i := range namespaces {
		r.l.Info("namespace", "id", namespaces[i].Id, "revision", namespaces[i].ResourceVersion)

		ns := corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "mctr-" + namespaces[i].Id,
				Labels: map[string]string{
					cosmeta.MetaClusterID:         c.Parameters.ClusterID,
					cosmeta.MetaNamespaceID:       namespaces[i].Id,
					cosmeta.MetaNamespaceRevision: fmt.Sprintf("%d", namespaces[i].ResourceVersion),
				},
			},
		}

		newNs := ns.DeepCopy()

		patched, err := resources.Apply(ctx, r.Client, &ns, newNs)
		if err != nil {
			return err
		}
		r.l.Info("namespace", "id", namespaces[i].Id, "revision", namespaces[i].ResourceVersion, "patched", patched)
	}

	connectors, cnErr := c.GetConnectors(ctx, 0)
	if cnErr != nil {
		return errors.Wrapf(nsErr, "failure polling for connectors")
	}

	for i := range connectors {
		r.l.Info("connector", "id", connectors[i].Id, "revision", connectors[i].Metadata.ResourceVersion)

		c := cosv2.ManagedConnector{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "mctr-" + *connectors[i].Spec.NamespaceId,
				Name:      "mctr-" + *connectors[i].Id,
				Labels: map[string]string{
					cosmeta.MetaClusterID:          c.Parameters.ClusterID,
					cosmeta.MetaNamespaceID:        namespaces[i].Id,
					cosmeta.MetaDeploymentID:       *connectors[i].Id,
					cosmeta.MetaDeploymentRevision: fmt.Sprintf("%d", connectors[i].Metadata.ResourceVersion),
					cosmeta.MetaConnectorID:        *connectors[i].Spec.ConnectorId,
					cosmeta.MetaConnectorRevision:  fmt.Sprintf("%d", connectors[i].Spec.ConnectorResourceVersion),
				},
			},
		}

		newC := c.DeepCopy()

		patched, err := resources.Apply(ctx, r.Client, &c, newC)
		if err != nil {
			return err
		}

		r.l.Info("connector", "id", connectors[i].Id, "revision", connectors[i].Metadata.ResourceVersion, "patched", patched)
	}

	return nil
}

func (r *ManagedConnectorClusterReconciler) update(ctx context.Context, named types.NamespacedName, mcc *cosv2.ManagedConnectorCluster) error {
	c, err := r.cluster(ctx, named, mcc)
	if err != nil {
		return err
	}

	return c.Client.UpdateClusterStatus(ctx, controlplane.ConnectorClusterStatus{})

}

func (r *ManagedConnectorClusterReconciler) cluster(ctx context.Context, named types.NamespacedName, mcc *cosv2.ManagedConnectorCluster) (Cluster, error) {
	if c, ok := r.clients[named]; ok {
		return c, nil
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mcc.Spec.Secret,
			Namespace: mcc.Namespace,
		},
	}

	if err := resources.Get(ctx, r, secret); err != nil {
		return Cluster{}, err
	}

	params, err := DecodeAddonsParams(secret.Data)
	if err != nil {
		return Cluster{}, err
	}

	cpUrl, err := url.Parse(params.BaseURL)
	if err != nil {
		return Cluster{}, err
	}

	autUrl, err := url.Parse(params.AuthURL)
	if err != nil {
		return Cluster{}, err
	}

	c, err := fleetmanager.NewClient(ctx, fleetmanager.Config{
		ApiURL:       cpUrl,
		AuthURL:      autUrl,
		AuthTokenURL: autUrl.JoinPath("auth", "realms", params.AuthRealm, "protocol", "openid-connect", "token"),
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
