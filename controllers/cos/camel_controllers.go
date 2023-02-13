package cos

import (
	"context"
	camelclient "github.com/apache/camel-k/pkg/client"
	itctrl "github.com/apache/camel-k/pkg/controller/integration"
	ikctl "github.com/apache/camel-k/pkg/controller/integrationkit"
	itpctrl "github.com/apache/camel-k/pkg/controller/integrationplatform"
	klbctrl "github.com/apache/camel-k/pkg/controller/kameletbinding"
	"github.com/pkg/errors"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/controller"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func SetupCamelControllers(mgr manager.Manager, options controller.Options) error {
	c, err := camelclient.FromManager(mgr)
	if err != nil {
		return errors.Wrap(err, "unable to create client from manager")
	}

	err = mgr.GetFieldIndexer().IndexField(context.Background(), &corev1.Pod{}, "status.phase",
		func(obj client.Object) []string {
			pod, _ := obj.(*corev1.Pod)
			return []string{string(pod.Status.Phase)}
		})

	if err != nil {
		return errors.Wrap(err, "unable to set up field indexer")
	}

	if err := itpctrl.Add(mgr, c); err != nil {
		return errors.Wrap(err, "unable to set up integration platform controller")
	}
	if err := ikctl.Add(mgr, c); err != nil {
		return errors.Wrap(err, "unable to set up integration kit controller")
	}
	if err := klbctrl.Add(mgr, c); err != nil {
		return errors.Wrap(err, "unable to set up kamelet binding controller")
	}
	if err := itctrl.Add(mgr, c); err != nil {
		return errors.Wrap(err, "unable to set up integration controller")
	}

	return nil
}
