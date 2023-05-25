// Copyright 2020 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0
package nrattributes

import (
	"errors"
	"time"

	"github.com/newrelic/opentelemetry-collector-components/confmap/converter/nrattributes/entity"
	"github.com/newrelic/opentelemetry-collector-components/confmap/converter/nrattributes/fingerprint"
	"github.com/sethvargo/go-retry"
	"go.uber.org/zap"
)

var (
	baseBackoff = 1 * time.Second
)

type IdentityLookupClient interface {
	// lookup for an existing entity guid with the given fingerprint
	Lookup(fingerprint fingerprint.Fingerprint) (entity.Identity, error)
}

type identityLookupService struct {
	logger             *zap.Logger
	fingerprintHarvest fingerprint.Harvester
	client             IdentityLookupClient

	lastFingerprint fingerprint.Fingerprint
}

// ErrEmptyEntityID is returned when the entityID is empty.
var ErrEmptyEntityID = errors.New("the agentID provided is empty. make sure you have connected if this is not expected")

func NewIdentityLookupService(license string) *identityLookupService {
	return &identityLookupService{
		client: newIdentityClient(license),
	}
}

func (ic *identityLookupService) Lookup() entity.Identity {

	backoff := retry.NewExponential(baseBackoff)

	for {

		f, err := ic.fingerprintHarvest.Harvest()
		if err != nil {
			ic.logger.Debug("unable to fetch fingerprint")
			time.Sleep(1 * time.Second)
			continue
		}

		ids, err := ic.client.Lookup(f)

		if !ids.ID.IsEmpty() {
			ic.logger.Sugar().Infow("lookup got id",
				"agent-id", ids.ID,
				"agent-guid", ids.GUID,
			)
			// save fingerprint for later (connect update)
			ic.lastFingerprint = f
			return ids
		}

		if err != nil {
			ic.logger.Sugar().Error("agent lookup attempt failed: %w", err)
		}

		sleepDuration, _ := backoff.Next()
		time.Sleep(sleepDuration)
	}
}
