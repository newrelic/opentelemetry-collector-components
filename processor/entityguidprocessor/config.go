package entityguidprocessor

import (
	"errors"

	"go.opentelemetry.io/collector/component"
)

type Config struct {
	StatusEndpoint string `mapstructure:"status_endpoint"`
}

var _ component.Config = (*Config)(nil)

// Validate checks if the processor configuration is valid
func (cfg *Config) Validate() error {
	if cfg.StatusEndpoint == "" {
		return errors.New("missing required field \"status_endpoint\"")
	}
	return nil
}
