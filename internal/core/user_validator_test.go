package core

import "testing"

func TestValidateRegistration(t *testing.T) {
	valid := RegisterUserRequest{
		Email:    "buyer@example.com",
		FullName: "Jane Doe",
		Role:     Buyer,
	}
	if err := ValidateRegistration(valid); err != nil {
		t.Errorf("ValidateRegistration failed for valid input: %v", err)
	}

	invalidEmail := RegisterUserRequest{
		Email:    "invalid-email",
		FullName: "Jane Doe",
		Role:     Buyer,
	}
	if err := ValidateRegistration(invalidEmail); err == nil {
		t.Error("ValidateRegistration accepted invalid email format")
	}

	invalidRole := RegisterUserRequest{
		Email:    "buyer@example.com",
		FullName: "Jane Doe",
		Role:     "SUPER_HERO",
	}
	if err := ValidateRegistration(invalidRole); err == nil {
		t.Error("ValidateRegistration accepted invalid user role")
	}
}

func TestValidateKYC(t *testing.T) {
	valid := KYCSubmissionRequest{
		DocumentType:      "PASSPORT",
		DocumentReference: "P12345678",
	}
	if err := ValidateKYC(valid); err != nil {
		t.Errorf("ValidateKYC failed for valid input: %v", err)
	}

	invalidRef := KYCSubmissionRequest{
		DocumentType:      "PASSPORT",
		DocumentReference: "AB",
	}
	if err := ValidateKYC(invalidRef); err == nil {
		t.Error("ValidateKYC accepted invalid short document reference")
	}
}
