package validator

import (
	"fmt"

	"github.com/giantswarm/microerror"
)

// UnknownAttributeError indicates there was an error due to unknown attributes
// within validated data structures.
type UnknownAttributeError struct {
	attribute string
	message   string
}

// Attribute returns the detected unknown attribute.
func (e UnknownAttributeError) Attribute() string {
	return e.attribute
}

// Error returns the actual error message of the UnknownAttributeError to
// implement the error interface.
func (e UnknownAttributeError) Error() string {
	return e.message
}

// IsUnknownAttribute asserts UnknownAttributeError.
func IsUnknownAttribute(err error) bool {
	_, ok := microerror.Cause(err).(UnknownAttributeError)
	return ok
}

// ToUnknownAttribute tries asserts the given error to UnknownAttributeError and
// returns it. ToUnknownAttribute panics in case the underlying error is not of
// type UnknownAttributeError. Therefore IsUnknownAttribute should always be
// used to verify the safe execution of ToUnknownAttribute beforehand.
func ToUnknownAttribute(err error) UnknownAttributeError {
	return microerror.Cause(err).(UnknownAttributeError)
}

// UnknownAttribute takes an arbitrary map and a map obtaining some expected
// structure. The first argument might represent an incoming request of some
// microservice. The second argument should then represent the datastructure of
// the associated request as it is expected to be provided. In case received
// contains fields which are not available in expected, an UnknownAttributeError
// is returned.
func UnknownAttribute(received, expected map[string]interface{}) error {
	for r := range received {
		var found bool

		for e := range expected {
			_, eIsMapStringInterface := expected[e].(map[string]interface{})
			_, rIsMapStringInterface := received[r].(map[string]interface{})

			if eIsMapStringInterface && rIsMapStringInterface {
				err := UnknownAttribute(received[r].(map[string]interface{}), expected[e].(map[string]interface{}))
				if err != nil {
					return microerror.Mask(err)
				}
			}

			if e == r {
				found = true
				break
			}
		}

		if found {
			continue
		}

		err := UnknownAttributeError{
			attribute: r,
			message:   fmt.Sprintf("unknown attribute: %s", r),
		}

		return microerror.Mask(err)
	}

	return nil
}
