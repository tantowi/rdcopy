package dumper

import (
	"context"
	"fmt"
	"github.com/appit-online/redis-dumper/pkg/core/logger"
	"github.com/appit-online/redis-dumper/pkg/core/restore"
	"github.com/redis/go-redis/v9"
	"sync"
)

type service struct {
	client         *redis.Client
	logger         logger.Service
	dumpChannel    <-chan string
	restoreChannel chan restore.Entry
}

type Service interface {
	Start(ctx context.Context, dumperRoutineCount int)
	GetRestorerChannel() <-chan restore.Entry
}

func CreateService(client *redis.Client, dumpChannel <-chan string, reporter logger.Service, parallelRestores int) Service {
	return &service{
		client:         client,
		logger:         reporter,
		restoreChannel: make(chan restore.Entry, parallelRestores),
		dumpChannel:    dumpChannel,
	}
}

func (s *service) Start(ctx context.Context, dumperRoutineCount int) {
	wgPull := new(sync.WaitGroup)
	wgPull.Add(dumperRoutineCount)

	// parallelize dumping of values and ttl by redis key
	for i := 0; i < dumperRoutineCount; i++ {
		go s.dumpValueRoutine(ctx, wgPull)
	}

	wgPull.Wait()
	close(s.restoreChannel)
}

func (s *service) GetRestorerChannel() <-chan restore.Entry {
	return s.restoreChannel
}

func (s *service) dumpValueRoutine(ctx context.Context, wg *sync.WaitGroup) {
	for key := range s.dumpChannel {
		// dump ttl and value
		pipe := s.client.Pipeline()
		ttl := pipe.Do(ctx, "PTTL", key)
		value := pipe.Do(ctx, "DUMP", key)

		_, err := pipe.Exec(ctx)
		if err != nil {
			fmt.Println(fmt.Errorf("could not dump entry: %w", err))
			continue
		}

		convertedTtl, ok := ttl.Val().(int64)
		if !ok {
			convertedTtl = 1000
		}

		// add value and ttl to channel
		s.restoreChannel <- restore.Entry{
			Key:   key,
			Ttl:   int(convertedTtl),
			Value: value.Val().(string),
		}
		s.logger.IncDumpedCounter(1)
	}

	wg.Done()
}
