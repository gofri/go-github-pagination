package github_pagination

import (
	"context"
	"net/http"
)

type Config struct {
	Disabled       bool
	DefaultPerPage int
	MaxNumOfPages  int
}

type ConfigOverridesKey struct{}

// ApplyOptions applies the options to the config.
func (c *Config) ApplyOptions(opts ...Option) {
	for _, o := range opts {
		if o == nil {
			continue
		}
		o(c)
	}
}

// GetRequestConfig returns the config overrides from the context, if any.
func (c *Config) GetRequestConfig(request *http.Request) *Config {
	overrides := GetConfigOverrides(request.Context())
	if overrides == nil {
		// no config override - use the default config (zero-copy)
		return c
	}
	reqConfig := *c
	reqConfig.ApplyOptions(overrides...)
	return &reqConfig
}

// WithOverrideConfig adds config overrides to the context.
// The overrides are applied on top of the existing config.
// Allows for request-specific overrides.
func WithOverrideConfig(ctx context.Context, opts ...Option) context.Context {
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
