package invalidserviceforjwtpolicy_test

import (
"testing"

. "github.com/onsi/ginkgo"
. "github.com/onsi/gomega"
)

func TestInvalidServiceForJWTPolicy(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Invalid Service For JWT Policy Suite")
}

