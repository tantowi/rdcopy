package scanner

import (
	"context"
	"log"

	"rdcopy/pkg/core/logger"

	"github.com/redis/go-redis/v9"
)

type service struct {
	client      *redis.Client
	options     Options
	logger      logger.Service
	dumpChannel chan string
}

type Service interface {
	Start(ctx context.Context)
	GetDumperChannel() <-chan string
}

func CreateService(client *redis.Client, options Options, reporter logger.Service) Service {
	return &service{
		client:      client,
		options:     options,
		logger:      reporter,
		dumpChannel: make(chan string, options.ParallelDumps),
	}
}

func (s *service) Start(ctx context.Context) {
	go s.scanKeys(ctx)
}

func (s *service) GetDumperChannel() <-chan string {
	return s.dumpChannel
}

func (s *service) scanKeys(ctx context.Context) {
	// start scanning keys
	var cursor uint64
	for {
		var keys []string
		var err error
		keys, cursor, err = s.client.Scan(ctx, cursor, s.options.SearchPattern, int64(s.options.RedisScanCount)).Result()
		if err != nil {
			log.Fatal(err)
		}

		// write keys to dump channel
		for _, key := range keys {
			s.dumpChannel <- key
			s.logger.IncScannedCounter(1)
		}

		if cursor == 0 {
			break
		}
	}

	close(s.dumpChannel)
}
