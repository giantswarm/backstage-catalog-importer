package githuburl

import "github.com/giantswarm/microerror"

var invalidURLError = &microerror.Error{
	Kind: "invalidURLError",
}

var fetchError = &microerror.Error{
	Kind: "fetchError",
}

var parseError = &microerror.Error{
	Kind: "parseError",
}
