package api

// Option is a functional option for configuring an API.
type Option func(*API)

// WithNamespace sets the Backstage namespace.
func WithNamespace(namespace string) Option {
	return func(a *API) {
		if namespace != "" {
			a.Namespace = namespace
		}
	}
}

// WithTitle sets the display title.
func WithTitle(title string) Option {
	return func(a *API) {
		a.Title = title
	}
}

// WithDescription sets the description.
func WithDescription(description string) Option {
	return func(a *API) {
		a.Description = description
	}
}

// WithOwner sets the owner reference.
func WithOwner(owner string) Option {
	return func(a *API) {
		if owner != "" {
			a.Owner = owner
		}
	}
}

// WithType sets the API type (e.g., "crd", "openapi").
func WithType(apiType string) Option {
	return func(a *API) {
		if apiType != "" {
			a.Type = apiType
		}
	}
}

// WithLifecycle sets the lifecycle stage.
func WithLifecycle(lifecycle string) Option {
	return func(a *API) {
		if lifecycle != "" {
			a.Lifecycle = lifecycle
		}
	}
}

// WithSystem sets the system reference.
func WithSystem(system string) Option {
	return func(a *API) {
		a.System = system
	}
}

// WithDefinition sets the API definition content.
func WithDefinition(definition string) Option {
	return func(a *API) {
		a.Definition = definition
	}
}

// WithTags adds tags to the API.
func WithTags(tags ...string) Option {
	return func(a *API) {
		a.Tags = append(a.Tags, tags...)
	}
}

// WithLabels sets the labels map.
func WithLabels(labels map[string]string) Option {
	return func(a *API) {
		a.Labels = labels
	}
}
