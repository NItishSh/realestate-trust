package core

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

type AuditEntry struct {
	Index        int64     `json:"index"`
	Timestamp    time.Time `json:"timestamp"`
	Payload      string    `json:"payload"`
	PreviousHash string    `json:"previousHash"`
	Hash         string    `json:"hash"`
	EventID      string    `json:"eventId,omitempty"`
}

// CalculateHash generates SHA256 code mapping payload content deterministically.
func (e *AuditEntry) CalculateHash() string {
	canonicalTime := e.Timestamp.UTC().Truncate(time.Microsecond).Format(time.RFC3339Nano)
	record := fmt.Sprintf("%d%s%s%s%s", e.Index, canonicalTime, e.Payload, e.PreviousHash, e.EventID)
	h := sha256.New()
	h.Write([]byte(record))
	return hex.EncodeToString(h.Sum(nil))
}
