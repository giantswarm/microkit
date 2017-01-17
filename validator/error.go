package validator

import (
	"github.com/juju/errgo"
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
	_, ok := ToUnknownAttribute(err)
	return ok
}

// ToUnknownAttribute tries to assert the given error to UnknownAttributeError
// and returns the asserted error and true in case this was successful.
func ToUnknownAttribute(err error) (UnknownAttributeError, bool) {
	e, ok := errgo.Cause(err).(UnknownAttributeError)
	return e, ok
}
