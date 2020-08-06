module github.com/aspenmesh/istio-vet

go 1.13

require (
	cloud.google.com/go v0.41.0 // indirect
	github.com/cnf/structhash v0.0.0-20180104161610-62a607eb0224
	github.com/ghodss/yaml v1.0.0
	github.com/gogo/protobuf v1.3.1
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/groupcache v0.0.0-20191002201903-404acd9df4cc // indirect
	github.com/golang/protobuf v1.3.5
	github.com/googleapis/gnostic v0.3.1 // indirect
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
	golang.org/x/time v0.0.0-20190921001708-c4c64cad1fd0 // indirect
	google.golang.org/appengine v1.6.5 // indirect
	google.golang.org/genproto v0.0.0-20190916214212-f660b8655731 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	istio.io/api v0.0.0-20200724154434-34e474846e0d
	istio.io/client-go v0.0.0-20200708142230-d7730fd90478
	k8s.io/api v0.18.1
	k8s.io/apimachinery v0.18.1
	k8s.io/client-go v0.18.1
)

replace (
	github.com/golang/protobuf => github.com/golang/protobuf v1.3.2
	k8s.io/client-go => k8s.io/client-go v0.18.0
)
