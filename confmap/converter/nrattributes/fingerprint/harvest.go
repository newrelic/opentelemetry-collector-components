package fingerprint

type HostHarvester struct{}

// Harvester harvest agent's fingerprint.
type Harvester interface {
	Harvest() (fp Fingerprint, err error)
}

func (h *HostHarvester) Harvest() (fp Fingerprint, err error) {
	return Fingerprint{}, nil
}
