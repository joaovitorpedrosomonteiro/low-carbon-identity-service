package user

import "time"

type DomainEvent struct {
	EventID     string    `json:"event_id"`
	EventType   string    `json:"event_type"`
	OccurredAt  time.Time `json:"occurred_at"`
	SchemaVer   string    `json:"schema_version"`
	Traceparent string    `json:"traceparent,omitempty"`
	Payload     any       `json:"payload"`
}

type UserCreatedPayload struct {
	UserID             string  `json:"userID"`
	Role               string  `json:"role"`
	Email              string  `json:"email"`
	TemporaryPassword  string  `json:"temporaryPassword"`
	CompanyID          *string `json:"companyID,omitempty"`
	BranchID           *string `json:"branchID,omitempty"`
}

type UserPasswordResetPayload struct {
	UserID            string `json:"userID"`
	Role              string `json:"role"`
	Email             string `json:"email"`
	TemporaryPassword string `json:"temporaryPassword"`
}

type AuditorAccessGrantedPayload struct {
	AuditorID       string  `json:"auditorID"`
	AuditorEmail    string  `json:"auditorEmail"`
	Scope           string  `json:"scope"`
	ResourceName    string  `json:"resourceName"`
	InventoryID     *string `json:"inventoryID,omitempty"`
	BranchID        *string `json:"branchID,omitempty"`
	CompanyID       *string `json:"companyID,omitempty"`
}

type AuditorAccessRevokedPayload struct {
	AuditorID string `json:"auditorID"`
	GrantID   string `json:"grantID"`
}

type PasswordResetRequestedPayload struct {
	UserID    string `json:"userID"`
	Email     string `json:"email"`
	ResetLink string `json:"resetLink"`
}

type CertificateExpiryWarningPayload struct {
	AuditorID        string `json:"auditorID"`
	AuditorEmail     string `json:"auditorEmail"`
	CompanyAdminEmail string `json:"companyAdminEmail"`
	DaysUntilExpiry  int    `json:"daysUntilExpiry"`
}

func NewUserCreatedEvent(id UserID, role string, email, tempPassword string, companyID, branchID *string) DomainEvent {
	return DomainEvent{
		EventID:    string(NewUserID()),
		EventType:  "UserCreated",
		OccurredAt: time.Now().UTC(),
		SchemaVer:  "1.0",
		Payload: UserCreatedPayload{
			UserID:            string(id),
			Role:              role,
			Email:             email,
			TemporaryPassword: tempPassword,
			CompanyID:         companyID,
			BranchID:          branchID,
		},
	}
}

func NewUserPasswordResetEvent(id UserID, role, email, tempPassword string) DomainEvent {
	return DomainEvent{
		EventID:    string(NewUserID()),
		EventType:  "UserPasswordReset",
		OccurredAt: time.Now().UTC(),
		SchemaVer:  "1.0",
		Payload: UserPasswordResetPayload{
			UserID:            string(id),
			Role:              role,
			Email:             email,
			TemporaryPassword: tempPassword,
		},
	}
}

func NewAuditorAccessGrantedEvent(auditorID, auditorEmail, scope, resourceName string, inventoryID, branchID, companyID *string) DomainEvent {
	return DomainEvent{
		EventID:    string(NewUserID()),
		EventType:  "AuditorAccessGranted",
		OccurredAt: time.Now().UTC(),
		SchemaVer:  "1.0",
		Payload: AuditorAccessGrantedPayload{
			AuditorID:    auditorID,
			AuditorEmail: auditorEmail,
			Scope:        scope,
			ResourceName: resourceName,
			InventoryID:  inventoryID,
			BranchID:     branchID,
			CompanyID:    companyID,
		},
	}
}

func NewAuditorAccessRevokedEvent(auditorID, grantID string) DomainEvent {
	return DomainEvent{
		EventID:    string(NewUserID()),
		EventType:  "AuditorAccessRevoked",
		OccurredAt: time.Now().UTC(),
		SchemaVer:  "1.0",
		Payload: AuditorAccessRevokedPayload{
			AuditorID: auditorID,
			GrantID:   grantID,
		},
	}
}

func NewPasswordResetRequestedEvent(userID, email, resetLink string) DomainEvent {
	return DomainEvent{
		EventID:    string(NewUserID()),
		EventType:  "PasswordResetRequested",
		OccurredAt: time.Now().UTC(),
		SchemaVer:  "1.0",
		Payload: PasswordResetRequestedPayload{
			UserID:    userID,
			Email:     email,
			ResetLink: resetLink,
		},
	}
}
