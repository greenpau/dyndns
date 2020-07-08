package record

import (
	"fmt"
)

// RegistrationRecord represents DNS record entry.
type RegistrationRecord struct {
	Name       string `json:"name" yaml:"name"`
	Type       string `json:"type" yaml:"type"`
	TimeToLive uint64 `json:"ttl" yaml:"ttl"`
	Version4   bool
	Version6   bool
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
