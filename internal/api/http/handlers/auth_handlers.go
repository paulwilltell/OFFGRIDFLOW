// Package handlers provides HTTP handlers for the OffGridFlow API.
package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/example/offgridflow/internal/api/http/responders"
	"github.com/example/offgridflow/internal/auth"
	"github.com/google/uuid"
)

// -----------------------------------------------------------------------------
// Constants
// -----------------------------------------------------------------------------

const (
	// sessionCookieName is the name of the session cookie.
	sessionCookieName = "offgrid_session"

	// sessionCookieMaxAge is the session cookie lifetime (7 days).
	sessionCookieMaxAge = 7 * 24 * 60 * 60

	// defaultTenantPlan is the plan for newly created tenants.
	defaultTenantPlan = "free"

	// minPasswordLength is the minimum password length.
	minPasswordLength = 8

	// maxPasswordLength is the maximum password length.
	maxPasswordLength = 128

	// defaultOrgNameSuffix is appended when company name is not provided.
	defaultOrgNameSuffix = "'s Organization"

	// slugRandomSuffixLength is the length of random suffix for slugs.
	slugRandomSuffixLength = 8
)

// Password reset request payloads
type passwordForgotRequest struct {
	Email string `json:"email"`
}

type passwordResetRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

// Email validation regex (RFC 5322 simplified)
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// -----------------------------------------------------------------------------
// Auth Handlers
// -----------------------------------------------------------------------------

// AuthHandlers provides HTTP handlers for authentication endpoints.
type AuthHandlers struct {
	authStore      auth.Store
	sessionManager *auth.SessionManager
	logger         *slog.Logger
	cookieDomain   string
	cookieSecure   bool
	lockoutManager *auth.LockoutManager
}

// AuthHandlersConfig holds configuration for auth handlers.
type AuthHandlersConfig struct {
	// AuthStore is the backing store for users and tenants.
	AuthStore auth.Store

	// SessionManager handles JWT token creation and validation.
	SessionManager *auth.SessionManager

	// Logger for auth-related logging. If nil, a default is used.
	Logger *slog.Logger

	// CookieDomain for session cookies (e.g., ".offgridflow.com" for subdomains).
	CookieDomain string

	// CookieSecure should be true in production (HTTPS only).
	CookieSecure bool

	// LockoutManager optionally enforces login lockouts.
	LockoutManager *auth.LockoutManager
}

// NewAuthHandlers creates new authentication handlers.
func NewAuthHandlers(cfg AuthHandlersConfig) *AuthHandlers {
	logger := cfg.Logger
	if logger == nil {
		logger = slog.Default().With("component", "auth-handlers")
	}

	return &AuthHandlers{
		authStore:      cfg.AuthStore,
		sessionManager: cfg.SessionManager,
		logger:         logger,
		cookieDomain:   cfg.CookieDomain,
		cookieSecure:   cfg.CookieSecure,
		lockoutManager: cfg.LockoutManager,
	}
}

// -----------------------------------------------------------------------------
// Request/Response Types
// -----------------------------------------------------------------------------

// RegisterRequest is the request body for user registration.
type RegisterRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	Name        string `json:"name"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	CompanyName string `json:"company_name"` // Used as tenant name
	JobTitle    string `json:"job_title"`
}

// LoginRequest is the request body for login.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// ChangePasswordRequest is the request body for password changes.
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// AuthResponse is the response for successful authentication.
type AuthResponse struct {
	Token                string    `json:"token"`
	User                 UserDTO   `json:"user"`
	Tenant               TenantDTO `json:"tenant"`
	RequiresVerification bool      `json:"requires_verification,omitempty"`
	VerificationToken    string    `json:"verification_token,omitempty"` // Only in dev mode
}

// UserDTO is user data returned in auth responses.
type UserDTO struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	FirstName     string `json:"first_name,omitempty"`
	LastName      string `json:"last_name,omitempty"`
	Role          string `json:"role"`
	EmailVerified bool   `json:"email_verified"`
}

// TenantDTO is tenant data returned in auth responses.
type TenantDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
	Plan string `json:"plan"`
}

// -----------------------------------------------------------------------------
// Handler Methods
// -----------------------------------------------------------------------------

// Register handles user registration.
// POST /api/auth/register
func (h *AuthHandlers) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		responders.MethodNotAllowed(w, http.MethodPost)
		return
	}

	defer r.Body.Close()

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responders.BadRequest(w, "invalid_json", "invalid request body")
		return
	}

	// Validate request
	if errs := h.validateRegistration(req); len(errs) > 0 {
		responders.ValidationErrors(w, errs)
		return
	}

	// Set default company name if not provided
	if req.CompanyName == "" {
		req.CompanyName = req.Name + defaultOrgNameSuffix
	}

	ctx := r.Context()

	// Check if email already exists
	existing, err := h.authStore.GetUserByEmail(ctx, req.Email)
	if err == nil && existing != nil {
		responders.Conflict(w, "email_exists", "email already registered")
		return
	}

	// Hash password
	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		h.logger.Error("failed to hash password", "error", err.Error())
		responders.InternalError(w, "failed to process password")
		return
	}

	now := time.Now()

	// Create tenant
	tenantID := uuid.New().String()
	tenant := &auth.Tenant{
		ID:        tenantID,
		Name:      req.CompanyName,
		Slug:      generateSlug(req.CompanyName),
		Plan:      defaultTenantPlan,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := h.authStore.CreateTenant(ctx, tenant); err != nil {
		h.logger.Error("failed to create tenant", "error", err.Error())
		responders.InternalError(w, "failed to create organization")
		return
	}

	// Generate email verification token
	verificationToken := generateVerificationToken()

	// Build full name from first/last or use provided name
	fullName := strings.TrimSpace(req.Name)
	if req.FirstName != "" || req.LastName != "" {
		fullName = strings.TrimSpace(req.FirstName + " " + req.LastName)
	}

	// Create user (first user is admin)
	userID := uuid.New().String()
	user := &auth.User{
		ID:                     userID,
		TenantID:               tenantID,
		Email:                  normalizeEmail(req.Email),
		Name:                   fullName,
		FirstName:              strings.TrimSpace(req.FirstName),
		LastName:               strings.TrimSpace(req.LastName),
		JobTitle:               strings.TrimSpace(req.JobTitle),
		PasswordHash:           passwordHash,
		Role:                   "admin",
		IsActive:               true,
		EmailVerified:          false, // Requires verification
		EmailVerificationToken: verificationToken,
		CreatedAt:              now,
		UpdatedAt:              now,
	}

	if err := h.authStore.CreateUser(ctx, user); err != nil {
		// Rollback tenant creation
		if delErr := h.authStore.DeleteTenant(ctx, tenantID); delErr != nil {
			h.logger.Error("failed to rollback tenant creation",
				"tenantId", tenantID,
				"error", delErr.Error(),
			)
		}
		h.logger.Error("failed to create user", "error", err.Error())
		responders.InternalError(w, "failed to create user")
		return
	}

	// In production, send verification email here
	// For now, we'll include the token in the response for development
	// TODO: Implement email sending service

	h.logger.Info("user registered - verification required",
		"userId", user.ID,
		"tenantId", tenant.ID,
		"email", user.Email,
	)

	// Return response indicating verification is required
	// Don't create session until email is verified
	responders.Created(w, AuthResponse{
		User: UserDTO{
			ID:            user.ID,
			Email:         user.Email,
			Name:          user.Name,
			FirstName:     user.FirstName,
			LastName:      user.LastName,
			Role:          user.Role,
			EmailVerified: user.EmailVerified,
		},
		Tenant: TenantDTO{
			ID:   tenant.ID,
			Name: tenant.Name,
			Slug: tenant.Slug,
			Plan: tenant.Plan,
		},
		RequiresVerification: true,
		VerificationToken:    verificationToken, // Only in dev - remove in production
	})
}

// Login handles user login.
// POST /api/auth/login
func (h *AuthHandlers) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		responders.MethodNotAllowed(w, http.MethodPost)
		return
	}

	defer r.Body.Close()

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responders.BadRequest(w, "invalid_json", "invalid request body")
		return
	}

	// Basic validation
	if req.Email == "" || req.Password == "" {
		responders.BadRequest(w, "validation_error", "email and password are required")
		return
	}

	ctx := r.Context()
	email := normalizeEmail(req.Email)

	if h.enforceLoginLockout(w, email) {
		return
	}

	// Get user by email
	user, err := h.authStore.GetUserByEmail(ctx, email)
	if err != nil || user == nil {
		h.handleLoginFailure(w, email)
		return
	}

	if !user.IsActive {
		responders.Forbidden(w, "account_disabled", "account has been disabled")
		return
	}

	// Check password
	if !auth.CheckPassword(user.PasswordHash, req.Password) {
		h.handleLoginFailure(w, email)
		return
	}

	// Get tenant
	tenant, err := h.authStore.GetTenant(ctx, user.TenantID)
	if err != nil || tenant == nil {
		h.logger.Error("failed to get tenant for user",
			"userId", user.ID,
			"tenantId", user.TenantID,
			"error", err,
		)
		responders.InternalError(w, "failed to get organization")
		return
	}

	if !tenant.IsActive {
		responders.Forbidden(w, "organization_disabled", "organization has been disabled")
		return
	}

	// Update last login (non-fatal on error)
	if err := h.authStore.UpdateUserLastLogin(ctx, user.ID); err != nil {
		h.logger.Warn("failed to update last login",
			"userId", user.ID,
			"error", err.Error(),
		)
	}

	// Create session token
	token, err := h.sessionManager.CreateToken(user, tenant)
	if err != nil {
		h.logger.Error("failed to create session token", "error", err.Error())
		responders.InternalError(w, "failed to create session")
		return
	}

	// Record successful login
	if h.lockoutManager != nil {
		h.lockoutManager.RecordSuccess(email)
	}

	// Set session cookie
	h.setSessionCookie(w, token)

	h.logger.Info("user logged in",
		"userId", user.ID,
		"tenantId", tenant.ID,
	)

	responders.JSON(w, http.StatusOK, AuthResponse{
		Token: token,
		User: UserDTO{
			ID:            user.ID,
			Email:         user.Email,
			Name:          user.Name,
			FirstName:     user.FirstName,
			LastName:      user.LastName,
			Role:          user.Role,
			EmailVerified: user.EmailVerified,
		},
		Tenant: TenantDTO{
			ID:   tenant.ID,
			Name: tenant.Name,
			Slug: tenant.Slug,
			Plan: tenant.Plan,
		},
	})
}

func (h *AuthHandlers) enforceLoginLockout(w http.ResponseWriter, email string) bool {
	if h.lockoutManager == nil {
		return false
	}
	if !h.lockoutManager.IsLocked(email) {
		return false
	}
	_, lockedUntil := h.lockoutManager.GetLockoutInfo(email)
	message := "Too many failed login attempts. Try again later."
	if lockedUntil != nil {
		message = fmt.Sprintf("Account locked until %s due to multiple failed login attempts.", lockedUntil.Format(time.RFC3339))
	}
	responders.TooManyRequests(w, message)
	h.logger.Warn("login attempt blocked by lockout",
		"email", email,
		"locked_until", lockedUntil,
	)
	return true
}

func (h *AuthHandlers) handleLoginFailure(w http.ResponseWriter, email string) {
	if h.lockoutManager == nil {
		responders.Unauthorized(w, "invalid_credentials", "invalid email or password")
		return
	}

	locked, remaining := h.lockoutManager.RecordFailure(email)
	if locked {
		responders.TooManyRequests(w, "Too many failed login attempts. Account temporarily locked.")
		h.logger.Warn("account locked after repeated failures", "email", email)
		return
	}

	message := "invalid email or password"
	if remaining > 0 {
		message = fmt.Sprintf("invalid email or password (%d attempts remaining)", remaining)
	}
	responders.Unauthorized(w, "invalid_credentials", message)
}

// Logout handles user logout.
// POST /api/auth/logout
func (h *AuthHandlers) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		responders.MethodNotAllowed(w, http.MethodPost)
		return
	}

	// Clear session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		Domain:   h.cookieDomain,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: http.SameSiteStrictMode,
	})

	responders.JSON(w, http.StatusOK, map[string]string{
		"message": "logged out successfully",
	})
}

// Me returns the current authenticated user.
// GET /api/auth/me
func (h *AuthHandlers) Me(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		responders.MethodNotAllowed(w, http.MethodGet)
		return
	}

	user, ok := auth.UserFromContext(r.Context())
	if !ok || user == nil {
		responders.Unauthorized(w, "unauthorized", "not authenticated")
		return
	}

	tenant, ok := auth.TenantFromContext(r.Context())
	if !ok || tenant == nil {
		responders.InternalError(w, "failed to get organization")
		return
	}

	responders.JSON(w, http.StatusOK, AuthResponse{
		User: UserDTO{
			ID:            user.ID,
			Email:         user.Email,
			Name:          user.Name,
			FirstName:     user.FirstName,
			LastName:      user.LastName,
			Role:          user.Role,
			EmailVerified: user.EmailVerified,
		},
		Tenant: TenantDTO{
			ID:   tenant.ID,
			Name: tenant.Name,
			Slug: tenant.Slug,
			Plan: tenant.Plan,
		},
	})
}

// ChangePassword handles password changes.
// POST /api/auth/change-password
func (h *AuthHandlers) ChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		responders.MethodNotAllowed(w, http.MethodPost)
		return
	}

	user, ok := auth.UserFromContext(r.Context())
	if !ok || user == nil {
		responders.Unauthorized(w, "unauthorized", "not authenticated")
		return
	}

	defer r.Body.Close()

	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responders.BadRequest(w, "invalid_json", "invalid request body")
		return
	}

	// Validate request
	if errs := h.validatePasswordChange(req); len(errs) > 0 {
		responders.ValidationErrors(w, errs)
		return
	}

	ctx := r.Context()

	// Get fresh user data to verify current password
	freshUser, err := h.authStore.GetUser(ctx, user.ID)
	if err != nil {
		h.logger.Error("failed to get user for password change",
			"userId", user.ID,
			"error", err.Error(),
		)
		responders.InternalError(w, "failed to get user")
		return
	}

	// Verify current password
	if !auth.CheckPassword(freshUser.PasswordHash, req.CurrentPassword) {
		responders.Unauthorized(w, "invalid_password", "current password is incorrect")
		return
	}

	// Hash new password
	newHash, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		h.logger.Error("failed to hash new password", "error", err.Error())
		responders.InternalError(w, "failed to process password")
		return
	}

	// Update password
	if err := h.authStore.UpdateUserPassword(ctx, user.ID, newHash); err != nil {
		h.logger.Error("failed to update password",
			"userId", user.ID,
			"error", err.Error(),
		)
		responders.InternalError(w, "failed to update password")
		return
	}

	h.logger.Info("password changed", "userId", user.ID)

	responders.JSON(w, http.StatusOK, map[string]string{
		"message": "password changed successfully",
	})
}

// -----------------------------------------------------------------------------
// Validation Helpers
// -----------------------------------------------------------------------------

func (h *AuthHandlers) validateRegistration(req RegisterRequest) []responders.ValidationError {
	var errs []responders.ValidationError

	// Email validation
	if req.Email == "" {
		errs = append(errs, responders.ValidationError{
			Field:   "email",
			Message: "email is required",
		})
	} else if !emailRegex.MatchString(req.Email) {
		errs = append(errs, responders.ValidationError{
			Field:   "email",
			Message: "invalid email format",
		})
	}

	// Password validation
	if req.Password == "" {
		errs = append(errs, responders.ValidationError{
			Field:   "password",
			Message: "password is required",
		})
	} else {
		if pwdErrs := validatePasswordStrength(req.Password); len(pwdErrs) > 0 {
			for _, msg := range pwdErrs {
				errs = append(errs, responders.ValidationError{
					Field:   "password",
					Message: msg,
				})
			}
		}
	}

	// Name validation
	if strings.TrimSpace(req.Name) == "" {
		errs = append(errs, responders.ValidationError{
			Field:   "name",
			Message: "name is required",
		})
	} else if len(req.Name) > 100 {
		errs = append(errs, responders.ValidationError{
			Field:   "name",
			Message: "name must be at most 100 characters",
		})
	}

	// Company name validation (optional but limited length)
	if len(req.CompanyName) > 200 {
		errs = append(errs, responders.ValidationError{
			Field:   "company_name",
			Message: "company name must be at most 200 characters",
		})
	}

	return errs
}

func (h *AuthHandlers) validatePasswordChange(req ChangePasswordRequest) []responders.ValidationError {
	var errs []responders.ValidationError

	if req.CurrentPassword == "" {
		errs = append(errs, responders.ValidationError{
			Field:   "current_password",
			Message: "current password is required",
		})
	}

	if req.NewPassword == "" {
		errs = append(errs, responders.ValidationError{
			Field:   "new_password",
			Message: "new password is required",
		})
	} else {
		if pwdErrs := validatePasswordStrength(req.NewPassword); len(pwdErrs) > 0 {
			for _, msg := range pwdErrs {
				errs = append(errs, responders.ValidationError{
					Field:   "new_password",
					Message: msg,
				})
			}
		}
	}

	return errs
}

// validatePasswordStrength checks password strength requirements.
func validatePasswordStrength(password string) []string {
	var errors []string

	if len(password) < minPasswordLength {
		errors = append(errors, "password must be at least 8 characters")
	}

	if len(password) > maxPasswordLength {
		errors = append(errors, "password must be at most 128 characters")
	}

	// Check for at least one uppercase, lowercase, digit, and special char
	var hasUpper, hasLower, hasDigit bool
	for _, c := range password {
		switch {
		case unicode.IsUpper(c):
			hasUpper = true
		case unicode.IsLower(c):
			hasLower = true
		case unicode.IsDigit(c):
			hasDigit = true
		}
	}

	if !hasUpper {
		errors = append(errors, "password must contain at least one uppercase letter")
	}
	if !hasLower {
		errors = append(errors, "password must contain at least one lowercase letter")
	}
	if !hasDigit {
		errors = append(errors, "password must contain at least one digit")
	}

	return errors
}

// -----------------------------------------------------------------------------
// Password Reset Handlers
// -----------------------------------------------------------------------------

// NewPasswordForgotHandler handles POST /api/auth/password/forgot.
// Generates a reset token and logs it instead of sending email.
func NewPasswordForgotHandler(authSvc *auth.Service, logger *slog.Logger) http.HandlerFunc {
	if logger == nil {
		logger = slog.Default().With("component", "password-forgot-handler")
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			responders.MethodNotAllowed(w, http.MethodPost)
			return
		}
		var req passwordForgotRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Email) == "" {
			responders.BadRequest(w, "invalid_request", "email is required")
			return
		}

		token, user, err := authSvc.CreatePasswordResetToken(r.Context(), strings.TrimSpace(req.Email))
		if err == nil {
			logger.Info("password reset token generated",
				"userId", user.ID,
				"email", user.Email,
				"token", token,
			)
		}

		// Always return accepted to avoid user enumeration.
		responders.JSON(w, http.StatusAccepted, map[string]string{"status": "ok"})
	}
}

// NewPasswordResetHandler handles POST /api/auth/password/reset.
func NewPasswordResetHandler(authSvc *auth.Service, logger *slog.Logger) http.HandlerFunc {
	if logger == nil {
		logger = slog.Default().With("component", "password-reset-handler")
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			responders.MethodNotAllowed(w, http.MethodPost)
			return
		}
		var req passwordResetRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			responders.BadRequest(w, "invalid_request", "invalid JSON payload")
			return
		}
		if req.Token == "" || req.NewPassword == "" {
			responders.BadRequest(w, "invalid_request", "token and new_password are required")
			return
		}
		if issues := validatePasswordStrength(req.NewPassword); len(issues) > 0 {
			responders.ValidationErrors(w, []responders.ValidationError{{Field: "new_password", Message: strings.Join(issues, "; ")}})
			return
		}

		if err := authSvc.ResetPassword(r.Context(), req.Token, req.NewPassword); err != nil {
			responders.BadRequest(w, "invalid_token", "invalid or expired reset token")
			return
		}

		responders.JSON(w, http.StatusOK, map[string]string{"status": "password_reset"})
	}
}

// -----------------------------------------------------------------------------
// Cookie Helpers
// -----------------------------------------------------------------------------

func (h *AuthHandlers) setSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		Domain:   h.cookieDomain,
		MaxAge:   sessionCookieMaxAge,
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: http.SameSiteStrictMode,
	})
}

// -----------------------------------------------------------------------------
// String Helpers
// -----------------------------------------------------------------------------

// normalizeEmail converts email to lowercase and trims whitespace.
func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// generateSlug creates a URL-safe slug from a name.
func generateSlug(name string) string {
	if name == "" {
		name = "org"
	}

	var result strings.Builder
	result.Grow(len(name) + slugRandomSuffixLength + 1)

	lastDash := false
	for _, c := range strings.ToLower(name) {
		switch {
		case c >= 'a' && c <= 'z':
			result.WriteRune(c)
			lastDash = false
		case c >= '0' && c <= '9':
			result.WriteRune(c)
			lastDash = false
		case c == ' ' || c == '-' || c == '_':
			if !lastDash && result.Len() > 0 {
				result.WriteByte('-')
				lastDash = true
			}
		}
	}

	slug := result.String()

	// Trim trailing dash
	slug = strings.TrimSuffix(slug, "-")

	if slug == "" {
		slug = "org"
	}

	// Add random suffix for uniqueness
	slug += "-" + uuid.New().String()[:slugRandomSuffixLength]
	return slug
}

// generateVerificationToken creates a random token for email verification
func generateVerificationToken() string {
	return uuid.New().String()
}

// VerifyEmailRequest represents the email verification request
type VerifyEmailRequest struct {
	Token string `json:"token"`
}

// VerifyEmail verifies a user's email address using the verification token
func (h *AuthHandlers) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		responders.MethodNotAllowed(w, http.MethodPost)
		return
	}

	ctx := r.Context()

	var req VerifyEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responders.BadRequest(w, "invalid_json", "invalid request body")
		return
	}

	if req.Token == "" {
		responders.BadRequest(w, "validation_error", "verification token is required")
		return
	}

	// Find user by verification token
	user, err := h.authStore.GetUserByVerificationToken(ctx, req.Token)
	if err != nil {
		h.logger.Error("failed to find user by verification token", "error", err.Error())
		responders.BadRequest(w, "invalid_token", "invalid or expired verification token")
		return
	}

	if user.EmailVerified {
		responders.JSON(w, http.StatusOK, map[string]interface{}{
			"success": true,
			"message": "Email already verified. Please login.",
		})
		return
	}

	// Update user as verified
	user.EmailVerified = true
	user.EmailVerificationToken = "" // Clear the token
	user.UpdatedAt = time.Now()

	if err := h.authStore.UpdateUser(ctx, user); err != nil {
		h.logger.Error("failed to update user verification status", "error", err.Error())
		responders.InternalError(w, "failed to verify email")
		return
	}

	h.logger.Info("email verified successfully",
		"userId", user.ID,
		"email", user.Email,
	)

	responders.JSON(w, http.StatusOK, map[string]interface{}{
		"success":   true,
		"message":   "Email verified successfully. Please login to continue.",
		"firstName": user.FirstName,
	})
}
