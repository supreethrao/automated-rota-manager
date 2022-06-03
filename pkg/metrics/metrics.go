package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var memberCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "rota_counter",
		Help: "Metrics which keep track of days each person was picked",
	},
	[]string{"name", "date"},
)

func init() {
	prometheus.MustRegister(memberCounter)
}
