package repositories

import "github.com/giantswarm/microerror"

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

var repositoryNotFoundError = &microerror.Error{
	Kind: "repositoryNotFoundError",
}

var dependenciesNotFoundError = &microerror.Error{
	Kind: "dependenciesNotFoundError",
	Desc: "Please enable the dependency graph feature for this repository",
}
