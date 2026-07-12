package core

import "testing"

func TestCanTransition(t *testing.T) {
	tests := []struct {
		current TransactionState
		next    TransactionState
		allowed bool
	}{
		{Draft, Escrow, true},
		{Draft, Cancelled, true},
		{Escrow, Funded, true},
		{Funded, Closed, true},
		{Draft, Funded, false},
		{Escrow, Closed, false},
		{Closed, Draft, false},
		{Cancelled, Escrow, false},
	}

	for _, tt := range tests {
		result := CanTransition(tt.current, tt.next)
		if result != tt.allowed {
			t.Errorf("CanTransition(%s, %s) = %v; want %v", tt.current, tt.next, result, tt.allowed)
		}
	}
}

func TestTransition(t *testing.T) {
	_, err := Transition(Draft, Escrow)
	if err != nil {
		t.Errorf("Transition(Draft, Escrow) failed: %v", err)
	}

	_, err = Transition(Draft, Funded)
	if err == nil {
		t.Error("Transition(Draft, Funded) succeeded, but should have failed")
	}
}
