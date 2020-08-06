module github.com/aspenmesh/istio-vet

go 1.13

require (
	cloud.google.com/go v0.50.0 // indirect
	github.com/cnf/structhash v0.0.0-20180104161610-62a607eb0224
	github.com/ghodss/yaml v1.0.0
	github.com/gogo/protobuf v1.3.1
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/groupcache v0.0.0-20191027212112-611e8accdfc9 // indirect
	github.com/golang/protobuf v1.3.5
	github.com/imdario/mergo v0.3.7 // indirect
	github.com/onsi/ginkgo v1.11.0
	github.com/onsi/gomega v1.9.0
	github.com/pelletier/go-toml v1.3.0 // indirect
	github.com/spf13/cobra v1.0.0
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.4.0
	istio.io/api v0.0.0-20200708135631-b736e804afd1
	istio.io/client-go v0.0.0-20200316192452-065c59267750
	k8s.io/api v0.18.1
	k8s.io/apimachinery v0.18.1
	k8s.io/client-go v0.18.0
)

replace (
	istio.io/api => istio.io/api v0.0.0-20200316215140-da46fe8e25be
	k8s.io/api => k8s.io/api v0.17.3
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.3
	k8s.io/client-go => k8s.io/client-go v0.17.3
)
