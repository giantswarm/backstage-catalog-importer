package ociregistry

import "github.com/giantswarm/microerror"

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

var couldNotCreateRegistryClientError = &microerror.Error{
	Kind: "couldNotCreateRegistryClientError",
}

var couldNotGetRepositoryError = &microerror.Error{
	Kind: "couldNotGetRepositoryError",
}

var couldNotGetRepositoryTagsError = &microerror.Error{
	Kind: "couldNotGetRepositoryTagsError",
}

var couldNotResolveTagError = &microerror.Error{
	Kind: "couldNotResolveTagError",
}

var couldNotGetRepositoryManifestError = &microerror.Error{
	Kind: "couldNotGetRepositoryManifestError",
}

var couldNotReadManifestError = &microerror.Error{
	Kind: "couldNotReadManifestError",
}

var couldNotUnmarshalManifestError = &microerror.Error{
	Kind: "couldNotUnmarshalManifestError",
}

var couldNotFetchConfigBlobError = &microerror.Error{
	Kind: "couldNotFetchConfigBlobError",
}

var couldNotReadConfigError = &microerror.Error{
	Kind: "couldNotReadConfigError",
}

var couldNotUnmarshalConfigError = &microerror.Error{
	Kind: "couldNotUnmarshalConfigError",
}
