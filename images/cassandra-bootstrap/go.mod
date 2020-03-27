module github.com/mesosphere/kudo-cassandra-operator/images/cassandra-bootstrap

go 1.13

replace k8s.io/api => k8s.io/api v0.0.0-20191016110408-35e52d86657a

replace k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20191004115801-a2eda9f80ab8

replace k8s.io/client-go => k8s.io/client-go v0.0.0-20191016111102-bec269661e48

require (
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.3.0
	github.com/yanniszark/go-nodetool v0.0.0-20191206125106-cd8f91fa16be // indirect
	k8s.io/apimachinery v0.0.0-20191004115801-a2eda9f80ab8
	k8s.io/client-go v0.0.0-00010101000000-000000000000
)
