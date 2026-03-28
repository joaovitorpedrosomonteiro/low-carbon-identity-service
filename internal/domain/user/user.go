package user

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/joaovitorpedrosomonteiro/low-carbon-identity-service/internal/domain/valueobject"
)

type UserID string

func NewUserID() UserID {
	b := make([]byte, 16)
	rand.Read(b)
	return UserID(hex.EncodeToString(b))
}

type User interface {
	ID() UserID
	Email() valueobject.Email
	Role() valueobject.Role
	MustChangePassword() bool
	VerifyPassword(plaintext string) bool
}

type SystemAdmin struct {
	id                  UserID
	email               valueobject.Email
	password            valueobject.Password
	mustChangePassword  bool
	createdAt           time.Time
	updatedAt           time.Time
}

func NewSystemAdmin(email valueobject.Email, password valueobject.Password) SystemAdmin {
	return SystemAdmin{
		id:                 NewUserID(),
		email:              email,
		password:           password,
		mustChangePassword: true,
		createdAt:          time.Now().UTC(),
		updatedAt:          time.Now().UTC(),
	}
}

func NewSystemAdminFromDB(id UserID, email valueobject.Email, password valueobject.Password, mustChange bool, createdAt, updatedAt time.Time) SystemAdmin {
	return SystemAdmin{id: id, email: email, password: password, mustChangePassword: mustChange, createdAt: createdAt, updatedAt: updatedAt}
}

func (s SystemAdmin) ID() UserID                  { return s.id }
func (s SystemAdmin) Email() valueobject.Email    { return s.email }
func (s SystemAdmin) Role() valueobject.Role      { return valueobject.RoleSystemAdmin }
func (s SystemAdmin) MustChangePassword() bool    { return s.mustChangePassword }
func (s SystemAdmin) Password() valueobject.Password { return s.password }
func (s SystemAdmin) CreatedAt() time.Time        { return s.createdAt }
func (s SystemAdmin) UpdatedAt() time.Time        { return s.updatedAt }

func (s SystemAdmin) VerifyPassword(plaintext string) bool {
	return s.password.Matches(plaintext)
}

func (s SystemAdmin) ChangePassword(newPassword valueobject.Password) SystemAdmin {
	s.password = newPassword
	s.mustChangePassword = false
	s.updatedAt = time.Now().UTC()
	return s
}

type CompanyAdmin struct {
	id                  UserID
	email               valueobject.Email
	password            valueobject.Password
	companyID           string
	onboardingCompleted bool
	mustChangePassword  bool
	createdAt           time.Time
	updatedAt           time.Time
}

func NewCompanyAdmin(email valueobject.Email, password valueobject.Password, companyID string) (CompanyAdmin, error) {
	if companyID == "" {
		return CompanyAdmin{}, errors.New("companyID is required")
	}
	return CompanyAdmin{
		id:                 NewUserID(),
		email:              email,
		password:           password,
		companyID:          companyID,
		mustChangePassword: true,
		createdAt:          time.Now().UTC(),
		updatedAt:          time.Now().UTC(),
	}, nil
}

func NewCompanyAdminFromDB(id UserID, email valueobject.Email, password valueobject.Password, companyID string, onboardingCompleted, mustChange bool, createdAt, updatedAt time.Time) CompanyAdmin {
	return CompanyAdmin{id: id, email: email, password: password, companyID: companyID, onboardingCompleted: onboardingCompleted, mustChangePassword: mustChange, createdAt: createdAt, updatedAt: updatedAt}
}

func (c CompanyAdmin) ID() UserID                  { return c.id }
func (c CompanyAdmin) Email() valueobject.Email    { return c.email }
func (c CompanyAdmin) Role() valueobject.Role      { return valueobject.RoleCompanyAdmin }
func (c CompanyAdmin) MustChangePassword() bool    { return c.mustChangePassword }
func (c CompanyAdmin) Password() valueobject.Password { return c.password }
func (c CompanyAdmin) CompanyID() string           { return c.companyID }
func (c CompanyAdmin) OnboardingCompleted() bool   { return c.onboardingCompleted }
func (c CompanyAdmin) CreatedAt() time.Time        { return c.createdAt }
func (c CompanyAdmin) UpdatedAt() time.Time        { return c.updatedAt }

func (c CompanyAdmin) VerifyPassword(plaintext string) bool {
	return c.password.Matches(plaintext)
}

func (c CompanyAdmin) ChangePassword(newPassword valueobject.Password) CompanyAdmin {
	c.password = newPassword
	c.mustChangePassword = false
	c.updatedAt = time.Now().UTC()
	return c
}

func (c CompanyAdmin) CompleteOnboarding() CompanyAdmin {
	c.onboardingCompleted = true
	c.updatedAt = time.Now().UTC()
	return c
}

type Employee struct {
	id                 UserID
	email              valueobject.Email
	password           valueobject.Password
	companyBranchID    string
	mustChangePassword bool
	createdAt          time.Time
	updatedAt          time.Time
}

func NewEmployee(email valueobject.Email, password valueobject.Password, companyBranchID string) (Employee, error) {
	if companyBranchID == "" {
		return Employee{}, errors.New("companyBranchID is required")
	}
	return Employee{
		id:                 NewUserID(),
		email:              email,
		password:           password,
		companyBranchID:    companyBranchID,
		mustChangePassword: true,
		createdAt:          time.Now().UTC(),
		updatedAt:          time.Now().UTC(),
	}, nil
}

func NewEmployeeFromDB(id UserID, email valueobject.Email, password valueobject.Password, companyBranchID string, mustChange bool, createdAt, updatedAt time.Time) Employee {
	return Employee{id: id, email: email, password: password, companyBranchID: companyBranchID, mustChangePassword: mustChange, createdAt: createdAt, updatedAt: updatedAt}
}

func (e Employee) ID() UserID                  { return e.id }
func (e Employee) Email() valueobject.Email    { return e.email }
func (e Employee) Role() valueobject.Role      { return valueobject.RoleEmployee }
func (e Employee) MustChangePassword() bool    { return e.mustChangePassword }
func (e Employee) Password() valueobject.Password { return e.password }
func (e Employee) CompanyBranchID() string     { return e.companyBranchID }
func (e Employee) CreatedAt() time.Time        { return e.createdAt }
func (e Employee) UpdatedAt() time.Time        { return e.updatedAt }

func (e Employee) VerifyPassword(plaintext string) bool {
	return e.password.Matches(plaintext)
}

func (e Employee) ChangePassword(newPassword valueobject.Password) Employee {
	e.password = newPassword
	e.mustChangePassword = false
	e.updatedAt = time.Now().UTC()
	return e
}

type Auditor struct {
	id                    UserID
	email                 valueobject.Email
	password              valueobject.Password
	icpCertificateChain   string
	icpCertificateSerial  string
	mustChangePassword    bool
	createdAt             time.Time
	updatedAt             time.Time
}

func NewAuditor(email valueobject.Email, password valueobject.Password, icpCertChain, icpCertSerial string) (Auditor, error) {
	if icpCertChain == "" {
		return Auditor{}, errors.New("ICP certificate chain is required")
	}
	if icpCertSerial == "" {
		return Auditor{}, errors.New("ICP certificate serial is required")
	}
	return Auditor{
		id:                   NewUserID(),
		email:                email,
		password:             password,
		icpCertificateChain:  icpCertChain,
		icpCertificateSerial: icpCertSerial,
		mustChangePassword:   true,
		createdAt:            time.Now().UTC(),
		updatedAt:            time.Now().UTC(),
	}, nil
}

func NewAuditorFromDB(id UserID, email valueobject.Email, password valueobject.Password, icpCertChain, icpCertSerial string, mustChange bool, createdAt, updatedAt time.Time) Auditor {
	return Auditor{id: id, email: email, password: password, icpCertificateChain: icpCertChain, icpCertificateSerial: icpCertSerial, mustChangePassword: mustChange, createdAt: createdAt, updatedAt: updatedAt}
}

func (a Auditor) ID() UserID                  { return a.id }
func (a Auditor) Email() valueobject.Email    { return a.email }
func (a Auditor) Role() valueobject.Role      { return valueobject.RoleAuditor }
func (a Auditor) MustChangePassword() bool    { return a.mustChangePassword }
func (a Auditor) Password() valueobject.Password { return a.password }
func (a Auditor) ICPCertificateChain() string  { return a.icpCertificateChain }
func (a Auditor) ICPCertificateSerial() string { return a.icpCertificateSerial }
func (a Auditor) CreatedAt() time.Time        { return a.createdAt }
func (a Auditor) UpdatedAt() time.Time        { return a.updatedAt }

func (a Auditor) VerifyPassword(plaintext string) bool {
	return a.password.Matches(plaintext)
}

func (a Auditor) ChangePassword(newPassword valueobject.Password) Auditor {
	a.password = newPassword
	a.mustChangePassword = false
	a.updatedAt = time.Now().UTC()
	return a
}

func (a Auditor) UpdateCertificate(chain, serial string) Auditor {
	a.icpCertificateChain = chain
	a.icpCertificateSerial = serial
	a.updatedAt = time.Now().UTC()
	return a
}

func GenerateTemporaryPassword() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
