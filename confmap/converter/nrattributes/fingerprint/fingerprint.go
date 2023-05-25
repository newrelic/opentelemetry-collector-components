package fingerprint

// Fingerprint is used in the agent connect step when communicating with the backend. Based on it
// the backend will uniquely identify the agent and respond with the entityKey and entityId.
type Fingerprint struct {
	FullHostname    string    `json:"fullHostname"`
	Hostname        string    `json:"hostname"`
	CloudProviderId string    `json:"cloudProviderId"`
	DisplayName     string    `json:"displayName"`
	BootID          string    `json:"bootId"`
	IpAddresses     Addresses `json:"ipAddresses"`
	MacAddresses    Addresses `json:"macAddresses"`
}

// Addresses will store the nic addresses mapped by the nickname.
type Addresses map[string][]string
