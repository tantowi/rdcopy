package scanner

import (
	"context"
	"github.com/mediocregopher/radix/v4"
	"redis-dumper/pkg/core/logger"
)

type service struct {
	client      radix.Client
	options     Options
	logger      logger.Service
	dumpChannel chan string
}

type Service interface {
	Start(ctx context.Context)
	GetScanChannel() <-chan string
}

func CreateService(client radix.Client, options Options, reporter logger.Service) Service {
	return &service{
		client:      client,
		options:     options,
		logger:      reporter,
		dumpChannel: make(chan string),
	}
}

func (s *service) Start(ctx context.Context) {
	go s.scanKeys(ctx)
}

func (s *service) GetScanChannel() <-chan string {
	return s.dumpChannel
}

func (s *service) scanKeys(ctx context.Context) {
	var key string
	scanConfig := radix.ScannerConfig{
		Command: "SCAN",
		Count:   s.options.RedisScanCount,
	}

	// add key pattern
	if s.options.SearchPattern != "*" {
		scanConfig.Pattern = s.options.SearchPattern
	}

	// start scanning keys
	redisScanner := scanConfig.New(s.client)
	for redisScanner.Next(ctx, &key) {
		s.logger.IncScannedCounter(1)
		s.dumpChannel <- key
	}

	close(s.dumpChannel)
}
