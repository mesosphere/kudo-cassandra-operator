module github.com/mesosphere/kudo-cassandra-operator/tests

go 1.13

require (
	github.com/imdario/mergo v0.3.8 // indirect
	github.com/kudobuilder/kudo v0.11.0
	github.com/kudobuilder/test-tools v0.2.2
	github.com/onsi/ginkgo v1.10.1
	github.com/onsi/gomega v1.7.1
	github.com/sirupsen/logrus v1.4.2
	github.com/thoas/go-funk v0.5.0
	golang.org/x/time v0.0.0-20190921001708-c4c64cad1fd0 // indirect
	gopkg.in/yaml.v2 v2.2.7
	k8s.io/api v0.17.2
	k8s.io/apimachinery v0.17.2
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible // indirect
)

replace k8s.io/api => k8s.io/api v0.0.0-20191016110408-35e52d86657a

replace k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20191004115801-a2eda9f80ab8

replace k8s.io/client-go => k8s.io/client-go v0.0.0-20191016111102-bec269661e48

replace github.com/kudobuilder/test-tools => /Users/aneumann/git/test-tools
