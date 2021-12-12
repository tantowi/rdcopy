package restore

import (
	"context"
	"log"
	"sync"

	"github.com/appit-online/redis-dumper/pkg/core/logger"
	"github.com/mediocregopher/radix/v4"
)

type service struct {
	client      radix.Client
	logger      logger.Service
	dumpChannel <-chan Entry
}

type Service interface {
	Start(ctx context.Context, wg *sync.WaitGroup, number int)
}

func CreateService(client radix.Client, dumpChannel <-chan Entry, reporter logger.Service) Service {
	return &service{
		client:      client,
		logger:      reporter,
		dumpChannel: dumpChannel,
	}
}

func (p *service) Start(ctx context.Context, wg *sync.WaitGroup, number int) {
	wg.Add(number)
	for i := 0; i < number; i++ {
		go p.restore(ctx, wg)
	}
}

func (p *service) restore(ctx context.Context, wg *sync.WaitGroup) {
	for dump := range p.dumpChannel {
		p.logger.IncRestoredCounter(1)

		// restore key and replace if still exists
		err := p.client.Do(ctx, radix.FlatCmd(nil, "RESTORE", dump.Key, dump.Ttl, dump.Value, "REPLACE"))
		if err != nil {
			log.Fatal(err)
		}
	}

	wg.Done()
}
