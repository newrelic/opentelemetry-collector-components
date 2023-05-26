package entityguidprocessor

import (
	"fmt"
	"io"

	"encoding/json"
	"net/http"
)

type statusEntity struct {
	Guid string
}

func getEntityGuid(uri string) (string, error) {

	client := http.Client{}

	// send a HTTP GET request
	resp, err := client.Get(uri)
	if err != nil {
		return "", fmt.Errorf("unable to download the file via HTTP GET for uri %q: %w ", uri, err)
	}
	defer resp.Body.Close()

	// check the HTTP status code
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to load resource from uri %q. status code: %d", uri, resp.StatusCode)
	}

	// read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("fail to read the response body from uri %q: %w", uri, err)
	}

	var currentGuid statusEntity

	err = json.Unmarshal(body, &currentGuid)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshall staus api body: %w", err)
	}

	return currentGuid.Guid, nil
}
