package mtlsprobes_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestMtlsprobes(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Mtlsprobes Suite")
}
