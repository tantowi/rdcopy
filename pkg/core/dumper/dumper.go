package dumper

import (
	"context"
	"log"
	"sync"

	"github.com/appit-online/redis-dumper/pkg/core/logger"
	"github.com/appit-online/redis-dumper/pkg/core/restore"
	"github.com/mediocregopher/radix/v4"
)

type service struct {
	client         radix.Client
	logger         logger.Service
	dumpChannel    <-chan string
	restoreChannel chan restore.Entry
}

type Service interface {
	Start(ctx context.Context, dumperRoutineCount int)
	GetDumpChannel() <-chan restore.Entry
}

func CreateService(client radix.Client, dumpChannel <-chan string, reporter logger.Service) Service {
	return &service{
		client:         client,
		logger:         reporter,
		restoreChannel: make(chan restore.Entry),
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

func (s *service) GetDumpChannel() <-chan restore.Entry {
	return s.restoreChannel
}

func (s *service) dumpValueRoutine(ctx context.Context, wg *sync.WaitGroup) {
	for key := range s.dumpChannel {
		var value string
		var ttl int

		// dump ttl and value
		p := radix.NewPipeline()
		p.Append(radix.Cmd(&ttl, "PTTL", key))
		p.Append(radix.Cmd(&value, "DUMP", key))

		if err := s.client.Do(ctx, p); err != nil {
			log.Fatal(err)
		}

		if ttl < 0 {
			ttl = 0
		}

		// and value and ttl to channel
		s.logger.IncDumpedCounter(1)
		s.restoreChannel <- restore.Entry{
			Key:   key,
			Ttl:   ttl,
			Value: value,
		}
	}

	wg.Done()
}
