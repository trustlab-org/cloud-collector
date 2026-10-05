package envelope

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
)

// Envelope matches the rsyslog/omkafka shape the engine consumes, with structured data in JSON instead of Raw.
type Envelope struct {
	Meta Meta            `json:"meta"`
	JSON json.RawMessage `json:"json"`
}

type Meta struct {
	Org           string `json:"org"`
	Server        string `json:"server"`
	SourceProgram string `json:"source_program"`
	EventID       string `json:"event_id"`
}

func (e Envelope) Marshal() ([]byte, error) {
	return json.Marshal(e)
}

// ID derives a stable event id from the source's natural key parts.
func ID(parts ...string) string {
	sum := sha256.Sum256([]byte(strings.Join(parts, "\x1f")))
	return hex.EncodeToString(sum[:16])
}
