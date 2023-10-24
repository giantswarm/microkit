module github.com/giantswarm/microkit

go 1.19

require (
	github.com/giantswarm/microerror v0.4.0
	github.com/giantswarm/micrologger v0.6.0
	github.com/giantswarm/versionbundle v1.0.0
	github.com/go-kit/kit v0.12.0
	github.com/gorilla/mux v1.8.0
	github.com/prometheus/client_golang v1.11.0
	github.com/spf13/cobra v1.3.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.10.0
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/fsnotify/fsnotify v1.5.1 // indirect
	github.com/go-kit/log v0.2.0 // indirect
	github.com/go-logfmt/logfmt v0.5.1 // indirect
	github.com/go-stack/stack v1.8.1 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/magiconair/properties v1.8.5 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/pelletier/go-toml v1.9.4 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.30.0 // indirect
	github.com/prometheus/procfs v0.7.3 // indirect
	github.com/spf13/afero v1.6.0 // indirect
	github.com/spf13/cast v1.4.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/subosito/gotenv v1.2.0 // indirect
	golang.org/x/net v0.13.0 // indirect
	golang.org/x/sync v0.2.0 // indirect
	golang.org/x/sys v0.10.0 // indirect
	golang.org/x/text v0.11.0 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
	gopkg.in/ini.v1 v1.66.2 // indirect
	gopkg.in/resty.v1 v1.12.0 // indirect
)

replace (
	github.com/hashicorp/consul/api => github.com/hashicorp/consul/api v1.25.1
	github.com/hashicorp/consul/sdk => github.com/hashicorp/consul/sdk v0.14.1
)
