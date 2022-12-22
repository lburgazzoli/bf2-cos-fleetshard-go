package controller

import ctrl "sigs.k8s.io/controller-runtime"

type Controller interface {
	Init(ctrl.Manager) error
	Reify(*ReconciliationContext) error
	Delete(*ReconciliationContext) error
	Stop(*ReconciliationContext) error
}
