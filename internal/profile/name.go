package profile

import (
	"fmt"
	"regexp"
)

var namePattern = regexp.MustCompile(`^[a-z][a-z0-9_-]*$`)

func ValidateName(name string) error {
	if !namePattern.MatchString(name) {
		return fmt.Errorf("invalid profile name %q: start with a lowercase letter (a-z). The remaining characters can be a-z, 0-9, '-' or '_' (example: work-2)", name)
	}
	return nil
}
