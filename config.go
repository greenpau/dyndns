package dyndns

import (
	"encoding/json"
	"fmt"
	"github.com/greenpau/dyndns/pkg/record"
	"gopkg.in/yaml.v2"
	"os"
	"path/filepath"
	"sync"
	"time"

	"strings"

	"go.uber.org/zap"
)

// Config is the configuration of the Server.
type Config struct {
	sync.Mutex
	name         string
	Provider     *RegistrationProvider      `json:"provider" yaml:"provider"`
	Record       *record.RegistrationRecord `json:"record" yaml:"record"`
	SyncInterval uint64                     `json:"sync_interval" yaml:"sync_interval"`
	LogLevel     string                     `json:"log_level" yaml:"log_level"`
	File         string                     `json:"conf_file" yaml:"conf_file"`
}

// LoadConfig loads configuration of the Server from a file.
func (s *Server) LoadConfig(configFile string) error {
	var configType string
	configDir, configFileName := filepath.Split(configFile)
	ext := filepath.Ext(configFileName)
	switch ext {
	case ".json":
		configType = "json"
	default:
		configType = "yaml"
	}
	configName := strings.TrimSuffix(configFile, ext)

	s.log.Info(
		"loading configuration file",
		zap.String("file_path", configFile),
		zap.String("file_dir", configDir),
		zap.String("file_basename", configName),
		zap.String("file_name", configFileName),
		zap.String("file_type", configType),
	)

	configFileHandler, err := os.Open(configFile)
	if err != nil {
		return err
	}
	defer configFileHandler.Close()

	switch configType {
	case "yaml":
		decoder := yaml.NewDecoder(configFileHandler)
		if err := decoder.Decode(s.cfg); err != nil {
			return err
		}
	case "json":
		decoder := json.NewDecoder(configFileHandler)
		if err := decoder.Decode(s.cfg); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported configuration type: %s", configType)
	}

	s.cfg.File = configFile

	if s.cfg.LogLevel != "" {
		if err := s.SetLogLevel(s.cfg.LogLevel); err != nil {
			return err
		}
	}

	if s.cfg.SyncInterval == 0 {
		s.cfg.SyncInterval = 60
	}

	if s.cfg.Provider == nil {
		return fmt.Errorf("dns provider failed to initialize due to invalid configuration")
	}

	if err := s.cfg.Provider.Validate(); err != nil {
		return fmt.Errorf("%s: invalid dns provider definition, error: %s", s.name, err.Error())
	}

	if s.cfg.Record == nil {
		return fmt.Errorf("dns record failed to initialize due to invalid configuration")
	}

	if err := s.cfg.Record.Validate(); err != nil {
		return fmt.Errorf("%s: invalid dns record definition, error: %s", s.name, err.Error())
	}

	return nil
}

// ValidateConfig validates configuration.
func (s *Server) ValidateConfig() error {
	if s.cfg.File == "" {
		return fmt.Errorf("%s: configuration file not provided", s.name)
	}
	if s.cfg.SyncInterval == 0 {
		return fmt.Errorf("%s: sync interval is null", s.name)
	}

	if s.cfg.Provider == nil {
		return fmt.Errorf("%s: dns provider failed to initialize due to invalid configuration", s.name)
	}

	if err := s.cfg.Provider.Validate(); err != nil {
		return fmt.Errorf("%s: invalid dns provider definition, error: %s", s.name, err.Error())
	}

	if err := s.cfg.Provider.Configure(s.log); err != nil {
		return fmt.Errorf("%s: dns provider configuration error: %s", s.name, err.Error())
	}

	if s.cfg.Record == nil {
		return fmt.Errorf("%s: dns record failed to initialize due to invalid configuration", s.name)
	}

	if err := s.cfg.Record.Validate(); err != nil {
		return fmt.Errorf("%s: invalid dns record definition, error: %s", s.name, err.Error())
	}

	return nil
}

// GetConfig returns an instance of Config.
func (s *Server) GetConfig() *Config {
	return s.cfg
}

func (s *Server) initConfig() {
	if s.cfg != nil {
		return
	}
	cfg := &Config{
		name: s.name,
	}
	s.cfg = cfg
	return
}

func runConfigManager(s *Server, wg *sync.WaitGroup) {
	defer wg.Done()
	var fn = s.name + "-config-mgr"
	s.log.Debug(
		"starting sybsystem",
		zap.String("subsystem", fn),
		zap.String("app", s.name),
		zap.Any("config", s.GetConfig()),
	)

	var count uint64
	var exitRoutine bool
	intervals := time.NewTicker(time.Millisecond * time.Duration(250))
	for range intervals.C {
		// Add configuration reload code
		if exitRoutine {
			break
		}
		count++
		select {
		case _ = <-s.ctx.exitRoutine:
			s.log.Debug(
				"shutting down subsystem",
				zap.String("subsystem", fn),
				zap.String("app", s.name),
			)
			exitRoutine = true
		default:
			// do nothing
		}
	}
	intervals.Stop()
	s.log.Debug(
		"stopped subsystem",
		zap.String("subsystem", fn),
		zap.String("app", s.name),
	)
	s.ctx.error <- nil
	return
}
