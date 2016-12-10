package logger

import (
	"os"
	"time"

	kitlog "github.com/go-kit/kit/log"
)

// Config represents the configuration used to create a new logger.
type Config struct {
	// Settings.
	TimestampFormatter kitlog.Valuer
}

// DefaultConfig provides a default configuration to create a new logger by best
// effort.
func DefaultConfig() Config {
	return Config{
		// Settings.
		TimestampFormatter: func() interface{} {
			return time.Now().UTC().Format("06-01-02 15:04:05.000")
		},
	}
}

// New creates a new configured logger.
func New(config Config) (Logger, error) {
	// Settings.
	if config.TimestampFormatter == nil {
		return nil, maskAnyf(invalidConfigError, "timestamp formatter must not be empty")
	}

	kitLogger := kitlog.NewJSONLogger(kitlog.NewSyncWriter(os.Stdout))
	kitLogger = kitlog.NewContext(kitLogger).With(
		"ts", config.TimestampFormatter,
		"caller", kitlog.DefaultCaller,
	)

	newLogger := &logger{
		Logger: kitLogger,
	}

	return newLogger, nil
}

// Logger implements a logging interface used to log messages.
type logger struct {
	Logger kitlog.Logger
}

func (l *logger) Log(keyvals ...interface{}) error {
	return l.Logger.Log(keyvals...)
}
