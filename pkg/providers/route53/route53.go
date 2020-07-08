package route53

import (
	"fmt"
	"github.com/greenpau/dyndns/pkg/record"

	"go.uber.org/zap"
)

// RegistrationProvider is a controller for updating DNS records hosted byo
// AWS Route 53 service.
type RegistrationProvider struct {
	Provider    string `json:"type" yaml:"type"`
	ZoneID      string `json:"zone_id" yaml:"zone_id"`
	Credentials string `json:"credentials" yaml:"credentials"`
	ProfileName string `json:"profile" yaml:"profile"`
	log         *zap.Logger
}

// Validate validates an instance op *RegistrationProvider.
func (p *RegistrationProvider) Validate() error {
	if p.ZoneID == "" {
		return fmt.Errorf("provider requires a hosted zone id")
	}
	if p.Credentials == "" {
		return fmt.Errorf("aws credentials not found")
	}
	if p.Provider != "route53" {
		return fmt.Errorf("provider mismatch: %s (config) vs. route53 (expected)", p.Provider)
	}
	return nil
}

// Configure configures  an instance op *RegistrationProvider.
func (p *RegistrationProvider) Configure(logger *zap.Logger) error {
	p.log = logger
	if err := p.Validate(); err != nil {
		return err
	}
	if p.ProfileName == "" {
		p.ProfileName = "default"
	}
	return nil
}

// GetProvider returns the provider name associated with RegistrationProvider.
func (p *RegistrationProvider) GetProvider() string {
	return p.Provider
}

// Register registers a record with RegistrationProvider.
func (p *RegistrationProvider) Register(r *record.RegistrationRecord) error {
	p.log.Debug("received registration request", zap.Any("request", r))
	return nil
}
