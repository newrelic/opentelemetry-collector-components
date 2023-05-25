package nrattributes

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/newrelic/opentelemetry-collector-components/confmap/converter/nrattributes/backend"
	"github.com/newrelic/opentelemetry-collector-components/confmap/converter/nrattributes/entity"
	"github.com/newrelic/opentelemetry-collector-components/confmap/converter/nrattributes/fingerprint"
)

const (
	defaultSecureFedralIdentityURL = "https://gov-identity-api.newrelic.com"
	defaultIdentityURL             = "https://identity-api.newrelic.com"
	defaultIdentityURLEu           = "https://identity-api.eu.newrelic.com"
	defaultIdentityStagingURLEu    = "https://staging-identity-api.eu.newrelic.com"
	defaultIdentityStagingURL      = "https://staging-identity-api.newrelic.com"
	defaultIdentityIngestEndpoint  = "/identity/v1" // default: V1 endpoint root (/connect, /register/batch)
)

func calculateIdentityURL(licenseKey string, staging, fedramp bool) string {
	if staging {
		return calculateIdentityStagingURL(licenseKey)
	}
	if fedramp {
		return defaultSecureFedralIdentityURL
	}
	return calculateIdentityProductionURL(licenseKey)
}

func calculateIdentityStagingURL(licenseKey string) string {
	// only EU supported
	if IsRegionEU(licenseKey) {
		return defaultIdentityStagingURLEu
	}
	return defaultIdentityStagingURL
}

func calculateIdentityProductionURL(licenseKey string) string {
	// only EU supported
	if IsRegionEU(licenseKey) {
		return defaultIdentityURLEu
	}
	return defaultIdentityURL
}

// Client sends a request and returns a response or error.
type Client func(req *http.Request) (*http.Response, error)

type identityClient struct {
	svcUrl           string
	licenseKey       string
	userAgent        string
	compressionLevel int
	containerized    bool

	httpClient Client
}

func newIdentityClient(license string) *identityClient {
	// generate indentityURL
	// TODO: param staging ferdram params
	identityURL := fmt.Sprintf(fmt.Sprintf("%s/%s", calculateIdentityURL(license, false, false), strings.TrimPrefix(defaultIdentityIngestEndpoint, "/")), "/")
	return &identityClient{
		svcUrl:           identityURL,
		licenseKey:       license,
		compressionLevel: gzip.BestCompression,
		containerized:    false,
	}
}

type postConnectBody struct {
	Fingerprint fingerprint.Fingerprint `json:"fingerprint"`
	Type        string                  `json:"type"`
	Protocol    string                  `json:"protocol"`
	EntityID    entity.ID               `json:"entityId,omitempty"`
}

type postConnectResponse struct {
	Identity IdentityResponse `json:"identity"`
}

type IdentityResponse struct {
	EntityId entity.ID `json:"entityId"`
	GUID     string    `json:"GUID"`
}

// ToIdentity converts response into entity identity
func (r *IdentityResponse) ToIdentity() entity.Identity {
	return entity.Identity{
		ID:   r.EntityId,
		GUID: entity.GUID(r.GUID),
	}
}

// agentType returns the type of the agent.
func (ic *identityClient) agentType() string {
	if ic.containerized {
		return "container"
	}
	return "host"
}

func (client *identityClient) Lookup(fingerprint fingerprint.Fingerprint) (entity.Identity, error) {
	buf, err := client.marshal(postConnectBody{
		Fingerprint: fingerprint,
		Type:        client.agentType(),
		Protocol:    "v1",
	})

	if err != nil {
		return entity.Identity{}, err
	}

	req, err := http.NewRequest("POST", client.makeURL("/lookup"), buf)
	if err != nil {
		return entity.Identity{}, fmt.Errorf("connect request failed: %s", err)
	}

	if client.compressionLevel > gzip.NoCompression {
		req.Header.Set("Content-Encoding", "gzip")
	}

	resp, err := client.do(req)
	if err != nil {
		return entity.Identity{}, fmt.Errorf("unable to connect: %s", err)
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			fmt.Println("Error closing ingest body response.")
		}
	}()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return entity.Identity{}, fmt.Errorf("unable to read server response: %s", err)
	}

	// hasError, cause := backend.IsResponseUnsuccessful(resp)

	response := &postConnectResponse{}
	if err = json.Unmarshal(body, response); err != nil {
		return entity.Identity{}, fmt.Errorf("unable to parse connect response: %s", err)
	}

	return response.Identity.ToIdentity(), nil
}

// HELPER FUNCTIONS

// Do performs an http.Request, augmenting it with auth headers
func (ic *identityClient) do(req *http.Request) (*http.Response, error) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", ic.userAgent)
	req.Header.Set(backend.LicenseHeader, ic.licenseKey)

	return ic.httpClient(req)
}

func (ic *identityClient) makeURL(requestPath string) string {
	requestPath = strings.TrimPrefix(requestPath, "/")
	return fmt.Sprintf("%s/%s", ic.svcUrl, requestPath)
}

func (ic *identityClient) marshal(b interface{}) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	if ic.compressionLevel > gzip.NoCompression {
		gzipWriter, err := gzip.NewWriterLevel(&buf, ic.compressionLevel)
		if err != nil {
			return nil, fmt.Errorf("unable to create gzip writer: %v", err)
		}
		defer func() {
			if err := gzipWriter.Close(); err != nil {
				// TODO: move to logger
				fmt.Println("Gzip writer did not close.")
			}
		}()
		if err := json.NewEncoder(gzipWriter).Encode(b); err != nil {
			return nil, fmt.Errorf("gzip writer was not able to write to request body: %s", err)
		}
	} else {
		if err := json.NewEncoder(&buf).Encode(b); err != nil {
			return nil, err
		}
	}
	return &buf, nil
}
