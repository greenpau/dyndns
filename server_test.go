package dyndns

import (
	"testing"
)

func TestServer(t *testing.T) {
	configFile := "./assets/conf/config.json"
	server := NewServer()

	if err := server.LoadConfig(configFile); err != nil {
		t.Fatalf("error reading configuration file: %s", err)
	}

	cfg := server.GetConfig()
	t.Logf("running configuration: %v", cfg)

	t.Logf("dns record: %v", cfg.Record)
	t.Logf("provider config: %s", cfg.Provider.config)

	if err := server.ValidateConfig(); err != nil {
		t.Fatalf("error validating configuration file: %s", err)
	}

	t.Logf("configuration is valid")
}
