package dyndns

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"
)

func runServiceManager(s *Server, wg *sync.WaitGroup, goRoutineCount int) {
	defer wg.Done()
	var fn = s.name + "-service-mgr"
	s.log.Debug(
		"starting sybsystem",
		zap.String("subsystem", fn),
		zap.String("app", s.name),
		zap.Any("config", s.GetConfig()),
	)

M:
	for {
		select {
		case err := <-s.ctx.error:
			if err != nil {
				s.ctx.fatalError = err
				s.log.Error(
					"notifying goroutines in response to subsystem error",
					zap.String("app", s.name),
					zap.String("error", err.Error()),
				)
				for i := 0; i < goRoutineCount; i++ {
					s.ctx.exitRoutine <- true
				}
			} else {
				goRoutineCount--
			}
		case reloadNotice := <-s.ctx.reload:
			s.log.Info(
				"notifying goroutines in response to reload signal",
				zap.String("app", s.name),
			)
			for i := 0; i < goRoutineCount; i++ {
				s.ctx.reloadRoutine <- reloadNotice
			}
		case exitNotice := <-s.ctx.exit:
			s.log.Debug(
				"notifying goroutines in response to graceful shutdown signal",
				zap.String("app", s.name),
			)
			for i := 0; i < goRoutineCount; i++ {
				s.ctx.exitRoutine <- exitNotice
				s.log.Debug(
					"notified goroutine in response to graceful shutdown signal",
					zap.String("app", s.name),
					zap.Int("routine_id", i),
				)

			}
		default:
			if goRoutineCount == 0 {
				s.log.Info(
					"all subsystems exited successfully",
					zap.String("app", s.name),
				)
				break M
			}
			time.Sleep(500 * time.Millisecond)
		}
	}
	return
}

func runSignalManager(s *Server) {
	sysChannel := make(chan os.Signal, 1)
	signal.Notify(sysChannel, os.Interrupt, syscall.SIGTERM)
	var fn = s.name + "-signal-mgr"
	s.log.Debug(
		"starting sybsystem",
		zap.String("subsystem", fn),
		zap.String("app", s.name),
		zap.Any("config", s.GetConfig()),
	)

	hardStopSeconds := time.Duration(10)
	signalID := <-sysChannel
	// TODO: implement reload
	s.log.Debug(
		"shutting down all subsystems in response to the received system",
		zap.String("app", s.name),
		zap.Duration("timeout", hardStopSeconds),
		zap.String("signal_name", signalID.String()),
		zap.Any("signal_id", signalID),
	)
	s.ctx.exit <- true
	time.Sleep(time.Second * hardStopSeconds)
	s.log.Warn(
		"some subsystems did not shutdown gracefully",
		zap.String("app", s.name),
	)
	os.Exit(1)

}
