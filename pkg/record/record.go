package record

import (
	"fmt"
)

// RegistrationRecord represents DNS record entry.
type RegistrationRecord struct {
	Name       string `json:"name" yaml:"name"`
	Type       string `json:"type" yaml:"type"`
	TimeToLive uint64 `json:"ttl" yaml:"ttl"`
	Version4   bool   `json:"v4" yaml:"v4"`
	Version6   bool   `json:"v6" yaml:"v6"`
	ip4        string
	ip6        string
}

// Validate validates RegistrationRecord.
func (r *RegistrationRecord) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("dns record name is empty")
	}
	if r.Type != "" {
		if r.Type != "A" && r.Type != "AAAA" && r.Type != "ALL" {
			return fmt.Errorf("dns record type %s is invalid, must be one of the following: A, AAAA, or ALL", r.Type)
		}
	} else {
		r.Type = "A"
	}
	r.Version4 = false
	r.Version6 = false
	if r.Type == "A" || r.Type == "ALL" {
		r.Version4 = true
	}
	if r.Type == "AAAA" || r.Type == "ALL" {
		r.Version6 = true
	}
	if r.TimeToLive == 0 {
		r.TimeToLive = 600
	}
	return nil
}

func validVersion(version int) error {
	if version != 4 && version != 6 {
		return fmt.Errorf("invalid ip version %d", version)
	}
	return nil
}

// SetAddress sets IP address associated with the record.
func (r *RegistrationRecord) SetAddress(addr string, version int) error {
	if err := validVersion(version); err != nil {
		return fmt.Errorf("%s for %s", err, addr)
	}
	if version == 4 {
		r.ip4 = addr
		return nil
	}
	r.ip6 = addr
	return nil

}

// GetAddress returns IP address associated with the record.
func (r *RegistrationRecord) GetAddress(version int) (string, error) {
	if err := validVersion(version); err != nil {
		return "", err
	}
	if version == 4 {
		return r.ip4, nil
	}
	return r.ip6, nil
}
