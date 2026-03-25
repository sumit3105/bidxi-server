package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"bidxi/internal/dto"
	"bidxi/internal/service"
	"bidxi/internal/utils"
)

// AuthHandler handles HTTP requests for authentication endpoints.
type AuthHandler struct {
	authService service.AuthService
}

// NewAuthHandler creates a new AuthHandler with the given auth service.
func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// RegisterRoutes registers auth-related routes on the given router group.
func (h *AuthHandler) RegisterRoutes(rg *gin.RouterGroup) {
	auth := rg.Group("/auth")
	{
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
	}
}

// RegisterProtectedRoutes registers auth routes that require JWT authentication.
func (h *AuthHandler) RegisterProtectedRoutes(rg *gin.RouterGroup) {
	auth := rg.Group("/auth")
	{
		auth.GET("/me", h.GetCurrentUser)
	}
}

// Register handles POST /api/auth/register
// @Summary      Register a new user
// @Description  Creates a new user account and returns a JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body  dto.RegisterRequest  true  "Registration payload"
// @Success      201   {object}  dto.AuthResponse
// @Failure      400   {object}  dto.ErrorResponse
// @Failure      409   {object}  dto.ErrorResponse
// @Router       /api/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, "Invalid request body", err)
		return
	}

	response, status, err := h.authService.Register(c.Request.Context(), &req)
	if err != nil {
		utils.RespondError(c, status, err.Error())
		return
	}

	utils.RespondSuccess(c, status, "User registered successfully", response)
}

// Login handles POST /api/auth/login
// @Summary      Login user
// @Description  Authenticates user and returns a JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body  dto.LoginRequest  true  "Login payload"
// @Success      200   {object}  dto.AuthResponse
// @Failure      400   {object}  dto.ErrorResponse
// @Failure      401   {object}  dto.ErrorResponse
// @Router       /api/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, "Invalid request body", err)
		return
	}

	response, status, err := h.authService.Login(c.Request.Context(), &req)
	if err != nil {
		utils.RespondError(c, status, err.Error())
		return
	}

	utils.RespondSuccess(c, status, "Login successful", response)
}

// GetCurrentUser handles GET /api/auth/me
// @Summary      Get current user
// @Description  Returns the authenticated user's profile data
// @Tags         auth
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  dto.UserDTO
// @Failure      401  {object}  dto.ErrorResponse
// @Router       /api/auth/me [get]
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	user, status, err := h.authService.GetCurrentUser(c.Request.Context(), userID.(string))
	if err != nil {
		utils.RespondError(c, status, err.Error())
		return
	}

	utils.RespondSuccess(c, status, "User retrieved successfully", user)
}
