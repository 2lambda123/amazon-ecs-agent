module github.com/aws/amazon-ecs-agent/agent

go 1.19

require (
	github.com/aws/amazon-ecs-agent/ecs-agent v0.0.0
	github.com/aws/aws-sdk-go v1.44.173
	github.com/awslabs/go-config-generator-for-fluentd-and-fluentbit v0.0.0-20210308162251-8959c62cb8f9
	github.com/cihub/seelog v0.0.0-20170130134532-f561c5e57575
	github.com/container-storage-interface/spec v1.9.0
	github.com/containerd/cgroups/v3 v3.0.3
	github.com/containernetworking/cni v1.1.2
	github.com/containernetworking/plugins v1.5.1
	github.com/deniswernert/udev v0.0.0-20170418162847-a12666f7b5a1
	github.com/docker/docker v24.0.6+incompatible
	github.com/docker/go-connections v0.4.0
	github.com/docker/go-units v0.5.0
	github.com/fsnotify/fsnotify v1.7.0
	github.com/golang/mock v1.6.0
	github.com/gorilla/mux v1.8.0
	github.com/hectane/go-acl v0.0.0-20190604041725-da78bae5fc95
	github.com/opencontainers/image-spec v1.1.0-rc2.0.20221005185240-3a7f492d3f1b
	github.com/opencontainers/runtime-spec v1.2.0
	github.com/pborman/uuid v1.2.1
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.19.0
	github.com/prometheus/client_model v0.6.0
	github.com/stretchr/testify v1.8.4
	github.com/vishvananda/netlink v1.2.1-beta.2
	go.etcd.io/bbolt v1.3.7
	golang.org/x/sys v0.21.0
	golang.org/x/tools v0.21.0
	google.golang.org/grpc v1.62.0
	google.golang.org/protobuf v1.33.0
	k8s.io/api v0.29.1
)

require (
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/cilium/ebpf v0.11.0 // indirect
	github.com/coreos/go-systemd/v22 v22.5.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/didip/tollbooth v4.0.2+incompatible // indirect
	github.com/docker/distribution v2.8.2+incompatible // indirect
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/godbus/dbus/v5 v5.1.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/moby/term v0.0.0-20221205130635-1aeaba878587 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/common v0.53.0 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	github.com/rootless-containers/rootlesskit v1.1.1 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/vishvananda/netns v0.0.4 // indirect
	golang.org/x/exp v0.0.0-20230224173230-c95f2b4c22f2 // indirect
	golang.org/x/mod v0.17.0 // indirect
	golang.org/x/net v0.25.0 // indirect
	golang.org/x/text v0.15.0 // indirect
	golang.org/x/time v0.3.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240123012728-ef4313101c80 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/apimachinery v0.29.1 // indirect
	k8s.io/klog/v2 v2.110.1 // indirect
	k8s.io/utils v0.0.0-20230726121419-3b25d923346b // indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.4.1 // indirect
)

replace github.com/aws/amazon-ecs-agent/ecs-agent => ../ecs-agent
