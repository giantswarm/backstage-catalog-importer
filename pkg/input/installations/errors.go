package installations

import "github.com/giantswarm/microerror"

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

var fileNotFoundError = &microerror.Error{
	Kind: "fileNotFoundError",
}

// IsFileNotFoundError returns true if error is fileNotFoundError.
func IsFileNotFoundError(err error) bool {
	return microerror.Cause(err) == fileNotFoundError
}
