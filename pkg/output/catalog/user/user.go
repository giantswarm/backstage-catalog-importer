// Represents a user to be exported into a Backstage catalog as a User entity.
package user

import (
	"fmt"
	"slices"

	bscatalog "github.com/giantswarm/backstage-catalog-importer/pkg/output/bscatalog/v1alpha1"
)

// Option is an option to configure a User.
type Option func(*User)

// User holds our internal representation of something that we want
// to export as a User entity.
type User struct {
	// User name (required)
	Name string

	// Namespace, defaults to "default"
	Namespace string

	// Display title of the user
	Title string

	Description string
	Email       string
	PictureURL  string

	// Names of groups the user is a member of
	Groups []string
}

func New(name string, options ...Option) (*User, error) {
	if name == "" {
		return nil, fmt.Errorf("name must not be empty")
	}

	c := &User{
		Name:      name,
		Namespace: "default",
	}

	for _, option := range options {
		option(c)
	}

	return c, nil
}

// Returns an entity representation of the user.
func (c *User) ToEntity() *bscatalog.Entity {
	slices.Sort(c.Groups)

	spec := bscatalog.UserSpec{
		Profile: bscatalog.UserProfile{
			DisplayName: c.Title,
			Picture:     c.PictureURL,
			Email:       c.Email,
		},
		MemberOf: c.Groups,
	}

	e := &bscatalog.Entity{
		APIVersion: "backstage.io/v1alpha1",
		Kind:       bscatalog.EntityKindUser,
		Metadata: bscatalog.EntityMetadata{
			Name:        c.Name,
			Description: c.Description,
			Namespace:   c.Namespace,
			Title:       c.Title,
		},
		Spec: spec,
	}

	return e
}
