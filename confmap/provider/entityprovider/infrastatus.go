package entityprovider

import (
	"context"
	"fmt"
	"io"
	"strings"

	"encoding/json"
	"net/http"

	"go.opentelemetry.io/collector/confmap"
)

type SchemeType string

const (
	HTTPScheme SchemeType = "http"
)

type infraStatus struct {
	scheme SchemeType
}

type statusEntity struct {
	Guid string
}

func new(scheme SchemeType) confmap.Provider {
	return &infraStatus{scheme: scheme}
}

func newNRRawConfig(entityGuid string) interface{} {
	return map[string]interface{}{
		"processors": map[string]interface{}{
			"resource/nr": map[string]interface{}{
				"attributes": []map[string]string{
					{
						"key":    "entity.guid",
						"value":  entityGuid,
						"action": "insert",
					},
				},
			},
		},
	}
}

func (infra *infraStatus) Retrieve(_ context.Context, uri string, _ confmap.WatcherFunc) (*confmap.Retrieved, error) {
	if !strings.HasPrefix(uri, string(infra.scheme)+":") {
		return nil, fmt.Errorf("%q uri is not supported by %q provider", uri, string(infra.scheme))
	}

	client := http.Client{}

	// send a HTTP GET request
	resp, err := client.Get(uri)
	if err != nil {
		return nil, fmt.Errorf("unable to download the file via HTTP GET for uri %q: %w ", uri, err)
	}
	defer resp.Body.Close()

	// check the HTTP status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to load resource from uri %q. status code: %d", uri, resp.StatusCode)
	}

	// read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("fail to read the response body from uri %q: %w", uri, err)
	}

	var currentGuid statusEntity

	err = json.Unmarshal(body, &currentGuid)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshall staus api body: %w", err)
	}

	retrievedConf, err := confmap.NewRetrieved(newNRRawConfig(currentGuid.Guid))
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshall staus api body: %w", err)
	}

	return retrievedConf, nil
}

func (infra *infraStatus) Scheme() string {
	return string(infra.scheme)
}

func (*infraStatus) Shutdown(context.Context) error {
	return nil
}
