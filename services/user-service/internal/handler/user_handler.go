// internal/handler/user_handler.go
package handler

import (
	"jcloud-project/user-service/internal/service"
	"net/http"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

//
// User Handler
//

type UserHandler struct {
	service service.UserService
}

func NewUserHandler(s service.UserService) *UserHandler {
	return &UserHandler{service: s}
}

// DTO for user registration request
type registerRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

func (h *UserHandler) Register(c echo.Context) error {
	var req registerRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request format"})
	}

	// Basic validation, can be replaced with a validator library later
	if req.Email == "" || len(req.Password) < 8 {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "email and password are required, password must be at least 8 characters"})
	}

	user, err := h.service.Register(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		// In a real app, you would check for specific errors, like duplicate email
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "could not create user"})
	}

	return c.JSON(http.StatusCreated, user)
}

// DTO for user login request
type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *UserHandler) Login(c echo.Context) error {
	var req loginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request format"})
	}

	token, err := h.service.Login(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		// For security, we return a generic "Unauthorized" error
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "invalid credentials"})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"token": token,
	})
}

func (h *UserHandler) Profile(c echo.Context) error {
	// The middleware places the parsed token in the context
	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*service.JwtCustomClaims)
	userID := claims.UserID

	user, err := h.service.GetProfile(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "user not found"})
	}

	return c.JSON(http.StatusOK, user)
}

// GetInternalUserDetails provides user details for service-to-service communication.
// It does not require JWT authentication as it's only exposed on the internal network.
func (h *UserHandler) GetInternalUserDetails(c echo.Context) error {
	userID, err := strconv.ParseInt(c.Param("userId"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid user id"})
	}

	user, err := h.service.GetProfile(c.Request().Context(), userID)
	if err != nil {
		// This could be a NotFound error, which is important to propagate
		return c.JSON(http.StatusNotFound, echo.Map{"error": "user not found"})
	}

	// Return only the necessary fields for internal clients
	return c.JSON(http.StatusOK, echo.Map{
		"id":    user.ID,
		"email": user.Email,
	})
}

//
// Admin Middleware
//

func AdminMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// The JWT middleware has already parsed the token and put it in the context.
		userToken, ok := c.Get("user").(*jwt.Token)
		if !ok {
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": "invalid token"})
		}

		claims, ok := userToken.Claims.(*service.JwtCustomClaims)
		if !ok {
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": "invalid token claims"})
		}

		if claims.Role != "ADMIN" {
			return c.JSON(http.StatusForbidden, echo.Map{"error": "access forbidden: administrator role required"})
		}

		return next(c)
	}
}

// DTO for user update request by an admin
type updateUserRequest struct {
	Email string `json:"email" validate:"required,email"`
	Role  string `json:"role" validate:"required"`
}

func (h *UserHandler) GetAllUsers(c echo.Context) error {
	users, err := h.service.GetAllUsers()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "could not retrieve users"})
	}
	return c.JSON(http.StatusOK, users)
}

func (h *UserHandler) UpdateUser(c echo.Context) error {
	userID, err := strconv.ParseInt(c.Param("userId"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid user id"})
	}

	var req updateUserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request format"})
	}
	// Basic validation
	if req.Email == "" || (req.Role != "USER" && req.Role != "ADMIN") {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "valid email and role are required"})
	}

	user, err := h.service.UpdateUser(c.Request().Context(), userID, req.Email, req.Role)
	if err != nil {
		if err.Error() == "user not found" {
			return c.JSON(http.StatusNotFound, echo.Map{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "could not update user"})
	}

	return c.JSON(http.StatusOK, user)
}
