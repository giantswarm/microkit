module github.com/giantswarm/microkit

go 1.25.0

toolchain go1.26.6

require (
	github.com/giantswarm/microerror v0.4.1
	github.com/giantswarm/micrologger v1.1.2
	github.com/giantswarm/versionbundle v1.2.0
	github.com/go-kit/kit v0.13.0
	github.com/gorilla/mux v1.8.1
	github.com/prometheus/client_golang v1.24.1
	github.com/spf13/cobra v1.10.2
	github.com/spf13/pflag v1.0.10
	github.com/spf13/viper v1.21.0
	go.yaml.in/yaml/v3 v3.0.5
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/coreos/go-semver v0.3.1 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/go-kit/log v0.2.1 // indirect
	github.com/go-logfmt/logfmt v0.6.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-stack/stack v1.8.1 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.70.1 // indirect
	github.com/prometheus/procfs v0.21.1 // indirect
	github.com/sagikazarmark/locafero v0.11.0 // indirect
	github.com/sourcegraph/conc v0.3.1-0.20240121214520-5f936abd7ae8 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	golang.org/x/net v0.58.0 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/text v0.41.0 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
	gopkg.in/resty.v1 v1.12.0 // indirect
)

replace (
	github.com/prometheus/client_golang => github.com/prometheus/client_golang v1.24.1
	golang.org/x/crypto => golang.org/x/crypto v0.55.0
	golang.org/x/net => golang.org/x/net v0.53.0
)

replace github.com/nats-io/nats-server/v2 v2.8.4 => github.com/nats-io/nats-server/v2 v2.14.6

replace github.com/rabbitmq/amqp091-go v1.2.0 => github.com/rabbitmq/amqp091-go v1.14.0

replace github.com/sirupsen/logrus v1.8.1 => github.com/sirupsen/logrus v1.10.2

replace github.com/yuin/goldmark v1.4.13 => github.com/yuin/goldmark v1.8.5

replace golang.org/x/mod v0.37.0 => golang.org/x/mod v0.40.0

replace google.golang.org/grpc v1.40.0 => google.golang.org/grpc v1.83.2
