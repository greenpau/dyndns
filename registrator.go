package dyndns

import (
	//	"fmt"
	"github.com/greenpau/dyndns/pkg/utils"
	"sync"
	"time"

	"go.uber.org/zap"
)

func runRegistrationManager(s *Server, parentWaitGroup *sync.WaitGroup) {
	defer parentWaitGroup.Done()
	var exitRoutine bool

	var fn = s.name + "-registration-mgr"
	syncInterval := float64(s.cfg.SyncInterval)
	record := s.cfg.Record
	provider := s.cfg.Provider
	s.log.Debug(
		"starting sybsystem",
		zap.String("subsystem", fn),
		zap.String("app", s.name),
		zap.Any("sync_interval", syncInterval),
		zap.Any("record", s.cfg.Record),
	)

	initialized := false
	timer := time.Now()
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
			elapsed := time.Since(timer)
			if elapsed.Seconds() < syncInterval {
				if initialized {
					continue
				}
				initialized = true
			}
			timer = time.Now()
			if record.Version4 {
				// Get the public IP address of the host running this service
				s.log.Debug(
					"checking public ip address",
					zap.String("subsystem", fn),
					zap.String("app", s.name),
				)

				addr, err := utils.GetPublicAddress(4)
				if err != nil {
					s.log.Error(
						"checking public ip address failed",
						zap.String("subsystem", fn),
						zap.String("app", s.name),
						zap.String("error", err.Error()),
					)
					continue
				}
				s.log.Debug(
					"obtained public ip address",
					zap.String("subsystem", fn),
					zap.String("app", s.name),
					zap.Any("address", addr),
				)

				// Resolve the IP address associated with DNS A/AAAA record

				s.log.Debug(
					"resolving dns record",
					zap.String("subsystem", fn),
					zap.String("app", s.name),
					zap.Any("record", record),
				)
				dnsAddrs, err := utils.ResolveName(record.Name, 4)
				if err != nil {
					s.log.Error(
						"resolving dns record failed",
						zap.String("subsystem", fn),
						zap.String("app", s.name),
						zap.Any("record", record),
						zap.String("error", err.Error()),
					)
					continue
				}
				s.log.Debug(
					"resolved dns record",
					zap.String("subsystem", fn),
					zap.String("app", s.name),
					zap.Any("record", record),
					zap.Any("addresses", dnsAddrs),
				)

				if len(dnsAddrs) == 1 {
					if dnsAddrs[0] == addr {
						s.log.Debug(
							"dns record is up to date",
							zap.String("subsystem", fn),
							zap.String("app", s.name),
							zap.Any("record", record),
							zap.Any("public_ip", addr),
							zap.Any("dns_addresses", dnsAddrs),
						)
						// TODO: remove comment
						//continue
					}
				}

				// Update DNS Zone
				s.log.Info(
					"dns record is not up to date",
					zap.String("subsystem", fn),
					zap.String("app", s.name),
					zap.Any("record", record),
					zap.Any("public_ip", addr),
					zap.Any("dns_addresses", dnsAddrs),
				)

				if err := record.SetAddress(addr, 4); err != nil {
					s.log.Error(
						"failed updating internal dns record",
						zap.String("subsystem", fn),
						zap.String("app", s.name),
						zap.Any("record", record),
						zap.String("error", err.Error()),
					)
					continue
				}

				if err := provider.Register(record); err != nil {
					s.log.Error(
						"dns record update failed",
						zap.String("subsystem", fn),
						zap.String("app", s.name),
						zap.Any("record", record),
						zap.Any("public_ip", addr),
						zap.Any("dns_addresses", dnsAddrs),
						zap.String("error", err.Error()),
					)
					continue
				}

				s.log.Info(
					"updated dns record",
					zap.String("subsystem", fn),
					zap.String("app", s.name),
					zap.Any("record", record),
					zap.Any("public_ip", addr),
					zap.Any("dns_addresses", dnsAddrs),
				)
			}

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
