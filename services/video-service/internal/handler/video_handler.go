// services/video-service/internal/handler/video_handler.go
package handler

import (
	commontypes "jcloud-project/libs/go-common/types/jwt"
	"jcloud-project/video-service/internal/service"
	"net/http"

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
	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*commontypes.JwtCustomClaims)
	// userID := claims.UserID

	title := c.FormValue("title")
	if title == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "title is required"})
	}
	description := c.FormValue("description")

	file, err := c.FormFile("videoFile")
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "video file is required"})
	}

	video, err := h.service.ProcessNewVideoUpload(c.Request().Context(), claims, title, description, file)
	if err != nil {
		// Передаем ошибку в центральный обработчик
		return err
	}

	return c.JSON(http.StatusCreated, video)
}
