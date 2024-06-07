// Package legacy contains deprecated functions that are being replaced soon.
package legacy

import (
	bscatalog "github.com/giantswarm/backstage-catalog-importer/pkg/bscatalog/v1alpha1"
)

// CreateUserEntity is the deprecated way of generating a user entity.
//
// Deprecated: Create a catalog.User struct and a ToEntity() method to use instead.
func CreateUserEntity(name, email, displayName, description, avatarURL string) bscatalog.Entity {
	e := bscatalog.Entity{
		APIVersion: bscatalog.APIVersion,
		Kind:       bscatalog.EntityKindUser,
		Metadata: bscatalog.EntityMetadata{
			Name: name,
		},
	}

	spec := bscatalog.UserSpec{
		MemberOf: []string{},
		Profile: bscatalog.UserProfile{
			Email: email,
		},
	}

	if description != "" {
		e.Metadata.Description = description
	}
	if displayName != "" {
		spec.Profile.DisplayName = displayName
	}
	if avatarURL != "" {
		spec.Profile.Picture = avatarURL
	}

	e.Spec = spec

	return e
}
