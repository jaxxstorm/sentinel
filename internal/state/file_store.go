package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jaxxstorm/sentinel/internal/snapshot"
)

type fileData struct {
	Snapshot        *snapshot.Snapshot   `json:"snapshot,omitempty"`
	IdempotencyKeys map[string]time.Time `json:"idempotency_keys,omitempty"`
}

type FileStore struct {
	path string
	now  func() time.Time
}

func NewFileStore(path string) *FileStore {
	return &FileStore{path: path, now: time.Now}
}

func (s *FileStore) LoadSnapshot() (snapshot.Snapshot, error) {
	data, err := s.read()
	if err != nil {
		return snapshot.Snapshot{}, err
	}
	if data.Snapshot == nil {
		return snapshot.Snapshot{}, ErrNoSnapshot
	}
	return *data.Snapshot, nil
}

func (s *FileStore) SaveSnapshot(in snapshot.Snapshot) error {
	data, err := s.read()
	if err != nil && !errors.Is(err, os.ErrNotExist) && !errors.Is(err, ErrNoSnapshot) {
		return err
	}
	if data.IdempotencyKeys == nil {
		data.IdempotencyKeys = map[string]time.Time{}
	}
	data.Snapshot = &in
	return s.write(data)
}

func (s *FileStore) SeenIdempotencyKey(key string) (bool, error) {
	data, err := s.read()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	now := s.now().UTC()
	for k, exp := range data.IdempotencyKeys {
		if exp.Before(now) {
			delete(data.IdempotencyKeys, k)
		}
	}
	exp, ok := data.IdempotencyKeys[key]
	if !ok || exp.Before(now) {
		_ = s.write(data)
		return false, nil
	}
	return true, nil
}

func (s *FileStore) RecordIdempotencyKey(key string, ttl time.Duration) error {
	data, err := s.read()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if data.IdempotencyKeys == nil {
		data.IdempotencyKeys = map[string]time.Time{}
	}
	data.IdempotencyKeys[key] = s.now().UTC().Add(ttl)
	return s.write(data)
}

func (s *FileStore) read() (fileData, error) {
	b, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fileData{}, os.ErrNotExist
		}
		return fileData{}, err
	}
	if len(b) == 0 {
		return fileData{}, nil
	}
	var data fileData
	if err := json.Unmarshal(b, &data); err != nil {
		corrupt := fmt.Sprintf("%s.corrupt-%d", s.path, s.now().Unix())
		_ = os.Rename(s.path, corrupt)
		return fileData{}, nil
	}
	if data.IdempotencyKeys == nil {
		data.IdempotencyKeys = map[string]time.Time{}
	}
	return data, nil
}

func (s *FileStore) write(data fileData) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}
