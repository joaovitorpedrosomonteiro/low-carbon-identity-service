package command

import (
	"context"
	"errors"

	"github.com/joaovitorpedrosomonteiro/low-carbon-identity-service/internal/domain/user"
	"github.com/joaovitorpedrosomonteiro/low-carbon-identity-service/internal/domain/valueobject"
)

type TokenPair struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int    `json:"expiresIn"`
}

type EventPublisher interface {
	Publish(ctx context.Context, event user.DomainEvent) error
}

type TokenService interface {
	GenerateTokenPair(userID, role string, companyID, branchID *string, mustChangePassword bool) (TokenPair, error)
	ValidateAccessToken(token string) (map[string]any, error)
	RefreshAccessToken(refreshToken string) (TokenPair, error)
	RevokeAllUserTokens(ctx context.Context, userID string) error
}

type LoginHandler struct {
	userRepo    user.UserRepository
	tokenSvc    TokenService
}

func NewLoginHandler(userRepo user.UserRepository, tokenSvc TokenService) *LoginHandler {
	return &LoginHandler{userRepo: userRepo, tokenSvc: tokenSvc}
}

func (h *LoginHandler) Handle(ctx context.Context, email, password string) (TokenPair, error) {
	Email, err := valueobject.NewEmail(email)
	if err != nil {
		return TokenPair{}, user.ErrInvalidCredentials
	}

	u, err := h.userRepo.FindByEmail(ctx, Email)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			return TokenPair{}, user.ErrInvalidCredentials
		}
		return TokenPair{}, err
	}

	if !u.VerifyPassword(password) {
		return TokenPair{}, user.ErrInvalidCredentials
	}

	var companyID, branchID *string
	switch typed := u.(type) {
	case user.CompanyAdmin:
		cid := typed.CompanyID()
		companyID = &cid
	case user.Employee:
		bid := typed.CompanyBranchID()
		branchID = &bid
	}

	return h.tokenSvc.GenerateTokenPair(
		string(u.ID()),
		u.Role().String(),
		companyID,
		branchID,
		u.MustChangePassword(),
	)
}

type RefreshTokenHandler struct {
	tokenSvc TokenService
}

func NewRefreshTokenHandler(tokenSvc TokenService) *RefreshTokenHandler {
	return &RefreshTokenHandler{tokenSvc: tokenSvc}
}

func (h *RefreshTokenHandler) Handle(ctx context.Context, refreshToken string) (TokenPair, error) {
	return h.tokenSvc.RefreshAccessToken(refreshToken)
}

type ChangePasswordHandler struct {
	userRepo    user.UserRepository
	tokenSvc    TokenService
	publisher   EventPublisher
}

func NewChangePasswordHandler(userRepo user.UserRepository, tokenSvc TokenService, publisher EventPublisher) *ChangePasswordHandler {
	return &ChangePasswordHandler{userRepo: userRepo, tokenSvc: tokenSvc, publisher: publisher}
}

func (h *ChangePasswordHandler) Handle(ctx context.Context, userID user.UserID, currentPassword, newPassword string) error {
	u, err := h.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	if !u.VerifyPassword(currentPassword) {
		return user.ErrInvalidCredentials
	}

	newPw, err := valueobject.NewPassword(newPassword)
	if err != nil {
		return err
	}

	switch typed := u.(type) {
	case user.SystemAdmin:
		updated := typed.ChangePassword(newPw)
		return h.userRepo.UpdateSystemAdmin(ctx, updated)
	case user.CompanyAdmin:
		updated := typed.ChangePassword(newPw)
		return h.userRepo.UpdateCompanyAdmin(ctx, updated)
	case user.Employee:
		updated := typed.ChangePassword(newPw)
		return h.userRepo.UpdateEmployee(ctx, updated)
	case user.Auditor:
		updated := typed.ChangePassword(newPw)
		return h.userRepo.UpdateAuditor(ctx, updated)
	}
	return errors.New("unknown user type")
}

type ForgotPasswordHandler struct {
	userRepo  user.UserRepository
	publisher EventPublisher
}

func NewForgotPasswordHandler(userRepo user.UserRepository, publisher EventPublisher) *ForgotPasswordHandler {
	return &ForgotPasswordHandler{userRepo: userRepo, publisher: publisher}
}

func (h *ForgotPasswordHandler) Handle(ctx context.Context, email string) error {
	Email, err := valueobject.NewEmail(email)
	if err != nil {
		return nil
	}

	u, err := h.userRepo.FindByEmail(ctx, Email)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			return nil
		}
		return err
	}

	resetToken := string(user.NewUserID())
	resetLink := "https://app.lowcarbon.com/reset-password?token=" + resetToken

	event := user.NewPasswordResetRequestedEvent(
		string(u.ID()),
		u.Email().String(),
		resetLink,
	)

	return h.publisher.Publish(ctx, event)
}

type ResetPasswordHandler struct {
	userRepo  user.UserRepository
	publisher EventPublisher
}

func NewResetPasswordHandler(userRepo user.UserRepository, publisher EventPublisher) *ResetPasswordHandler {
	return &ResetPasswordHandler{userRepo: userRepo, publisher: publisher}
}

func (h *ResetPasswordHandler) Handle(ctx context.Context, userID user.UserID, newPassword string) error {
	newPw, err := valueobject.NewPassword(newPassword)
	if err != nil {
		return err
	}

	u, err := h.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	var tempPw string
	switch typed := u.(type) {
	case user.SystemAdmin:
		updated := typed.ChangePassword(newPw)
		tempPw = updated.Password().Hash()
		if err := h.userRepo.UpdateSystemAdmin(ctx, updated); err != nil {
			return err
		}
	case user.CompanyAdmin:
		updated := typed.ChangePassword(newPw)
		tempPw = updated.Password().Hash()
		if err := h.userRepo.UpdateCompanyAdmin(ctx, updated); err != nil {
			return err
		}
	case user.Employee:
		updated := typed.ChangePassword(newPw)
		tempPw = updated.Password().Hash()
		if err := h.userRepo.UpdateEmployee(ctx, updated); err != nil {
			return err
		}
	case user.Auditor:
		updated := typed.ChangePassword(newPw)
		tempPw = updated.Password().Hash()
		if err := h.userRepo.UpdateAuditor(ctx, updated); err != nil {
			return err
		}
	}
	_ = tempPw
	return nil
}

type CreateUserHandler struct {
	userRepo  user.UserRepository
	publisher EventPublisher
}

func NewCreateUserHandler(userRepo user.UserRepository, publisher EventPublisher) *CreateUserHandler {
	return &CreateUserHandler{userRepo: userRepo, publisher: publisher}
}

type CreateSystemAdminInput struct {
	Email string
}

func (h *CreateUserHandler) CreateSystemAdmin(ctx context.Context, input CreateSystemAdminInput) (user.UserID, error) {
	Email, err := valueobject.NewEmail(input.Email)
	if err != nil {
		return "", err
	}

	exists, err := h.userRepo.ExistsByEmail(ctx, Email)
	if err != nil {
		return "", err
	}
	if exists {
		return "", user.ErrEmailAlreadyExists
	}

	tempPass, err := user.GenerateTemporaryPassword()
	if err != nil {
		return "", err
	}

	pw, err := valueobject.NewPassword(tempPass)
	if err != nil {
		return "", err
	}

	admin := user.NewSystemAdmin(Email, pw)
	if err := h.userRepo.SaveSystemAdmin(ctx, admin); err != nil {
		return "", err
	}

	event := user.NewUserCreatedEvent(
		admin.ID(),
		admin.Role().String(),
		admin.Email().String(),
		tempPass,
		nil, nil,
	)

	if err := h.publisher.Publish(ctx, event); err != nil {
		return "", err
	}

	return admin.ID(), nil
}

type CreateCompanyAdminInput struct {
	Email     string
	CompanyID string
}

func (h *CreateUserHandler) CreateCompanyAdmin(ctx context.Context, input CreateCompanyAdminInput) (user.UserID, error) {
	Email, err := valueobject.NewEmail(input.Email)
	if err != nil {
		return "", err
	}

	exists, err := h.userRepo.ExistsByEmail(ctx, Email)
	if err != nil {
		return "", err
	}
	if exists {
		return "", user.ErrEmailAlreadyExists
	}

	tempPass, err := user.GenerateTemporaryPassword()
	if err != nil {
		return "", err
	}

	pw, err := valueobject.NewPassword(tempPass)
	if err != nil {
		return "", err
	}

	admin, err := user.NewCompanyAdmin(Email, pw, input.CompanyID)
	if err != nil {
		return "", err
	}

	if err := h.userRepo.SaveCompanyAdmin(ctx, admin); err != nil {
		return "", err
	}

	cid := input.CompanyID
	event := user.NewUserCreatedEvent(
		admin.ID(),
		admin.Role().String(),
		admin.Email().String(),
		tempPass,
		&cid, nil,
	)

	if err := h.publisher.Publish(ctx, event); err != nil {
		return "", err
	}

	return admin.ID(), nil
}

type CreateEmployeeInput struct {
	Email            string
	CompanyBranchID  string
}

func (h *CreateUserHandler) CreateEmployee(ctx context.Context, input CreateEmployeeInput) (user.UserID, error) {
	Email, err := valueobject.NewEmail(input.Email)
	if err != nil {
		return "", err
	}

	exists, err := h.userRepo.ExistsByEmail(ctx, Email)
	if err != nil {
		return "", err
	}
	if exists {
		return "", user.ErrEmailAlreadyExists
	}

	tempPass, err := user.GenerateTemporaryPassword()
	if err != nil {
		return "", err
	}

	pw, err := valueobject.NewPassword(tempPass)
	if err != nil {
		return "", err
	}

	emp, err := user.NewEmployee(Email, pw, input.CompanyBranchID)
	if err != nil {
		return "", err
	}

	if err := h.userRepo.SaveEmployee(ctx, emp); err != nil {
		return "", err
	}

	bid := input.CompanyBranchID
	event := user.NewUserCreatedEvent(
		emp.ID(),
		emp.Role().String(),
		emp.Email().String(),
		tempPass,
		nil, &bid,
	)

	if err := h.publisher.Publish(ctx, event); err != nil {
		return "", err
	}

	return emp.ID(), nil
}

type CreateAuditorInput struct {
	Email              string
	ICPCertificatePem  string
}

func (h *CreateUserHandler) CreateAuditor(ctx context.Context, input CreateAuditorInput, companyID string) (user.UserID, error) {
	Email, err := valueobject.NewEmail(input.Email)
	if err != nil {
		return "", err
	}

	exists, err := h.userRepo.ExistsByEmail(ctx, Email)
	if err != nil {
		return "", err
	}
	if exists {
		return "", user.ErrEmailAlreadyExists
	}

	tempPass, err := user.GenerateTemporaryPassword()
	if err != nil {
		return "", err
	}

	pw, err := valueobject.NewPassword(tempPass)
	if err != nil {
		return "", err
	}

	serial := extractCertSerial(input.ICPCertificatePem)

	auditor, err := user.NewAuditor(Email, pw, input.ICPCertificatePem, serial)
	if err != nil {
		return "", err
	}

	if err := h.userRepo.SaveAuditor(ctx, auditor); err != nil {
		return "", err
	}

	event := user.NewUserCreatedEvent(
		auditor.ID(),
		auditor.Role().String(),
		auditor.Email().String(),
		tempPass,
		&companyID, nil,
	)

	if err := h.publisher.Publish(ctx, event); err != nil {
		return "", err
	}

	return auditor.ID(), nil
}

func extractCertSerial(pem string) string {
	return "extracted-serial"
}

type UpdateAuditorCertificateHandler struct {
	userRepo user.UserRepository
}

func NewUpdateAuditorCertificateHandler(userRepo user.UserRepository) *UpdateAuditorCertificateHandler {
	return &UpdateAuditorCertificateHandler{userRepo: userRepo}
}

func (h *UpdateAuditorCertificateHandler) Handle(ctx context.Context, auditorID user.UserID, certPem string) error {
	auditorInterface, err := h.userRepo.FindByID(ctx, auditorID)
	if err != nil {
		return err
	}

	auditor, ok := auditorInterface.(user.Auditor)
	if !ok {
		return errors.New("user is not an auditor")
	}

	serial := extractCertSerial(certPem)
	updated := auditor.UpdateCertificate(certPem, serial)
	return h.userRepo.UpdateAuditor(ctx, updated)
}

type GrantAuditorAccessHandler struct {
	grantRepo user.AuditorAccessGrantRepository
	publisher EventPublisher
	userRepo  user.UserRepository
}

func NewGrantAuditorAccessHandler(grantRepo user.AuditorAccessGrantRepository, publisher EventPublisher, userRepo user.UserRepository) *GrantAuditorAccessHandler {
	return &GrantAuditorAccessHandler{grantRepo: grantRepo, publisher: publisher, userRepo: userRepo}
}

type GrantAuditorAccessInput struct {
	AuditorID       string
	Scope           valueobject.AuditorAccessScope
	InventoryID     *string
	CompanyBranchID *string
	CompanyID       *string
	GrantedBy       string
}

func (h *GrantAuditorAccessHandler) Handle(ctx context.Context, input GrantAuditorAccessInput) error {
	grant, err := valueobject.NewAuditorAccessGrant(
		input.AuditorID,
		input.Scope,
		input.GrantedBy,
		input.InventoryID,
		input.CompanyBranchID,
		input.CompanyID,
	)
	if err != nil {
		return err
	}

	if err := h.grantRepo.Save(ctx, grant); err != nil {
		return err
	}

	auditorInterface, err := h.userRepo.FindByID(ctx, user.UserID(input.AuditorID))
	if err != nil {
		return err
	}
	auditorEmail := auditorInterface.Email().String()

	resourceName := string(grant.Scope())
	event := user.NewAuditorAccessGrantedEvent(
		grant.AuditorID(),
		auditorEmail,
		resourceName,
		resourceName,
		grant.InventoryID(),
		grant.CompanyBranchID(),
		grant.CompanyID(),
	)

	return h.publisher.Publish(ctx, event)
}

type RevokeAuditorAccessHandler struct {
	grantRepo user.AuditorAccessGrantRepository
	publisher EventPublisher
}

func NewRevokeAuditorAccessHandler(grantRepo user.AuditorAccessGrantRepository, publisher EventPublisher) *RevokeAuditorAccessHandler {
	return &RevokeAuditorAccessHandler{grantRepo: grantRepo, publisher: publisher}
}

func (h *RevokeAuditorAccessHandler) Handle(ctx context.Context, grantID string) error {
	grant, err := h.grantRepo.FindByID(ctx, grantID)
	if err != nil {
		return err
	}

	if err := h.grantRepo.Delete(ctx, grantID); err != nil {
		return err
	}

	event := user.NewAuditorAccessRevokedEvent(grant.AuditorID(), grantID)
	return h.publisher.Publish(ctx, event)
}

type LogoutHandler struct {
	tokenSvc TokenService
}

func NewLogoutHandler(tokenSvc TokenService) *LogoutHandler {
	return &LogoutHandler{tokenSvc: tokenSvc}
}

func (h *LogoutHandler) Handle(ctx context.Context, userID string) error {
	return h.tokenSvc.RevokeAllUserTokens(ctx, userID)
}

type CertificateExpiryChecker struct {
	userRepo  user.UserRepository
	publisher EventPublisher
}

func NewCertificateExpiryChecker(userRepo user.UserRepository, publisher EventPublisher) *CertificateExpiryChecker {
	return &CertificateExpiryChecker{userRepo: userRepo, publisher: publisher}
}

func (h *CertificateExpiryChecker) Run(ctx context.Context) error {
	return nil
}
