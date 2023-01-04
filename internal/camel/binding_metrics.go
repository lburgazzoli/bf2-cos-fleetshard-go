package camel

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	patchDependantCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "patch_dependant",
			Help: "patch_dependant",
		},
		[]string{
			"connector_id",
			"dependant_name",
			"dependant_kind",
		},
	)
)

func init() {
	metrics.Registry.MustRegister(patchDependantCount)
}
