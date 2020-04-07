module github.com/mesosphere/kudo-cassandra-operator/tests

go 1.13

require (
	github.com/imdario/mergo v0.3.8 // indirect
	github.com/kudobuilder/test-tools v0.2.7-0.20200407112611-e97f7e5dcc75
	github.com/onsi/ginkgo v1.11.0
	github.com/onsi/gomega v1.8.1
	github.com/sirupsen/logrus v1.4.2
	github.com/thoas/go-funk v0.5.0
	golang.org/x/time v0.0.0-20190921001708-c4c64cad1fd0 // indirect
	gopkg.in/yaml.v2 v2.2.8
	k8s.io/api v0.17.4
	k8s.io/apimachinery v0.17.4
)

replace k8s.io/api => k8s.io/api v0.17.4

replace k8s.io/apimachinery => k8s.io/apimachinery v0.17.4

replace k8s.io/client-go => k8s.io/client-go v0.17.4
