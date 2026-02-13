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

	SubjectPeer    = "peer"
	SubjectDaemon  = "daemon"
	SubjectPrefs   = "prefs"
	SubjectTailnet = "tailnet"

	TypePeerOnline  = "peer.online"
	TypePeerOffline = "peer.offline"
	TypePeerAdded   = "peer.added"
	TypePeerRemoved = "peer.removed"

	TypePeerRoutesChanged            = "peer.routes.changed"
	TypePeerTagsChanged              = "peer.tags.changed"
	TypePeerMachineAuthorizedChanged = "peer.machine_authorized.changed"
	TypePeerKeyExpiryChanged         = "peer.key_expiry.changed"
	TypePeerKeyExpired               = "peer.key_expired"
	TypePeerHostinfoChanged          = "peer.hostinfo.changed"

	TypeDaemonStateChanged = "daemon.state.changed"

	TypePrefsAdvertiseRoutesChanged = "prefs.advertise_routes.changed"
	TypePrefsExitNodeChanged        = "prefs.exit_node.changed"
	TypePrefsRunSSHChanged          = "prefs.run_ssh.changed"
	TypePrefsShieldsUpChanged       = "prefs.shields_up.changed"

	TypeTailnetDomainChanged     = "tailnet.domain.changed"
	TypeTailnetTKAEnabledChanged = "tailnet.tka_enabled.changed"

	SeverityInfo = "info"
)

var knownEventTypes = map[string]struct{}{
	TypePeerOnline:  {},
	TypePeerOffline: {},
	TypePeerAdded:   {},
	TypePeerRemoved: {},

	TypePeerRoutesChanged:            {},
	TypePeerTagsChanged:              {},
	TypePeerMachineAuthorizedChanged: {},
	TypePeerKeyExpiryChanged:         {},
	TypePeerKeyExpired:               {},
	TypePeerHostinfoChanged:          {},

	TypeDaemonStateChanged: {},

	TypePrefsAdvertiseRoutesChanged: {},
	TypePrefsExitNodeChanged:        {},
	TypePrefsRunSSHChanged:          {},
	TypePrefsShieldsUpChanged:       {},

	TypeTailnetDomainChanged:     {},
	TypeTailnetTKAEnabledChanged: {},
}

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

func IsKnownType(eventType string) bool {
	_, ok := knownEventTypes[eventType]
	return ok
}

func NewEvent(eventType, subjectType, subjectID, beforeHash, afterHash string, payload map[string]any, now time.Time) Event {
	e := Event{
		SchemaVersion: SchemaVersion,
		EventType:     eventType,
		Severity:      SeverityInfo,
		Timestamp:     now.UTC(),
		SubjectID:     subjectID,
		SubjectType:   subjectType,
		BeforeHash:    beforeHash,
		AfterHash:     afterHash,
		Payload:       payload,
	}
	e.EventID = DeriveEventID(e)
	return e
}

func NewPeerEvent(eventType, subjectID, beforeHash, afterHash string, payload map[string]any, now time.Time) Event {
	return NewEvent(eventType, SubjectPeer, subjectID, beforeHash, afterHash, payload, now)
}

func NewDaemonEvent(eventType, subjectID, beforeHash, afterHash string, payload map[string]any, now time.Time) Event {
	return NewEvent(eventType, SubjectDaemon, subjectID, beforeHash, afterHash, payload, now)
}

func NewPrefsEvent(eventType, subjectID, beforeHash, afterHash string, payload map[string]any, now time.Time) Event {
	return NewEvent(eventType, SubjectPrefs, subjectID, beforeHash, afterHash, payload, now)
}

func NewTailnetEvent(eventType, subjectID, beforeHash, afterHash string, payload map[string]any, now time.Time) Event {
	return NewEvent(eventType, SubjectTailnet, subjectID, beforeHash, afterHash, payload, now)
}

func NewPresenceEvent(eventType, subjectID, beforeHash, afterHash string, payload map[string]any, now time.Time) Event {
	return NewPeerEvent(eventType, subjectID, beforeHash, afterHash, payload, now)
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
