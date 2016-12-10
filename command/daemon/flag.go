package daemon

import (
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Flags is the global flag structure used to apply certain configuration to it.
// This is used to bundle configuration for the command, server and service
// initialisation.
var Flags = struct {
	Config struct {
		Dir  string
		File string
	}
	Server struct {
		Listen struct {
			Address string
		}
	}
}{}

// MergeFlags merges the given flag set with an internal viper configuration.
// That way command line flags, environment variables and config files will be
// merged.
func MergeFlags(fs *pflag.FlagSet) error {
	v := viper.New()

	// Check the defined config file.
	v.AddConfigPath(Flags.Config.Dir)
	v.SetConfigName(Flags.Config.File)
	err := v.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// In case there is no config file given we simply go ahead to check the
			// process environment.
		} else {
			return maskAny(err)
		}
	}

	// We merge the defined flags with their upper case counterparts from the
	// environment .
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	v.BindPFlags(fs)

	fs.VisitAll(func(f *pflag.Flag) {
		if f.Changed {
			// The current flag was set via the command line. We definitly want to use
			// the set value. Therefore we do not merge anything into it.
			return
		}
		if !v.IsSet(f.Name) {
			// There is neither configuration in the provided config file nor in the
			// process environment. That means we cannot use it to merge it into any
			// defined flag.
			return
		}

		f.Value.Set(v.GetString(f.Name))
	})

	return nil
}
