// Validator provides validation primitives for microservices.
package validator

import (
	"fmt"

	microerror "github.com/giantswarm/microkit/error"
)

// UnknownAttributes takes an arbitrary map and a map obtaining some expected
// structure. The first argument might represent an incoming request of some
// microservice. The second argument should then represent the datastructure of
// the associated request as it is expected to be provided. In case received
// contains fields which are not available in expected, an UnknownAttributeError
// is returned.
func UnknownAttributes(received, expected map[string]interface{}) error {
	for r, _ := range received {
		var found bool

		for e, _ := range expected {
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

		return microerror.MaskAny(err)
	}

	return nil
}
