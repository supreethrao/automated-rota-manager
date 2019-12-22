package metrics_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestMetricsLogic(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test Suite for support bot metrics")
}
