package handler

import (
	"encoding/json"
	"net/http"

	"github.com/joaovitorpedrosomonteiro/low-carbon-identity-service/internal/application/command"
	"github.com/joaovitorpedrosomonteiro/low-carbon-identity-service/internal/application/query"
	"github.com/joaovitorpedrosomonteiro/low-carbon-identity-service/internal/domain/user"
	"github.com/joaovitorpedrosomonteiro/low-carbon-identity-service/internal/domain/valueobject"
)

type UserHandler struct {
	createHandler *command.CreateUserHandler
	updateCertHandler *command.UpdateAuditorCertificateHandler
}

func NewUserHandler(createHandler *command.CreateUserHandler, updateCertHandler *command.UpdateAuditorCertificateHandler) *UserHandler {
	return &UserHandler{createHandler: createHandler, updateCertHandler: updateCertHandler}
}

type createSystemAdminRequest struct {
	Email string `json:"email"`
}

func (h *UserHandler) CreateSystemAdmin(w http.ResponseWriter, r *http.Request) {
	var req createSystemAdminRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	id, err := h.createHandler.CreateSystemAdmin(r.Context(), command.CreateSystemAdminInput{Email: req.Email})
	if err != nil {
		if err == user.ErrEmailAlreadyExists {
			writeError(w, http.StatusConflict, "EMAIL_EXISTS", "Email already registered")
			return
		}
		writeError(w, http.StatusBadRequest, "CREATION_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"id": string(id)})
}

type createCompanyAdminRequest struct {
	Email     string `json:"email"`
	CompanyID string `json:"companyId"`
}

func (h *UserHandler) CreateCompanyAdmin(w http.ResponseWriter, r *http.Request) {
	var req createCompanyAdminRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	id, err := h.createHandler.CreateCompanyAdmin(r.Context(), command.CreateCompanyAdminInput{
		Email:     req.Email,
		CompanyID: req.CompanyID,
	})
	if err != nil {
		if err == user.ErrEmailAlreadyExists {
			writeError(w, http.StatusConflict, "EMAIL_EXISTS", "Email already registered")
			return
		}
		writeError(w, http.StatusBadRequest, "CREATION_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"id": string(id)})
}

type createEmployeeRequest struct {
	Email           string `json:"email"`
	CompanyBranchID string `json:"companyBranchId"`
}

func (h *UserHandler) CreateEmployee(w http.ResponseWriter, r *http.Request) {
	var req createEmployeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	id, err := h.createHandler.CreateEmployee(r.Context(), command.CreateEmployeeInput{
		Email:           req.Email,
		CompanyBranchID: req.CompanyBranchID,
	})
	if err != nil {
		if err == user.ErrEmailAlreadyExists {
			writeError(w, http.StatusConflict, "EMAIL_EXISTS", "Email already registered")
			return
		}
		writeError(w, http.StatusBadRequest, "CREATION_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"id": string(id)})
}

type createAuditorRequest struct {
	Email             string `json:"email"`
	ICPCertificatePem string `json:"icpCertificatePem"`
}

func (h *UserHandler) CreateAuditor(w http.ResponseWriter, r *http.Request) {
	var req createAuditorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	companyID := r.Header.Get("X-Company-ID")
	if companyID == "" {
		companyID = "unknown"
	}

	id, err := h.createHandler.CreateAuditor(r.Context(), command.CreateAuditorInput{
		Email:             req.Email,
		ICPCertificatePem: req.ICPCertificatePem,
	}, companyID)
	if err != nil {
		if err == user.ErrEmailAlreadyExists {
			writeError(w, http.StatusConflict, "EMAIL_EXISTS", "Email already registered")
			return
		}
		writeError(w, http.StatusBadRequest, "CREATION_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"id": string(id)})
}

type updateCertificateRequest struct {
	ICPCertificatePem string `json:"icpCertificatePem"`
}

func (h *UserHandler) UpdateAuditorCertificate(w http.ResponseWriter, r *http.Request) {
	auditorID := r.PathValue("id")
	if auditorID == "" {
		writeError(w, http.StatusBadRequest, "MISSING_ID", "Auditor ID is required")
		return
	}

	var req updateCertificateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if err := h.updateCertHandler.Handle(r.Context(), user.UserID(auditorID), req.ICPCertificatePem); err != nil {
		writeError(w, http.StatusBadRequest, "UPDATE_FAILED", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type AuditorAccessHandler struct {
	grantHandler  *command.GrantAuditorAccessHandler
	revokeHandler *command.RevokeAuditorAccessHandler
	listHandler   *query.ListAuditorAccessHandler
}

func NewAuditorAccessHandler(
	grantHandler *command.GrantAuditorAccessHandler,
	revokeHandler *command.RevokeAuditorAccessHandler,
	listHandler *query.ListAuditorAccessHandler,
) *AuditorAccessHandler {
	return &AuditorAccessHandler{
		grantHandler:  grantHandler,
		revokeHandler: revokeHandler,
		listHandler:   listHandler,
	}
}

type grantAccessRequest struct {
	AuditorID       string `json:"auditorId"`
	Scope           string `json:"scope"`
	InventoryID     *string `json:"inventoryId,omitempty"`
	CompanyBranchID *string `json:"companyBranchId,omitempty"`
	CompanyID       *string `json:"companyId,omitempty"`
}

func (h *AuditorAccessHandler) Grant(w http.ResponseWriter, r *http.Request) {
	var req grantAccessRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	grantedBy := r.Header.Get("X-User-ID")

	if err := h.grantHandler.Handle(r.Context(), command.GrantAuditorAccessInput{
		AuditorID:       req.AuditorID,
		Scope:           valueobject.AuditorAccessScope(req.Scope),
		InventoryID:     req.InventoryID,
		CompanyBranchID: req.CompanyBranchID,
		CompanyID:       req.CompanyID,
		GrantedBy:       grantedBy,
	}); err != nil {
		writeError(w, http.StatusBadRequest, "GRANT_FAILED", err.Error())
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *AuditorAccessHandler) Revoke(w http.ResponseWriter, r *http.Request) {
	grantID := r.PathValue("grantId")
	if grantID == "" {
		writeError(w, http.StatusBadRequest, "MISSING_ID", "Grant ID is required")
		return
	}

	if err := h.revokeHandler.Handle(r.Context(), grantID); err != nil {
		writeError(w, http.StatusBadRequest, "REVOKE_FAILED", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *AuditorAccessHandler) List(w http.ResponseWriter, r *http.Request) {
	companyID := r.Header.Get("X-Company-ID")
	if companyID == "" {
		writeError(w, http.StatusBadRequest, "MISSING_COMPANY", "Company ID is required")
		return
	}

	grants, err := h.listHandler.HandleByCompany(r.Context(), companyID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": grants})
}
