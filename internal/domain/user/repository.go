package user

import (
	"context"
	"errors"

	"github.com/joaovitorpedrosomonteiro/low-carbon-identity-service/internal/domain/valueobject"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type UserRepository interface {
	FindByID(ctx context.Context, id UserID) (User, error)
	FindByEmail(ctx context.Context, email valueobject.Email) (User, error)
	ExistsByEmail(ctx context.Context, email valueobject.Email) (bool, error)
	SaveSystemAdmin(ctx context.Context, admin SystemAdmin) error
	SaveCompanyAdmin(ctx context.Context, admin CompanyAdmin) error
	SaveEmployee(ctx context.Context, emp Employee) error
	SaveAuditor(ctx context.Context, auditor Auditor) error
	UpdateSystemAdmin(ctx context.Context, admin SystemAdmin) error
	UpdateCompanyAdmin(ctx context.Context, admin CompanyAdmin) error
	UpdateEmployee(ctx context.Context, emp Employee) error
	UpdateAuditor(ctx context.Context, auditor Auditor) error
}

type AuditorAccessGrantRepository interface {
	Save(ctx context.Context, grant valueobject.AuditorAccessGrant) error
	FindByAuditorID(ctx context.Context, auditorID string) ([]valueobject.AuditorAccessGrant, error)
	FindByCompanyID(ctx context.Context, companyID string) ([]valueobject.AuditorAccessGrant, error)
	FindByID(ctx context.Context, grantID string) (valueobject.AuditorAccessGrant, error)
	Delete(ctx context.Context, grantID string) error
}

type KnownCompanyRepository interface {
	Upsert(ctx context.Context, companyID, name string) error
	Exists(ctx context.Context, companyID string) (bool, error)
}
