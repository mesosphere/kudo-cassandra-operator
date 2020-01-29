package suites

import (
	"fmt"

	"github.com/mesosphere/kudo-cassandra-operator/tests/utils/k8s"
	"github.com/mesosphere/kudo-cassandra-operator/tests/utils/kudo"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("external service", func() {
	NodeCount = 3

	It("Installs the operator from the current directory", func() {
		err := kudo.InstallOperator(
			OperatorDirectory, TestNamespace, TestInstance, []string{}, true,
		)
		if err != nil {
			Fail(
				"Failing the full suite: failed to install operator instance that the " +
					"following tests depend on",
			)
		}
		Expect(err).To(BeNil())

		assertNumberOfCassandraNodes(NodeCount)
	})

	It("Allows external access to the cassandra cluster", func() {
		err := kudo.UpdateInstanceParameters(
			TestNamespace,
			TestInstance,
			map[string]string{"EXTERNAL_NATIVE_TRANSPORT": "true"},
			false,
		)
		Expect(err).To(BeNil())

		assertNumberOfCassandraNodes(NodeCount)

		svc, err := k8s.GetService(TestNamespace, fmt.Sprintf("%s-svc-external", TestInstance))
		Expect(err).To(BeNil())
		Expect(len(svc.Spec.Ports)).To(Equal(1))
	})

	It("Opens a second port if rpc is enabled", func() {
		err := kudo.UpdateInstanceParameters(
			TestNamespace,
			TestInstance,
			map[string]string{
				"START_RPC":    "true",
				"EXTERNAL_RPC": "true",
			},
			true,
		)
		Expect(err).To(BeNil())

		assertNumberOfCassandraNodes(NodeCount)

		svc, err := k8s.GetService(TestNamespace, fmt.Sprintf("%s-svc-external", TestInstance))
		Expect(err).To(BeNil())
		Expect(len(svc.Spec.Ports)).To(Equal(2))
	})

	It("Uninstalls the operator", func() {
		err := kudo.UninstallOperator(OperatorName, TestNamespace, TestInstance)
		Expect(err).To(BeNil())
	})
})
