package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/joaovitorpedrosomonteiro/low-carbon-identity-service/internal/application/command"
	"github.com/joaovitorpedrosomonteiro/low-carbon-identity-service/internal/application/query"
	"github.com/joaovitorpedrosomonteiro/low-carbon-identity-service/internal/domain/user"
)

type AuthHandler struct {
	loginHandler     *command.LoginHandler
	refreshHandler   *command.RefreshTokenHandler
	logoutHandler    *command.LogoutHandler
	changePwHandler  *command.ChangePasswordHandler
	forgotPwHandler  *command.ForgotPasswordHandler
	resetPwHandler   *command.ResetPasswordHandler
	currentUserQuery *query.GetCurrentUserHandler
}

func NewAuthHandler(
	loginHandler *command.LoginHandler,
	refreshHandler *command.RefreshTokenHandler,
	logoutHandler *command.LogoutHandler,
	changePwHandler *command.ChangePasswordHandler,
	forgotPwHandler *command.ForgotPasswordHandler,
	resetPwHandler *command.ResetPasswordHandler,
	currentUserQuery *query.GetCurrentUserHandler,
) *AuthHandler {
	return &AuthHandler{
		loginHandler:     loginHandler,
		refreshHandler:   refreshHandler,
		logoutHandler:    logoutHandler,
		changePwHandler:  changePwHandler,
		forgotPwHandler:  forgotPwHandler,
		resetPwHandler:   resetPwHandler,
		currentUserQuery: currentUserQuery,
	}
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	tokens, err := h.loginHandler.Handle(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, user.ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid email or password")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		return
	}

	writeJSON(w, http.StatusOK, tokens)
}

type refreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	refreshToken := ""

	cookie, err := r.Cookie("refresh_token")
	if err == nil {
		refreshToken = cookie.Value
	} else {
		var req refreshRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
			refreshToken = req.RefreshToken
		}
	}

	if refreshToken == "" {
		writeError(w, http.StatusBadRequest, "MISSING_REFRESH_TOKEN", "Refresh token is required")
		return
	}

	tokens, err := h.refreshHandler.Handle(r.Context(), refreshToken)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "INVALID_REFRESH_TOKEN", "Invalid or expired refresh token")
		return
	}

	writeJSON(w, http.StatusOK, tokens)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing user context")
		return
	}

	if err := h.logoutHandler.Handle(r.Context(), userID); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type changePasswordRequest struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing user context")
		return
	}

	var req changePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if err := h.changePwHandler.Handle(r.Context(), user.UserID(userID), req.CurrentPassword, req.NewPassword); err != nil {
		if errors.Is(err, user.ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Current password is incorrect")
			return
		}
		writeError(w, http.StatusBadRequest, "PASSWORD_CHANGE_FAILED", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type forgotPasswordRequest struct {
	Email string `json:"email"`
}

func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req forgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	h.forgotPwHandler.Handle(r.Context(), req.Email)
	w.WriteHeader(http.StatusOK)
}

type resetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"newPassword"`
}

func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req resetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if err := h.resetPwHandler.Handle(r.Context(), user.UserID(req.Token), req.NewPassword); err != nil {
		writeError(w, http.StatusBadRequest, "RESET_FAILED", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing user context")
		return
	}

	resp, err := h.currentUserQuery.Handle(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

type errorResponse struct {
	Error errorBody `json:"error"`
}

type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(errorResponse{
		Error: errorBody{Code: code, Message: message},
	})
}
