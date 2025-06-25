package restore

import (
	"context"
	"fmt"
	"sync"

	"rdcopy/pkg/core/logger"

	"github.com/redis/go-redis/v9"
)

type service struct {
	client              *redis.Client
	logger              logger.Service
	dumpChannel         <-chan Entry
	replaceExistingKeys bool
}

type Service interface {
	Start(ctx context.Context, wg *sync.WaitGroup, number int)
}

func CreateService(client *redis.Client, dumpChannel <-chan Entry, reporter logger.Service, replaceExistingKeys bool) Service {
	return &service{
		client:              client,
		logger:              reporter,
		dumpChannel:         dumpChannel,
		replaceExistingKeys: replaceExistingKeys,
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
		// restore key and replace if still exists
		var err error
		if p.replaceExistingKeys {
			err = p.client.Do(ctx, "RESTORE", dump.Key, dump.Ttl, dump.Value, "REPLACE").Err()
		} else {
			err = p.client.Do(ctx, "RESTORE", dump.Key, dump.Ttl, dump.Value).Err()
		}

		if err != nil {
			fmt.Println(fmt.Errorf("could not restore entry: %w", err))
			continue
		}
		p.logger.IncRestoredCounter(1)
	}

	wg.Done()
}
