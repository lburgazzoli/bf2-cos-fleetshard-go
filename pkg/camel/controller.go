package camel

import (
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/controller"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Controller struct {
}

func (c *Controller) Init(ctrl.Manager) error {
	return nil
}

func (c *Controller) Reify(ctx *controller.ReconciliationContext) error {
	return nil
}

func (c *Controller) Delete(ctx *controller.ReconciliationContext) error {
	return nil
}

func (c *Controller) Stop(ctx *controller.ReconciliationContext) error {
	return nil
}
