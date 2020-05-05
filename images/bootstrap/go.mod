module github.com/mesosphere/kudo-cassandra-operator/images/bootstrap

go 1.13

replace k8s.io/api => k8s.io/api v0.0.0-20191016110408-35e52d86657a

replace k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20191004115801-a2eda9f80ab8

replace k8s.io/client-go => k8s.io/client-go v0.0.0-20191016111102-bec269661e48

require (
	github.com/fatih/color v1.9.0 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.3 // indirect
	github.com/mattn/go-colorable v0.1.6 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/sirupsen/logrus v1.6.0
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/objx v0.1.1 // indirect
	github.com/stretchr/testify v1.3.0
	github.com/yanniszark/go-nodetool v0.0.0-20191206125106-cd8f91fa16be // indirect
	golang.org/x/crypto v0.0.0-20200429183012-4b2356b1ed79 // indirect
	golang.org/x/sync v0.0.0-20200317015054-43a5402ce75a // indirect
	golang.org/x/sys v0.0.0-20200501145240-bc7a7d42d5c3 // indirect
	gotest.tools/gotestsum v0.4.2 // indirect
	k8s.io/api v0.0.0-20191016110408-35e52d86657a
	k8s.io/apimachinery v0.0.0-20191004115801-a2eda9f80ab8
	k8s.io/client-go v0.0.0-00010101000000-000000000000
)
