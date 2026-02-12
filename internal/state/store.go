package state

import (
	"errors"
	"time"

	"github.com/jaxxstorm/sentinel/internal/snapshot"
)

type StateStore interface {
	LoadSnapshot() (snapshot.Snapshot, error)
	SaveSnapshot(snapshot.Snapshot) error
	SeenIdempotencyKey(key string) (bool, error)
	RecordIdempotencyKey(key string, ttl time.Duration) error
}

var ErrNoSnapshot = errors.New("no snapshot")
