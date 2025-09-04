package cache

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const defaultUpdateTime = time.Minute * 10

type Cache struct {
	// sync.RWMutex because we assume that there will be more readers.
	// If we have a very big numbers of concurent readers better to use sync.Map instead of sync.RWMutex.
	// sync.Map also solving the problem of atomic increment of readers between big amount of cores.
	mu  sync.RWMutex
	log *zap.Logger
	// I used "set" as a datastructure just for the current performance.
	// Guess we need to keep whole event in hashmap.
	events       map[uuid.UUID]struct{}
	backupTicker *time.Ticker
}

func New(ctx context.Context, log *zap.Logger) *Cache {
	c := &Cache{
		backupTicker: time.NewTicker(defaultUpdateTime),
		log:          log,
		events:       make(map[uuid.UUID]struct{}),
	}

	_ = c.wakeUp()
	// c.backupWorker()

	return c
}

func (c *Cache) Set(eventID uuid.UUID) {
	c.mu.Lock()
	c.events[eventID] = struct{}{}
	c.mu.Unlock()
}

// IsSet - Over time, our table will grow therefore we will need to scale it.
// We will use sharding (horizontal). And for searching on different shards,
// we will use the "Distributed Search Function" and gob format for cooperation(I can explain in detail)
func (c *Cache) IsSet(id uuid.UUID) bool {
	c.mu.RLock()
	_, ok := c.events[id]
	c.mu.RUnlock()
	return ok
}

// wakeUp - In case of crush of our app we are able to restore Events data
func (c *Cache) wakeUp() error {
	// getting stored data from Redis
	// ...
	// c.events = dataFromRedis

	return nil
}

func (c *Cache) BackupWorker(ctx context.Context) {
	c.log.Info("starting cache backup worker ")

	defer func() {
		c.backupTicker.Stop()
		c.log.Info("cache backup worker gracefully stopped")
	}()

	for {
		select {
		case <-c.backupTicker.C:
			_ = c.toRedis()
		case <-ctx.Done():
			return
		}
	}
}

func (c *Cache) toRedis() error {
	// write c.events data to redis
	// ...

	// Also we need to have a constant storage beside Redis
	// so we can have a worker that taking from redis and store to Postgres.
	// For example using "replication" master node can save into Postgres and redis, then update,
	// replicas can only get from redis.

	return nil
}
