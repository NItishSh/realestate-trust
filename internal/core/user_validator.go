package core

import (
	"errors"
	"html"
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

	RoleBuyer   = Buyer
	RoleSeller  = Seller
	RoleBroker  = Broker
	RoleOfficer = Officer
	RoleAdmin   = Admin
)

// SanitizeString removes leading/trailing spaces and neutralizes HTML injection.
func SanitizeString(s string) string {
	trimmed := strings.TrimSpace(s)
	return html.EscapeString(trimmed)
}

type RegisterUserRequest struct {
	Email    string   `json:"email"`
	Password string   `json:"password"`
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

	if len(req.Password) < 8 {
		return errors.New("password must be at least 8 characters long")
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
