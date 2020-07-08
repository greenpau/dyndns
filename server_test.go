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

	t.Logf("running configuration: %v", server.GetConfig())

	if err := server.ValidateConfig(); err != nil {
		t.Fatalf("error validating configuration file: %s", err)
	}

	t.Logf("configuration is valid")
}
