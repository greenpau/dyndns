package record

import (
	"fmt"
)

// RegistrationRecord represents DNS record entry.
type RegistrationRecord struct {
	Name       string `json:"name" yaml:"name"`
	Type       string `json:"type" yaml:"type"`
	TimeToLive uint64 `json:"ttl" yaml:"ttl"`
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
	if r.TimeToLive == 0 {
		r.TimeToLive = 600
	}
	return nil
}
