package crdconfig

import "github.com/giantswarm/microerror"

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

var fileNotFoundError = &microerror.Error{
	Kind: "fileNotFoundError",
}

var readError = &microerror.Error{
	Kind: "readError",
}

var parseError = &microerror.Error{
	Kind: "parseError",
}

var validationError = &microerror.Error{
	Kind: "validationError",
}
