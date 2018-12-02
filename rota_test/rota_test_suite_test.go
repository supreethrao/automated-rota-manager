package rota_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestRotaTest(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test Suite for support rota")
}
