package cache

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestCache_SetAndIsSet(t *testing.T) {
	logger := zap.NewNop()
	cache := New(context.Background(), logger)

	id := uuid.New()

	tests := []struct {
		name     string
		eventID  uuid.UUID
		expected bool
	}{
		{"Not Set UUID", uuid.New(), false},
		{"Set UUID", id, true},
	}

	cache.Set(id)

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			isSet := cache.IsSet(tt.eventID)
			require.Equal(t, tt.expected, isSet)
		})
	}
}

func TestCache_wakeUp(t *testing.T) {
	logger := zap.NewNop()
	cache := New(context.Background(), logger)

	err := cache.wakeUp()
	require.NoError(t, err)
}

func TestCache_toRedis(t *testing.T) {
	logger := zap.NewNop()
	cache := New(context.Background(), logger)

	err := cache.toRedis()
	require.NoError(t, err)
}

func TestCache_BackupWorker_Cancel(t *testing.T) {
	logger := zap.NewNop()
	ctx, cancel := context.WithCancel(context.Background())
	cache := New(ctx, logger)

	done := make(chan struct{})
	go func() {
		cache.BackupWorker(ctx)
		close(done)
	}()

	cancel()
	select {
	case <-done:
	case <-time.After(time.Second * 2):
		t.Fatal("backup worker did not stop on context cancel")
	}
}
