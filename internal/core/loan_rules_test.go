package core

import "testing"

func TestIsWithinLTVLimit(t *testing.T) {
	if !IsWithinLTVLimit(80000, 100000) {
		t.Error("LTV of 80% should be accepted")
	}
	if !IsWithinLTVLimit(50000, 100000) {
		t.Error("LTV of 50% should be accepted")
	}
	if IsWithinLTVLimit(81000, 100000) {
		t.Error("LTV of 81% should be rejected")
	}
	if IsWithinLTVLimit(50000, 0) {
		t.Error("LTV with 0 property value should be rejected")
	}
}

func TestValidateRequest(t *testing.T) {
	err := ValidateRequest(75000, 100000)
	if err != nil {
		t.Errorf("ValidateRequest(75000, 100000) failed: %v", err)
	}

	err = ValidateRequest(0, 100000)
	if err == nil {
		t.Error("ValidateRequest(0, 100000) succeeded but should have failed")
	}

	err = ValidateRequest(85000, 100000)
	if err == nil {
		t.Error("ValidateRequest(85000, 100000) succeeded but should have failed")
	}
}
