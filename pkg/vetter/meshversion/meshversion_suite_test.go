package meshversion

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestMeshversion(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Meshversion Suite")
}
