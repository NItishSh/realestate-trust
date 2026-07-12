package core

import (
	"errors"
	"net/mail"
	"strings"
)

type UserRole string

const (
	Buyer   UserRole = "BUYER"
	Seller  UserRole = "SELLER"
	Broker  UserRole = "BROKER"
	Officer UserRole = "OFFICER"
	Admin   UserRole = "ADMIN"
)

type RegisterUserRequest struct {
	Email    string   `json:"email"`
	FullName string   `json:"fullName"`
	Role     UserRole `json:"role"`
}

type KYCSubmissionRequest struct {
	DocumentType      string `json:"documentType"`
	DocumentReference string `json:"documentReference"`
}

// ValidateRegistration checks structural validity of registration input.
func ValidateRegistration(req RegisterUserRequest) error {
	_, err := mail.ParseAddress(req.Email)
	if err != nil {
		return errors.New("invalid email address format")
	}

	if len(strings.TrimSpace(req.FullName)) < 2 {
		return errors.New("full name must be at least 2 characters long")
	}

	validRoles := map[UserRole]bool{
		Buyer:   true,
		Seller:  true,
		Broker:  true,
		Officer: true,
		Admin:   true,
	}

	if !validRoles[req.Role] {
		return errors.New("invalid user role")
	}

	return nil
}

// ValidateKYC checks structural validity of KYC inputs.
func ValidateKYC(req KYCSubmissionRequest) error {
	docType := strings.ToUpper(req.DocumentType)
	if docType != "PASSPORT" && docType != "NATIONAL_ID" && docType != "DRIVERS_LICENSE" {
		return errors.New("invalid document type")
	}

	refLen := len(strings.TrimSpace(req.DocumentReference))
	if refLen < 5 || refLen > 50 {
		return errors.New("document reference must be between 5 and 50 characters")
	}

	return nil
}
