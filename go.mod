module github.com/aspenmesh/istio-vet

go 1.12

require (
	cloud.google.com/go v0.41.0 // indirect
	github.com/aspenmesh/istio-client-go v0.0.0-20191010215625-4de6e89009c4
	github.com/cnf/structhash v0.0.0-20180104161610-62a607eb0224
	github.com/ghodss/yaml v1.0.0
	github.com/gogo/protobuf v1.3.0
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/protobuf v1.3.2
	github.com/hashicorp/hcl v0.0.0-20171017181929-23c074d0eceb // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/json-iterator/go v1.1.7 // indirect
	github.com/kr/pretty v0.1.0 // indirect
	github.com/magiconair/properties v1.7.4 // indirect
	github.com/mitchellh/mapstructure v0.0.0-20180111000720-b4575eea38cc // indirect
	github.com/onsi/ginkgo v1.10.2
	github.com/onsi/gomega v1.5.0
	github.com/pelletier/go-toml v1.0.1 // indirect
	github.com/spf13/afero v1.2.2 // indirect
	github.com/spf13/cast v1.1.0 // indirect
	github.com/spf13/cobra v0.0.3
	github.com/spf13/jwalterweatherman v0.0.0-20180109140146-7c0cea34c8ec // indirect
	github.com/spf13/pflag v1.0.3
	github.com/spf13/viper v1.0.0
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
	istio.io/api v0.0.0-20190820204432-483f2547d882
	k8s.io/api v0.0.0-20190620084959-7cf5895f2711
	k8s.io/apimachinery v0.0.0-20190612205821-1799e75a0719
	k8s.io/client-go v0.0.0-20190620085101-78d2af792bab
	k8s.io/klog v1.0.0 // indirect
)

replace (
	github.com/golang/protobuf => github.com/golang/protobuf v1.3.2
	k8s.io/api => k8s.io/api v0.0.0-20190620084959-7cf5895f2711
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190612205821-1799e75a0719
	k8s.io/client-go => k8s.io/client-go v0.0.0-20190620085101-78d2af792bab
)
