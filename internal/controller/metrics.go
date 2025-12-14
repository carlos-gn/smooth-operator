package controller

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	mcpServersTotal = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "mcpserver_total",
		Help: "Total number of MCPserver resources",
	}, []string{"namespace"})

	mcpServersByPhase = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{Name: "mcpserver_phase", Help: "Number of MCPservers by phase"},
		[]string{"namespace", "phase"},
	)

	deploymentCreationErrors = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "mcpserver_deployment_creation_errors_total",
		Help: "Total deployment creation errors",
	}, []string{"namespace", "name"})

	serviceCreationErrors = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "mcpserver_service_creation_errors_total",
		Help: "Total service creation errors",
	}, []string{"namespace", "name"})

	reconcileTime = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "mcpserver_reconcile_duration_seconds",
		Help:    "Time taken to reconcile MCPServer",
		Buckets: prometheus.DefBuckets,
	}, []string{"namespace"})
)

func init() {
	metrics.Registry.MustRegister(mcpServersTotal, mcpServersByPhase, deploymentCreationErrors, serviceCreationErrors, reconcileTime)
}
