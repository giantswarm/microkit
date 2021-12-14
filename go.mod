module github.com/giantswarm/microkit

go 1.14

require (
	github.com/giantswarm/microerror v0.4.0
	github.com/giantswarm/micrologger v0.3.1
	github.com/giantswarm/versionbundle v0.2.0
	github.com/go-kit/kit v0.12.0
	github.com/gorilla/mux v1.8.0
	github.com/prometheus/client_golang v1.11.0
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1
	gopkg.in/yaml.v2 v2.4.0
)

// Apply fix for CVE-2020-15114 not yet released in github.com/spf13/viper.
replace github.com/bketelsen/crypt => github.com/bketelsen/crypt v0.0.3
