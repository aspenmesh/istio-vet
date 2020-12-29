module github.com/aspenmesh/istio-vet

go 1.13

require (
	github.com/cnf/structhash v0.0.0-20180104161610-62a607eb0224
	github.com/ghodss/yaml v1.0.0
	github.com/gogo/protobuf v1.3.1
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/protobuf v1.4.2
	github.com/hashicorp/golang-lru v0.5.3 // indirect
	github.com/hashicorp/hcl v0.0.0-20171017181929-23c074d0eceb // indirect
	github.com/imdario/mergo v0.3.7 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/magiconair/properties v1.7.4 // indirect
	github.com/mitchellh/mapstructure v0.0.0-20180111000720-b4575eea38cc // indirect
	github.com/onsi/ginkgo v1.11.0
	github.com/onsi/gomega v1.7.0
	github.com/pelletier/go-toml v1.0.1 // indirect
	github.com/spf13/cast v1.1.0 // indirect
	github.com/spf13/cobra v0.0.3
	github.com/spf13/jwalterweatherman v0.0.0-20180109140146-7c0cea34c8ec // indirect
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.0.0
	github.com/wadey/gocovmerge v0.0.0-20160331181800-b5bfa59ec0ad // indirect
	istio.io/api v0.0.0-20201112235759-fa4ee46c5dc2
	istio.io/client-go v0.0.0-20200908160912-f99162621a1a
	k8s.io/api v0.19.3
	k8s.io/apimachinery v0.19.3
	k8s.io/client-go v0.19.3
)

replace (
	github.com/golang/protobuf => github.com/golang/protobuf v1.4.2
	k8s.io/client-go => k8s.io/client-go v0.19.3
)
