package query

import (
	"context"

	"github.com/joaovitorpedrosomonteiro/low-carbon-identity-service/internal/domain/user"
)

type GetCurrentUserHandler struct {
	userRepo user.UserRepository
}

func NewGetCurrentUserHandler(userRepo user.UserRepository) *GetCurrentUserHandler {
	return &GetCurrentUserHandler{userRepo: userRepo}
}

type UserResponse struct {
	ID                   string  `json:"id"`
	Email                string  `json:"email"`
	Role                 string  `json:"role"`
	MustChangePassword   bool    `json:"mustChangePassword"`
	CompanyID            *string `json:"companyId,omitempty"`
	CompanyBranchID      *string `json:"companyBranchId,omitempty"`
	OnboardingCompleted  *bool   `json:"onboardingCompleted,omitempty"`
	ICPCertificateSerial *string `json:"icpCertificateSerial,omitempty"`
}

func (h *GetCurrentUserHandler) Handle(ctx context.Context, userID string) (UserResponse, error) {
	u, err := h.userRepo.FindByID(ctx, user.UserID(userID))
	if err != nil {
		return UserResponse{}, err
	}

	resp := UserResponse{
		ID:                 string(u.ID()),
		Email:              u.Email().String(),
		Role:               u.Role().String(),
		MustChangePassword: u.MustChangePassword(),
	}

	switch typed := u.(type) {
	case user.CompanyAdmin:
		cid := typed.CompanyID()
		resp.CompanyID = &cid
		done := typed.OnboardingCompleted()
		resp.OnboardingCompleted = &done
	case user.Employee:
		bid := typed.CompanyBranchID()
		resp.CompanyBranchID = &bid
	case user.Auditor:
		serial := typed.ICPCertificateSerial()
		resp.ICPCertificateSerial = &serial
	}

	return resp, nil
}

type ListAuditorAccessHandler struct {
	grantRepo user.AuditorAccessGrantRepository
}

func NewListAuditorAccessHandler(grantRepo user.AuditorAccessGrantRepository) *ListAuditorAccessHandler {
	return &ListAuditorAccessHandler{grantRepo: grantRepo}
}

type AuditorAccessGrantResponse struct {
	ID              string `json:"id"`
	AuditorID       string `json:"auditorId"`
	Scope           string `json:"scope"`
	InventoryID     *string `json:"inventoryId,omitempty"`
	CompanyBranchID *string `json:"companyBranchId,omitempty"`
	CompanyID       *string `json:"companyId,omitempty"`
	GrantedBy       string `json:"grantedBy"`
	GrantedAt       string `json:"grantedAt"`
}

func (h *ListAuditorAccessHandler) HandleByCompany(ctx context.Context, companyID string) ([]AuditorAccessGrantResponse, error) {
	grants, err := h.grantRepo.FindByCompanyID(ctx, companyID)
	if err != nil {
		return nil, err
	}

	result := make([]AuditorAccessGrantResponse, len(grants))
	for i, g := range grants {
		result[i] = AuditorAccessGrantResponse{
			ID:              g.ID(),
			AuditorID:       g.AuditorID(),
			Scope:           string(g.Scope()),
			InventoryID:     g.InventoryID(),
			CompanyBranchID: g.CompanyBranchID(),
			CompanyID:       g.CompanyID(),
			GrantedBy:       g.GrantedBy(),
			GrantedAt:       g.GrantedAt().Format("2006-01-02T15:04:05Z"),
		}
	}
	return result, nil
}

type ListAuditorAccessByAuditorHandler struct {
	grantRepo user.AuditorAccessGrantRepository
}

func NewListAuditorAccessByAuditorHandler(grantRepo user.AuditorAccessGrantRepository) *ListAuditorAccessByAuditorHandler {
	return &ListAuditorAccessByAuditorHandler{grantRepo: grantRepo}
}

func (h *ListAuditorAccessByAuditorHandler) Handle(ctx context.Context, auditorID string) ([]AuditorAccessGrantResponse, error) {
	grants, err := h.grantRepo.FindByAuditorID(ctx, auditorID)
	if err != nil {
		return nil, err
	}

	result := make([]AuditorAccessGrantResponse, len(grants))
	for i, g := range grants {
		result[i] = AuditorAccessGrantResponse{
			ID:              g.ID(),
			AuditorID:       g.AuditorID(),
			Scope:           string(g.Scope()),
			InventoryID:     g.InventoryID(),
			CompanyBranchID: g.CompanyBranchID(),
			CompanyID:       g.CompanyID(),
			GrantedBy:       g.GrantedBy(),
			GrantedAt:       g.GrantedAt().Format("2006-01-02T15:04:05Z"),
		}
	}
	return result, nil
}

type KnownCompanyRepository = user.KnownCompanyRepository

type ValidateCompanyExistsHandler struct {
	companyRepo KnownCompanyRepository
}

func NewValidateCompanyExistsHandler(companyRepo KnownCompanyRepository) *ValidateCompanyExistsHandler {
	return &ValidateCompanyExistsHandler{companyRepo: companyRepo}
}

func (h *ValidateCompanyExistsHandler) Handle(ctx context.Context, companyID string) (bool, error) {
	return h.companyRepo.Exists(ctx, companyID)
}
