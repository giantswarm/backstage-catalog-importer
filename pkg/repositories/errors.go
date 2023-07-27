package repositories

import "github.com/giantswarm/microerror"

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

var repositoryNotFoundError = &microerror.Error{
	Kind: "repositoryNotFoundError",
}
