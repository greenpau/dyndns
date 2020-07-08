package dyndns

import (
	//	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

func runRegistrationManager(s *Server, parentWaitGroup *sync.WaitGroup) {
	defer parentWaitGroup.Done()
	var exitRoutine bool

	var fn = s.name + "-dyndns-mgr"
	s.log.Debug(
		"starting sybsystem",
		zap.String("subsystem", fn),
		zap.String("app", s.name),
	)

	intervals := time.NewTicker(time.Millisecond * time.Duration(250))
	for range intervals.C {
		// Add configuration reload code
		if exitRoutine {
			break
		}
		select {
		case _ = <-s.ctx.exitRoutine:
			s.log.Debug(
				"shutting down subsystem",
				zap.String("subsystem", fn),
				zap.String("app", s.name),
			)
			exitRoutine = true
			s.log.Debug(
				"stopped dyndns services",
				zap.String("subsystem", fn),
				zap.String("app", s.name),
			)
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
