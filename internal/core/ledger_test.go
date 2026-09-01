package core

import (
	"testing"
	"time"
)

func TestCalculateHash_Determinism(t *testing.T) {
	fixedTime := time.Date(2026, 9, 1, 12, 0, 0, 123456789, time.UTC)

	entry1 := &AuditEntry{
		Index:        1,
		Timestamp:    fixedTime,
		Payload:      "test-payload",
		PreviousHash: "prev-hash-000",
		EventID:      "event-123",
	}

	hash1 := entry1.CalculateHash()
	if hash1 == "" {
		t.Fatalf("expected non-empty hash")
	}

	// Recompute on new object with same data
	entry2 := &AuditEntry{
		Index:        1,
		Timestamp:    fixedTime,
		Payload:      "test-payload",
		PreviousHash: "prev-hash-000",
		EventID:      "event-123",
	}

	hash2 := entry2.CalculateHash()
	if hash1 != hash2 {
		t.Fatalf("expected hashes to match, got %s != %s", hash1, hash2)
	}

	// Simulate time without monotonic clock vs with monotonic clock (time.Now())
	now := time.Now()
	entryWithMonotonic := &AuditEntry{
		Index:        2,
		Timestamp:    now,
		Payload:      "monotonic-test",
		PreviousHash: hash1,
		EventID:      "event-456",
	}
	hashWithMono := entryWithMonotonic.CalculateHash()

	// Strip monotonic clock reading as happens after DB persistence (round-trip)
	entryStripped := &AuditEntry{
		Index:        2,
		Timestamp:    now.Round(0),
		Payload:      "monotonic-test",
		PreviousHash: hash1,
		EventID:      "event-456",
	}
	hashStripped := entryStripped.CalculateHash()

	if hashWithMono != hashStripped {
		t.Fatalf("hash changed due to monotonic clock: %s != %s", hashWithMono, hashStripped)
	}
}
