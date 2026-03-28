package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/joaovitorpedrosomonteiro/low-carbon-identity-service/internal/domain/user"
	"github.com/joaovitorpedrosomonteiro/low-carbon-identity-service/internal/domain/valueobject"
)

type UserPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewUserPostgresRepository(pool *pgxpool.Pool) *UserPostgresRepository {
	return &UserPostgresRepository{pool: pool}
}

func (r *UserPostgresRepository) FindByID(ctx context.Context, id user.UserID) (user.User, error) {
	var (
		emailStr, roleStr, passwordHash string
		mustChange                      bool
		createdAt, updatedAt            time.Time
		companyID                       *string
		branchID                        *string
		onboardingCompleted             *bool
		icpCertChain, icpCertSerial     *string
	)

	err := r.pool.QueryRow(ctx,
		`SELECT email, role, password_hash, must_change_password, created_at, updated_at,
		        company_id, branch_id, onboarding_completed, icp_certificate_chain, icp_certificate_serial
		 FROM users WHERE id = $1`, string(id),
	).Scan(&emailStr, &roleStr, &passwordHash, &mustChange, &createdAt, &updatedAt,
		&companyID, &branchID, &onboardingCompleted, &icpCertChain, &icpCertSerial)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, user.ErrUserNotFound
		}
		return nil, err
	}

	return scanUser(string(id), emailStr, roleStr, passwordHash, mustChange, createdAt, updatedAt,
		companyID, branchID, onboardingCompleted, icpCertChain, icpCertSerial)
}

func (r *UserPostgresRepository) FindByEmail(ctx context.Context, email valueobject.Email) (user.User, error) {
	var (
		idStr, roleStr, passwordHash     string
		mustChange                       bool
		createdAt, updatedAt             time.Time
		companyID                        *string
		branchID                         *string
		onboardingCompleted              *bool
		icpCertChain, icpCertSerial      *string
	)

	err := r.pool.QueryRow(ctx,
		`SELECT id, role, password_hash, must_change_password, created_at, updated_at,
		        company_id, branch_id, onboarding_completed, icp_certificate_chain, icp_certificate_serial
		 FROM users WHERE email = $1`, email.String(),
	).Scan(&idStr, &roleStr, &passwordHash, &mustChange, &createdAt, &updatedAt,
		&companyID, &branchID, &onboardingCompleted, &icpCertChain, &icpCertSerial)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, user.ErrUserNotFound
		}
		return nil, err
	}

	return scanUser(idStr, email.String(), roleStr, passwordHash, mustChange, createdAt, updatedAt,
		companyID, branchID, onboardingCompleted, icpCertChain, icpCertSerial)
}

func (r *UserPostgresRepository) ExistsByEmail(ctx context.Context, email valueobject.Email) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`, email.String()).Scan(&exists)
	return exists, err
}

func (r *UserPostgresRepository) SaveSystemAdmin(ctx context.Context, admin user.SystemAdmin) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO users (id, email, role, password_hash, must_change_password, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		string(admin.ID()), admin.Email().String(), admin.Role().String(),
		admin.Password().Hash(), admin.MustChangePassword(), admin.CreatedAt(), admin.UpdatedAt(),
	)
	return err
}

func (r *UserPostgresRepository) SaveCompanyAdmin(ctx context.Context, admin user.CompanyAdmin) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO users (id, email, role, password_hash, must_change_password, company_id, onboarding_completed, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		string(admin.ID()), admin.Email().String(), admin.Role().String(),
		admin.Password().Hash(), admin.MustChangePassword(),
		admin.CompanyID(), admin.OnboardingCompleted(), admin.CreatedAt(), admin.UpdatedAt(),
	)
	return err
}

func (r *UserPostgresRepository) SaveEmployee(ctx context.Context, emp user.Employee) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO users (id, email, role, password_hash, must_change_password, branch_id, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		string(emp.ID()), emp.Email().String(), emp.Role().String(),
		emp.Password().Hash(), emp.MustChangePassword(),
		emp.CompanyBranchID(), emp.CreatedAt(), emp.UpdatedAt(),
	)
	return err
}

func (r *UserPostgresRepository) SaveAuditor(ctx context.Context, auditor user.Auditor) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO users (id, email, role, password_hash, must_change_password, icp_certificate_chain, icp_certificate_serial, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		string(auditor.ID()), auditor.Email().String(), auditor.Role().String(),
		auditor.Password().Hash(), auditor.MustChangePassword(),
		auditor.ICPCertificateChain(), auditor.ICPCertificateSerial(),
		auditor.CreatedAt(), auditor.UpdatedAt(),
	)
	return err
}

func (r *UserPostgresRepository) UpdateSystemAdmin(ctx context.Context, admin user.SystemAdmin) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET password_hash = $1, must_change_password = $2, updated_at = $3 WHERE id = $4`,
		admin.Password().Hash(), admin.MustChangePassword(), admin.UpdatedAt(), string(admin.ID()),
	)
	return err
}

func (r *UserPostgresRepository) UpdateCompanyAdmin(ctx context.Context, admin user.CompanyAdmin) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET password_hash = $1, must_change_password = $2, onboarding_completed = $3, updated_at = $4 WHERE id = $5`,
		admin.Password().Hash(), admin.MustChangePassword(), admin.OnboardingCompleted(), admin.UpdatedAt(), string(admin.ID()),
	)
	return err
}

func (r *UserPostgresRepository) UpdateEmployee(ctx context.Context, emp user.Employee) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET password_hash = $1, must_change_password = $2, updated_at = $3 WHERE id = $4`,
		emp.Password().Hash(), emp.MustChangePassword(), emp.UpdatedAt(), string(emp.ID()),
	)
	return err
}

func (r *UserPostgresRepository) UpdateAuditor(ctx context.Context, auditor user.Auditor) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET password_hash = $1, must_change_password = $2, icp_certificate_chain = $3, icp_certificate_serial = $4, updated_at = $5 WHERE id = $6`,
		auditor.Password().Hash(), auditor.MustChangePassword(),
		auditor.ICPCertificateChain(), auditor.ICPCertificateSerial(),
		auditor.UpdatedAt(), string(auditor.ID()),
	)
	return err
}

func scanUser(id, emailStr, roleStr, passwordHash string, mustChange bool, createdAt, updatedAt time.Time,
	companyID, branchID *string, onboardingCompleted *bool, icpCertChain, icpCertSerial *string) (user.User, error) {

	email, err := valueobject.NewEmail(emailStr)
	if err != nil {
		return nil, err
	}
	pw := valueobject.NewPasswordFromHash(passwordHash)

	switch valueobject.Role(roleStr) {
	case valueobject.RoleSystemAdmin:
		return user.NewSystemAdminFromDB(user.UserID(id), email, pw, mustChange, createdAt, updatedAt), nil
	case valueobject.RoleCompanyAdmin:
		cid := ""
		if companyID != nil {
			cid = *companyID
		}
		done := false
		if onboardingCompleted != nil {
			done = *onboardingCompleted
		}
		return user.NewCompanyAdminFromDB(user.UserID(id), email, pw, cid, done, mustChange, createdAt, updatedAt), nil
	case valueobject.RoleEmployee:
		bid := ""
		if branchID != nil {
			bid = *branchID
		}
		return user.NewEmployeeFromDB(user.UserID(id), email, pw, bid, mustChange, createdAt, updatedAt), nil
	case valueobject.RoleAuditor:
		chain := ""
		serial := ""
		if icpCertChain != nil {
			chain = *icpCertChain
		}
		if icpCertSerial != nil {
			serial = *icpCertSerial
		}
		return user.NewAuditorFromDB(user.UserID(id), email, pw, chain, serial, mustChange, createdAt, updatedAt), nil
	}

	return nil, errors.New("unknown user role")
}

type AuditorAccessPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewAuditorAccessPostgresRepository(pool *pgxpool.Pool) *AuditorAccessPostgresRepository {
	return &AuditorAccessPostgresRepository{pool: pool}
}

func (r *AuditorAccessPostgresRepository) Save(ctx context.Context, grant valueobject.AuditorAccessGrant) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO auditor_access_grants (id, auditor_id, scope, inventory_id, company_branch_id, company_id, granted_by, granted_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		grant.ID(), grant.AuditorID(), string(grant.Scope()),
		grant.InventoryID(), grant.CompanyBranchID(), grant.CompanyID(),
		grant.GrantedBy(), grant.GrantedAt(),
	)
	return err
}

func (r *AuditorAccessPostgresRepository) FindByAuditorID(ctx context.Context, auditorID string) ([]valueobject.AuditorAccessGrant, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, auditor_id, scope, inventory_id, company_branch_id, company_id, granted_by, granted_at
		 FROM auditor_access_grants WHERE auditor_id = $1`, auditorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var grants []valueobject.AuditorAccessGrant
	for rows.Next() {
		var g scanGrant
		if err := rows.Scan(&g.id, &g.auditorID, &g.scope, &g.inventoryID, &g.companyBranchID, &g.companyID, &g.grantedBy, &g.grantedAt); err != nil {
			return nil, err
		}
		grants = append(grants, g.toDomain())
	}
	return grants, nil
}

func (r *AuditorAccessPostgresRepository) FindByCompanyID(ctx context.Context, companyID string) ([]valueobject.AuditorAccessGrant, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, auditor_id, scope, inventory_id, company_branch_id, company_id, granted_by, granted_at
		 FROM auditor_access_grants WHERE company_id = $1 OR company_branch_id IN (SELECT id FROM company_branches WHERE company_id = $1)`,
		companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var grants []valueobject.AuditorAccessGrant
	for rows.Next() {
		var g scanGrant
		if err := rows.Scan(&g.id, &g.auditorID, &g.scope, &g.inventoryID, &g.companyBranchID, &g.companyID, &g.grantedBy, &g.grantedAt); err != nil {
			return nil, err
		}
		grants = append(grants, g.toDomain())
	}
	return grants, nil
}

func (r *AuditorAccessPostgresRepository) FindByID(ctx context.Context, grantID string) (valueobject.AuditorAccessGrant, error) {
	var g scanGrant
	err := r.pool.QueryRow(ctx,
		`SELECT id, auditor_id, scope, inventory_id, company_branch_id, company_id, granted_by, granted_at
		 FROM auditor_access_grants WHERE id = $1`, grantID,
	).Scan(&g.id, &g.auditorID, &g.scope, &g.inventoryID, &g.companyBranchID, &g.companyID, &g.grantedBy, &g.grantedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return valueobject.AuditorAccessGrant{}, errors.New("grant not found")
		}
		return valueobject.AuditorAccessGrant{}, err
	}
	return g.toDomain(), nil
}

func (r *AuditorAccessPostgresRepository) Delete(ctx context.Context, grantID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM auditor_access_grants WHERE id = $1`, grantID)
	return err
}

type scanGrant struct {
	id, auditorID, scope string
	inventoryID, companyBranchID, companyID *string
	grantedBy string
	grantedAt time.Time
}

func (g scanGrant) toDomain() valueobject.AuditorAccessGrant {
	return valueobject.NewAuditorAccessGrantFromDB(
		g.id, g.auditorID, valueobject.AuditorAccessScope(g.scope),
		g.inventoryID, g.companyBranchID, g.companyID,
		g.grantedBy, g.grantedAt,
	)
}

type KnownCompanyPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewKnownCompanyPostgresRepository(pool *pgxpool.Pool) *KnownCompanyPostgresRepository {
	return &KnownCompanyPostgresRepository{pool: pool}
}

func (r *KnownCompanyPostgresRepository) Upsert(ctx context.Context, companyID, name string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO known_companies (id, name, created_at) VALUES ($1, $2, $3)
		 ON CONFLICT (id) DO UPDATE SET name = $2`,
		companyID, name, time.Now().UTC(),
	)
	return err
}

func (r *KnownCompanyPostgresRepository) Exists(ctx context.Context, companyID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM known_companies WHERE id = $1)`, companyID).Scan(&exists)
	return exists, err
}
