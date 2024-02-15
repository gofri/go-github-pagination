package github_pagination

const (
	MaxPerPage = 100
)

type Option func(*Config)

// WithPaginationEnabled enables the pagination for paginated requests.
// This is the default behavior.
// This may be used to override a previous WithPaginationDisabled option,
// e.g., on a per-request basis.
func WithPaginationEnabled() Option {
	return func(c *Config) {
		c.Disabled = false
	}
}

// WithPaginationDisabled disables the pagination for paginated requests.
// This may be used to override a previous WithPaginationEnabled option,
// e.g., on a per-request basis.
func WithPaginationDisabled() Option {
	return func(c *Config) {
		c.Disabled = true
	}
}

// WithPerPage sets the default per-page value for paginated requests.
func WithPerPage(perPage int) Option {
	return func(c *Config) {
		c.DefaultPerPage = perPage
	}
}

// WithMaxNumOfPages sets the maximum number of pages for paginated requests.
// This enables the client to limit the number of pages to be fetched.
func WithMaxNumOfPages(maxNumOfPages int) Option {
	return func(c *Config) {
		c.MaxNumOfPages = maxNumOfPages
	}
}
