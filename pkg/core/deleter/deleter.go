package deleter

import (
	"context"
	"fmt"
	"sync"

	"github.com/appit-online/redis-dumper/pkg/core/logger"
	"github.com/redis/go-redis/v9"
)

type service struct {
	client        *redis.Client
	logger        logger.Service
	deleteChannel <-chan string
}

type Service interface {
	Start(ctx context.Context, deleteRoutineCount int)
}

func CreateService(client *redis.Client, deleteChannel <-chan string, logger logger.Service) Service {
	return &service{
		client:        client,
		logger:        logger,
		deleteChannel: deleteChannel,
	}
}

func (s *service) Start(ctx context.Context, deleteRoutineCount int) {
	wgPull := new(sync.WaitGroup)
	wgPull.Add(deleteRoutineCount)

	// parallelize deleting of redis key
	for i := 0; i < deleteRoutineCount; i++ {
		go s.delete(ctx, wgPull)
	}

	wgPull.Wait()
}

func (s *service) delete(ctx context.Context, wg *sync.WaitGroup) {
	for key := range s.deleteChannel {
		if err := s.client.Del(ctx, key).Err(); err != nil {
			fmt.Println(fmt.Errorf("could not delete entry: %w", err))
		}

		s.logger.IncDeletedCounter(1)
	}

	wg.Done()
}
