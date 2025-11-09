// internal/handler/video_handler.go
package handler

import (
	"net/http"

	commontypes "jcloud-project/libs/go-common/types/jwt"
	"jcloud-project/video-service/internal/service"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type VideoHandler struct {
	service service.VideoService
}

func NewVideoHandler(s service.VideoService) *VideoHandler {
	return &VideoHandler{service: s}
}

func (h *VideoHandler) UploadVideo(c echo.Context) error {
	// Extract user ID from JWT token
	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*commontypes.JwtCustomClaims)
	userID := claims.UserID

	// Extract metadata from form values
	title := c.FormValue("title")
	if title == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "title is required"})
	}
	description := c.FormValue("description")

	// Extract file from form
	file, err := c.FormFile("videoFile")
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "video file is required"})
	}

	video, err := h.service.UploadVideo(c.Request().Context(), userID, title, description, file)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "could not process video upload"})
	}

	return c.JSON(http.StatusCreated, video)
}
