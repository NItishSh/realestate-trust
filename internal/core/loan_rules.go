package core

import "errors"

const MaxLTVRatio = 0.80

// IsWithinLTVLimit checks if loan request matches LTV criteria.
func IsWithinLTVLimit(requestedAmount, propertyValue float64) bool {
	if propertyValue <= 0 {
		return false
	}
	ltv := requestedAmount / propertyValue
	return ltv <= MaxLTVRatio
}

// ValidateRequest returns an error if LTV thresholds are violated.
func ValidateRequest(requestedAmount, propertyValue float64) error {
	if requestedAmount <= 0 {
		return errors.New("requested loan amount must be positive")
	}
	if !IsWithinLTVLimit(requestedAmount, propertyValue) {
		return errors.New("requested amount exceeds the maximum allowed Loan-to-Value ratio of 80%")
	}
	return nil
}
