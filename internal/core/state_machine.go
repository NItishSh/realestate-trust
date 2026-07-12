package core

import "fmt"

type TransactionState string

const (
	Draft     TransactionState = "DRAFT"
	Escrow    TransactionState = "ESCROW"
	Funded    TransactionState = "FUNDED"
	Closed    TransactionState = "CLOSED"
	Cancelled TransactionState = "CANCELLED"
)

var validTransitions = map[TransactionState][]TransactionState{
	Draft:     {Escrow, Cancelled},
	Escrow:    {Funded, Cancelled},
	Funded:    {Closed, Cancelled},
	Closed:    {},
	Cancelled: {},
}

// CanTransition checks if the requested transition is valid.
func CanTransition(current, next TransactionState) bool {
	allowedStates, ok := validTransitions[current]
	if !ok {
		return false
	}
	for _, state := range allowedStates {
		if state == next {
			return true
		}
	}
	return false
}

// Transition changes state or returns an error.
func Transition(current, next TransactionState) (TransactionState, error) {
	if !CanTransition(current, next) {
		return current, fmt.Errorf("invalid transition from %s to %s", current, next)
	}
	return next, nil
}
