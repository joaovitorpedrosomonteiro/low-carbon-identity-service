package valueobject

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"
)

type AuditorAccessScope string

const (
	AuditorAccessScopeSingleInventory AuditorAccessScope = "single_inventory"
	AuditorAccessScopeCompanyBranch   AuditorAccessScope = "company_branch"
	AuditorAccessScopeCompany         AuditorAccessScope = "company"
)

type AuditorAccessGrant struct {
	id              string
	auditorID       string
	scope           AuditorAccessScope
	inventoryID     *string
	companyBranchID *string
	companyID       *string
	grantedBy       string
	grantedAt       time.Time
}

func NewAuditorAccessGrant(
	auditorID string,
	scope AuditorAccessScope,
	grantedBy string,
	inventoryID *string,
	companyBranchID *string,
	companyID *string,
) (AuditorAccessGrant, error) {
	if auditorID == "" {
		return AuditorAccessGrant{}, errors.New("auditorID is required")
	}
	if grantedBy == "" {
		return AuditorAccessGrant{}, errors.New("grantedBy is required")
	}
	switch scope {
	case AuditorAccessScopeSingleInventory:
		if inventoryID == nil {
			return AuditorAccessGrant{}, errors.New("inventoryID required for single_inventory scope")
		}
	case AuditorAccessScopeCompanyBranch:
		if companyBranchID == nil {
			return AuditorAccessGrant{}, errors.New("companyBranchID required for company_branch scope")
		}
	case AuditorAccessScopeCompany:
		if companyID == nil {
			return AuditorAccessGrant{}, errors.New("companyID required for company scope")
		}
	default:
		return AuditorAccessGrant{}, errors.New("invalid auditor access scope")
	}

	idBytes := make([]byte, 16)
	if _, err := rand.Read(idBytes); err != nil {
		return AuditorAccessGrant{}, errors.New("failed to generate grant ID")
	}

	return AuditorAccessGrant{
		id:              hex.EncodeToString(idBytes),
		auditorID:       auditorID,
		scope:           scope,
		inventoryID:     inventoryID,
		companyBranchID: companyBranchID,
		companyID:       companyID,
		grantedBy:       grantedBy,
		grantedAt:       time.Now().UTC(),
	}, nil
}

func NewAuditorAccessGrantFromDB(
	id string,
	auditorID string,
	scope AuditorAccessScope,
	inventoryID *string,
	companyBranchID *string,
	companyID *string,
	grantedBy string,
	grantedAt time.Time,
) AuditorAccessGrant {
	return AuditorAccessGrant{
		id:              id,
		auditorID:       auditorID,
		scope:           scope,
		inventoryID:     inventoryID,
		companyBranchID: companyBranchID,
		companyID:       companyID,
		grantedBy:       grantedBy,
		grantedAt:       grantedAt,
	}
}

func (g AuditorAccessGrant) ID() string                    { return g.id }
func (g AuditorAccessGrant) AuditorID() string             { return g.auditorID }
func (g AuditorAccessGrant) Scope() AuditorAccessScope     { return g.scope }
func (g AuditorAccessGrant) InventoryID() *string          { return g.inventoryID }
func (g AuditorAccessGrant) CompanyBranchID() *string      { return g.companyBranchID }
func (g AuditorAccessGrant) CompanyID() *string            { return g.companyID }
func (g AuditorAccessGrant) GrantedBy() string             { return g.grantedBy }
func (g AuditorAccessGrant) GrantedAt() time.Time          { return g.grantedAt }

func (g AuditorAccessGrant) CoversInventory(inventoryID, branchID, companyID string) bool {
	switch g.scope {
	case AuditorAccessScopeSingleInventory:
		return g.inventoryID != nil && *g.inventoryID == inventoryID
	case AuditorAccessScopeCompanyBranch:
		return g.companyBranchID != nil && *g.companyBranchID == branchID
	case AuditorAccessScopeCompany:
		return g.companyID != nil && *g.companyID == companyID
	}
	return false
}
