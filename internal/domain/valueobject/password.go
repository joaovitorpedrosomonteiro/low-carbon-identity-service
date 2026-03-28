package valueobject

import (
	"errors"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

type Password struct {
	hash string
}

func NewPassword(plaintext string) (Password, error) {
	if err := validatePasswordStrength(plaintext); err != nil {
		return Password{}, err
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(plaintext), bcryptCost)
	if err != nil {
		return Password{}, errors.New("failed to hash password")
	}
	return Password{hash: string(hashed)}, nil
}

func NewPasswordFromHash(hash string) Password {
	return Password{hash: hash}
}

func (p Password) Matches(plaintext string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(p.hash), []byte(plaintext))
	return err == nil
}

func (p Password) Hash() string {
	return p.hash
}

func validatePasswordStrength(plaintext string) error {
	if len(plaintext) < 12 {
		return errors.New("password must be at least 12 characters")
	}
	var hasUpper, hasNumber, hasSymbol bool
	for _, ch := range plaintext {
		switch {
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsDigit(ch):
			hasNumber = true
		case unicode.IsPunct(ch) || unicode.IsSymbol(ch):
			hasSymbol = true
		}
	}
	if !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}
	if !hasNumber {
		return errors.New("password must contain at least one number")
	}
	if !hasSymbol {
		return errors.New("password must contain at least one symbol")
	}
	return nil
}
