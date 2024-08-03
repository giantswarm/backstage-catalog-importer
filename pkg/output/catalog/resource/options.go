package resource

// Option is an option to configure a Resource.
type Option func(*Resource)

func WithNamespace(namespace string) Option {
	return func(c *Resource) {
		c.Namespace = namespace
	}
}

func WithDescription(description string) Option {
	return func(c *Resource) {
		c.Description = description
	}
}

func WithTitle(title string) Option {
	return func(c *Resource) {
		c.Title = title
	}
}

func WithType(t string) Option {
	return func(c *Resource) {
		c.Type = t
	}
}
