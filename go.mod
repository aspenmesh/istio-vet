module github.com/aspenmesh/istio-vet

go 1.12

require (
	cloud.google.com/go v0.41.0
	github.com/aspenmesh/istio-client-go v0.0.0-20200122202704-9695ccefca79
	github.com/cnf/structhash v0.0.0-20180104161610-62a607eb0224
	github.com/davecgh/go-spew v1.1.1
	github.com/fsnotify/fsnotify v1.4.7
	github.com/ghodss/yaml v1.0.0
	github.com/gogo/protobuf v1.3.0
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/protobuf v1.3.2
	github.com/google/btree v1.0.0
	github.com/google/gofuzz v1.0.0
	github.com/googleapis/gnostic v0.3.1
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79
	github.com/hashicorp/golang-lru v0.5.3
	github.com/hashicorp/hcl v0.0.0-20171017181929-23c074d0eceb
	github.com/hpcloud/tail v1.0.0
	github.com/imdario/mergo v0.3.7
	github.com/inconshreveable/mousetrap v1.0.0
	github.com/json-iterator/go v1.1.7
	github.com/magiconair/properties v1.7.4
	github.com/mitchellh/mapstructure v0.0.0-20180111000720-b4575eea38cc
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd
	github.com/modern-go/reflect2 v1.0.1
	github.com/onsi/ginkgo v1.10.2
	github.com/onsi/gomega v1.5.0
	github.com/pelletier/go-toml v1.0.1
	github.com/petar/GoLLRB v0.0.0-20130427215148-53be0d36a84c
	github.com/peterbourgon/diskv v2.0.1+incompatible
	github.com/spf13/afero v1.2.2
	github.com/spf13/cast v1.1.0
	github.com/spf13/cobra v0.0.3
	github.com/spf13/jwalterweatherman v0.0.0-20180109140146-7c0cea34c8ec
	github.com/spf13/pflag v1.0.3
	github.com/spf13/viper v1.0.0
	github.com/wadey/gocovmerge v0.0.0-20160331181800-b5bfa59ec0ad
	golang.org/x/crypto v0.0.0-20190911031432-227b76d455e7
	golang.org/x/net v0.0.0-20190912160710-24e19bdeb0f2
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	golang.org/x/sys v0.0.0-20190912141932-bc967efca4b8
	golang.org/x/text v0.3.2
	golang.org/x/time v0.0.0-20190308202827-9d24e82272b4
	golang.org/x/tools v0.0.0-20190624190245-7f2218787638
	google.golang.org/appengine v1.6.2
	gopkg.in/inf.v0 v0.9.1
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7
	gopkg.in/yaml.v2 v2.2.2
	istio.io/api v0.0.0-20191115173247-e1a1952e5b81
	k8s.io/api v0.0.0-20190918195907-bd6ac527cfd2
	k8s.io/apimachinery v0.0.0-20190913080033-27d36303b655
	k8s.io/client-go v0.0.0-20190620085101-78d2af792bab
	k8s.io/code-generator v0.0.0-20190923155300-6206bfaf5c98 // indirect
)

replace (
	github.com/golang/protobuf => github.com/golang/protobuf v1.3.2
	k8s.io/api => k8s.io/api v0.0.0-20191003000013-35e20aa79eb8
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190913080033-27d36303b655
	k8s.io/client-go => k8s.io/client-go v0.0.0-20190918200256-06eb1244587a
	k8s.io/code-generator => k8s.io/code-generator v0.0.0-20190927045949-f81bca4f5e85
)
