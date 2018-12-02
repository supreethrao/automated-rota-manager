package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var memberCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "support_days_counter",
		Help: "Metrics which keep track of each person's support days",
	},
	[]string{"name", "date"},
)

func init() {
	prometheus.MustRegister(memberCounter)
}