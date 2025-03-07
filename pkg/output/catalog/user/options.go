package user

func WithNamespace(namespace string) Option {
	return func(c *User) {
		c.Namespace = namespace
	}
}

func WithEmail(email string) Option {
	return func(c *User) {
		c.Email = email
	}
}

func WithDescription(description string) Option {
	return func(c *User) {
		c.Description = description
	}
}

func WithDisplayName(displayName string) Option {
	return func(c *User) {
		c.DisplayName = displayName
	}
}

func WithTitle(title string) Option {
	return func(c *User) {
		c.Title = title
	}
}

func WithPictureURL(url string) Option {
	return func(c *User) {
		c.PictureURL = url
	}
}

func WithGroups(names ...string) Option {
	return func(c *User) {
		c.Groups = names
	}
}

func WithGitHubHandle(handle string) Option {
	return func(c *User) {
		c.GitHubHandle = handle
	}
}

func WithGitHubID(id int64) Option {
	return func(c *User) {
		c.GitHubID = id
	}
}
