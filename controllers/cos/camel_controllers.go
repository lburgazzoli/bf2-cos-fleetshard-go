package cos

import (
	camelclient "github.com/apache/camel-k/pkg/client"
	itctrl "github.com/apache/camel-k/pkg/controller/integration"
	itpctrl "github.com/apache/camel-k/pkg/controller/integrationplatform"
	klbctrl "github.com/apache/camel-k/pkg/controller/kameletbinding"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func SetupCamelControllers(mgr manager.Manager, options controller.Options) error {
	c, err := camelclient.FromManager(mgr)
	if err != nil {
		return err
	}

	if err := itpctrl.Add(mgr, c); err != nil {
		return err
	}
	if err := klbctrl.Add(mgr, c); err != nil {
		return err
	}
	if err := itctrl.Add(mgr, c); err != nil {
		return err
	}

	return nil
}
