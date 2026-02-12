package event

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

const (
	SchemaVersion = "v1"

	TypePeerOnline  = "peer.online"
	TypePeerOffline = "peer.offline"

	SeverityInfo = "info"
)

type Event struct {
	SchemaVersion string         `json:"schema_version"`
	EventID       string         `json:"event_id"`
	EventType     string         `json:"event_type"`
	Severity      string         `json:"severity"`
	Timestamp     time.Time      `json:"timestamp"`
	SubjectID     string         `json:"subject_id"`
	SubjectType   string         `json:"subject_type"`
	BeforeHash    string         `json:"before_hash"`
	AfterHash     string         `json:"after_hash"`
	Payload       map[string]any `json:"payload,omitempty"`
}

func NewPresenceEvent(eventType, subjectID, beforeHash, afterHash string, payload map[string]any, now time.Time) Event {
	e := Event{
		SchemaVersion: SchemaVersion,
		EventType:     eventType,
		Severity:      SeverityInfo,
		Timestamp:     now.UTC(),
		SubjectID:     subjectID,
		SubjectType:   "peer",
		BeforeHash:    beforeHash,
		AfterHash:     afterHash,
		Payload:       payload,
	}
	e.EventID = DeriveEventID(e)
	return e
}

func DeriveEventID(e Event) string {
	msg := fmt.Sprintf("%s|%s|%s|%s|%s", e.SchemaVersion, e.EventType, e.SubjectID, e.BeforeHash, e.AfterHash)
	sum := sha256.Sum256([]byte(msg))
	return hex.EncodeToString(sum[:16])
}

func DeriveIdempotencyKey(e Event) string {
	payload, _ := json.Marshal(e.Payload)
	msg := fmt.Sprintf("%s|%s|%s|%s|%s|%s",
		e.EventType,
		e.SubjectID,
		e.BeforeHash,
		e.AfterHash,
		payload,
		e.Timestamp.UTC().Format(time.RFC3339Nano),
	)
	sum := sha256.Sum256([]byte(msg))
	return hex.EncodeToString(sum[:])
}
