module github.com/mesosphere/kudo-cassandra-operator/tests

go 1.13

require (
	github.com/avast/retry-go v2.4.1+incompatible
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/imdario/mergo v0.3.8 // indirect
	github.com/kudobuilder/kudo v0.7.4
	github.com/mitchellh/go-homedir v1.1.0
	github.com/onsi/ginkgo v1.8.0
	github.com/onsi/gomega v1.5.0
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/time v0.0.0-20190921001708-c4c64cad1fd0 // indirect
	k8s.io/api v0.0.0-20190409021203-6e4e0e4f393b
	k8s.io/apiextensions-apiserver v0.0.0-20190409022649-727a075fdec8
	k8s.io/apimachinery v0.0.0-20190704094520-6f131bee5e2c
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
)
