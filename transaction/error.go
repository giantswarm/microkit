package transaction

import (
	"github.com/juju/errgo"
)

var invalidConfigError = errgo.New("invalid config")

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return errgo.Cause(err) == invalidConfigError
}

var invalidExecutionError = errgo.New("invalid execution")

// IsInvalidExecution asserts invalidExecutionError.
func IsInvalidExecution(err error) bool {
	return errgo.Cause(err) == invalidExecutionError
}
