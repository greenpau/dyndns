package dyndns

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/greenpau/dyndns/pkg/providers/route53"
	"github.com/greenpau/dyndns/pkg/record"
	"go.uber.org/zap"
)

// RegistrationProvider is a receiving instance.
type RegistrationProvider struct {
	config []byte
	engine RegistrationEngine
}

// RegistrationEngine is a receiving instance interface.
type RegistrationEngine interface {
	Configure(*zap.Logger) error
	Validate() error
	GetProvider() string
	Register(*record.RegistrationRecord) error
}

// Register registers DNS record with RegistrationEngine.
func (p *RegistrationProvider) Register(r *record.RegistrationRecord) error {
	return p.engine.Register(r)
}

// GetProvider returns the Provider associated with RegistrationProvider.
func (p *RegistrationProvider) GetProvider() string {
	return p.engine.GetProvider()
}

// Validate validates RegistrationProvider instance.
func (p *RegistrationProvider) Validate() error {
	if p.engine == nil {
		return fmt.Errorf("failed to initialize dns provider instance")
	}
	return p.engine.Validate()
}

// Configure configures RegistrationProvider instance.
func (p *RegistrationProvider) Configure(logger *zap.Logger) error {
	if p.engine == nil {
		return fmt.Errorf("failed to initialize dns provider instance")
	}
	return p.engine.Configure(logger)
}

// MarshalJSON packs configuration of RegistrationProvider JSON byte array
func (p RegistrationProvider) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.engine)
}

// UnmarshalJSON unpacks configuration into appropriate structures.
func (p *RegistrationProvider) UnmarshalJSON(inputConfig []byte) error {
	if len(inputConfig) < 10 {
		return fmt.Errorf("invalid dns provider configuration: %s", inputConfig)
	}

	data := bytes.Replace(inputConfig, []byte("\": \""), []byte("\":\""), -1)

	route53Config := []byte("\"type\":\"route53\"")
	if bytes.Contains(data, route53Config) {
		engine := &route53.RegistrationProvider{}
		if err := json.Unmarshal(data, engine); err != nil {
			return fmt.Errorf("invalid dns provider configuration, error: %s, config: %s", err, data)
		}
		p.engine = engine
		p.config = data
		return nil
	}

	return fmt.Errorf("valid dns provider not found, invalid configuration: %s", data)
}
