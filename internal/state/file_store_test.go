package state

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jaxxstorm/sentinel/internal/snapshot"
)

func TestFileStoreAtomicSnapshotAndKeyPersistence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	store := NewFileStore(path)

	snap := snapshot.Snapshot{Hash: "hash1", Peers: []snapshot.Peer{{ID: "peer1", Online: true}}}
	if err := store.SaveSnapshot(snap); err != nil {
		t.Fatal(err)
	}
	loaded, err := store.LoadSnapshot()
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Hash != "hash1" {
		t.Fatalf("expected hash1, got %s", loaded.Hash)
	}

	if err := store.RecordIdempotencyKey("k1", time.Hour); err != nil {
		t.Fatal(err)
	}
	seen, err := store.SeenIdempotencyKey("k1")
	if err != nil {
		t.Fatal(err)
	}
	if !seen {
		t.Fatal("expected idempotency key to be seen")
	}
}

func TestFileStoreRecoversFromCorruption(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")
	if err := os.WriteFile(path, []byte("{invalid"), 0o600); err != nil {
		t.Fatal(err)
	}
	store := NewFileStore(path)
	if _, err := store.LoadSnapshot(); err == nil {
		t.Fatal("expected no snapshot error")
	}
	matches, err := filepath.Glob(path + ".corrupt-*")
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) == 0 {
		t.Fatal("expected corrupt backup file to be created")
	}
}
