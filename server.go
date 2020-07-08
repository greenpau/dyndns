package dyndns

import (
	"sync"

	"go.uber.org/zap"
)

// Server represents dynamic DNS registration server.
type Server struct {
	name string
	log  *zap.Logger
	ctx  *Context
	cfg  *Config
}

// NewServer return an instance of Server.
func NewServer() *Server {
	s := &Server{
		name: "dyndns",
	}
	s.initLogger()
	s.initConfig()
	s.initContext()
	return s
}

// Run is starts the Server.
func (s *Server) Run() error {
	var wg sync.WaitGroup
	var sybsystemCount int

	go runSignalManager(s)

	// Configuration Management
	wg.Add(1)
	sybsystemCount++
	go runConfigManager(s, &wg)

	// Dynamic DNS Registration
	wg.Add(1)
	sybsystemCount++
	go runRegistrationManager(s, &wg)

	// Service Management
	wg.Add(1)
	go runServiceManager(s, &wg, sybsystemCount)

	s.log.Info(
		"started all subsystems",
		zap.String("app", s.name),
	)
	wg.Wait()
	return s.ctx.fatalError
}
