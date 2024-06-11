package group

func WithNamespace(namespace string) Option {
	return func(c *Group) {
		c.Namespace = namespace
	}
}

func WithDescription(description string) Option {
	return func(c *Group) {
		c.Description = description
	}
}

func WithTitle(title string) Option {
	return func(c *Group) {
		c.Title = title
	}
}

func WithType(t string) Option {
	return func(c *Group) {
		c.Type = t
	}
}

func WithPictureURL(url string) Option {
	return func(c *Group) {
		c.PictureURL = url
	}
}

func WithGrafanaDashboardSelector(selector string) Option {
	return func(c *Group) {
		c.GrafanaDashboardSelector = selector
	}
}

func WithOpsgenieTeamName(name string) Option {
	return func(c *Group) {
		c.OpsgenieTeamName = name
	}
}

func WithChildrenNames(names ...string) Option {
	return func(c *Group) {
		c.ChildrenNames = names
	}
}

func WithParentName(name string) Option {
	return func(c *Group) {
		c.ParentName = name
	}
}

func WithMemberNames(names ...string) Option {
	return func(c *Group) {
		c.MemberNames = names
	}
}
