package githubpagination

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gofri/go-github-pagination/githubpagination/drivers"
)

type Config struct {
	Disabled       bool
	DefaultPerPage int
	MaxNumOfPages  int
	Driver         PaginationDriver
}

type ConfigOverridesKey struct{}

func newConfig(opts ...Option) *Config {
	var c Config
	c.ApplyOptions(opts...)
	return &c
}

// ApplyOptions applies the options to the config.
func (c *Config) ApplyOptions(opts ...Option) {
	for _, o := range opts {
		if o == nil {
			continue
		}
		o(c)
	}
}

// GetContextedConfig returns the config overrides from the context, if any.
func (c *Config) GetContextedConfig(ctx context.Context) *Config {
	overrides := GetConfigOverrides(ctx)
	if overrides == nil {
		// no config override - use the default config (zero-copy)
		return c
	}
	reqConfig := *c
	reqConfig.ApplyOptions(overrides...)
	return &reqConfig
}

// GetRequestConfig returns the config overrides from the request context, if any.
func (c *Config) GetRequestConfig(request *http.Request) *Config {
	return c.GetContextedConfig(request.Context())
}

func (c *Config) UpdateRequest(request *http.Request) *http.Request {
	if c.DefaultPerPage == 0 {
		return request
	}
	query := request.URL.Query()
	query.Set("per_page", strconv.Itoa(c.DefaultPerPage))
	request.URL.RawQuery = query.Encode()
	return request
}

func (c *Config) IsPaginationOverflow(pageCount int) bool {
	return c.MaxNumOfPages > 0 && pageCount > c.MaxNumOfPages
}

func (c *Config) GetDriver() PaginationDriver {
	if c.Driver != nil {
		return c.Driver
	}
	return drivers.NewSyncPaginationDriver()
}

// WithOverrideConfig adds config overrides to the context.
// The overrides are applied on top of the existing config.
// Allows for request-specific overrides.
func WithOverrideConfig(ctx context.Context, opts ...Option) context.Context {
	overrides := GetConfigOverrides(ctx)
	opts = append(opts, overrides...) // append on multiple calls
	return context.WithValue(ctx, ConfigOverridesKey{}, opts)
}

// GetConfigOverrides returns the config overrides from the context, if any.
func GetConfigOverrides(ctx context.Context) []Option {
	cfg := ctx.Value(ConfigOverridesKey{})
	if cfg == nil {
		return nil
	}
	return cfg.([]Option)
}
