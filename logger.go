package dyndns

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func newLogEncoderConfig() zapcore.EncoderConfig {
	cfg := zap.NewProductionEncoderConfig()
	cfg.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.TimeKey = "time"
	return cfg
}

func newLogger(logAtom zap.AtomicLevel) *zap.Logger {
	logEncoderConfig := newLogEncoderConfig()
	logger := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(logEncoderConfig),
		zapcore.Lock(os.Stdout),
		logAtom,
	))
	return logger

}

func (s *Server) initLogger() {
	if s.log != nil {
		return
	}

	logAtom := zap.NewAtomicLevel()
	s.log = newLogger(logAtom)
	// TODO: what to do with the below?
	// defer s.log.Sync()
}

// GetLogger returns the instance of logger for a server.
func (s *Server) GetLogger() *zap.Logger {
	return s.log
}

// SetLogLevel sets the server logging level.
func (s *Server) SetLogLevel(logLevel string) error {

	logAtom := zap.NewAtomicLevel()

	switch logLevel {
	case "info":
		logAtom.SetLevel(zapcore.InfoLevel)
	case "warn":
		logAtom.SetLevel(zapcore.WarnLevel)
	case "debug":
		logAtom.SetLevel(zapcore.DebugLevel)
	case "error":
		logAtom.SetLevel(zapcore.ErrorLevel)
	case "fatal":
		logAtom.SetLevel(zapcore.FatalLevel)
	default:
		return fmt.Errorf("unsupported log level %s", logLevel)
	}

	s.log = newLogger(logAtom)
	s.cfg.LogLevel = logLevel

	return nil
}
