package entityprovider

import "go.opentelemetry.io/collector/confmap"

func New() confmap.Provider {
	return new(HTTPScheme)
}
