// services/user-service/internal/handler/error_handler.go
package handler

import (
	"errors"
	"jcloud-project/libs/go-common/ierr"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
)

// CustomHTTPErrorHandler преобразует ошибки из сервисного слоя в HTTP-ответы.
func CustomHTTPErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	// Определяем HTTP статус код по типу ошибки
	var httpCode int
	var errMsg string

	switch {
	case errors.Is(err, ierr.ErrNotFound):
		httpCode = http.StatusNotFound
		errMsg = ierr.ErrNotFound.Error()
	case errors.Is(err, ierr.ErrInvalidCredentials):
		httpCode = http.StatusUnauthorized
		errMsg = "invalid email or password" // Не выдаем точную причину
	case errors.Is(err, ierr.ErrConflict):
		httpCode = http.StatusConflict
		errMsg = err.Error()
	case errors.Is(err, ierr.ErrForbidden):
		httpCode = http.StatusForbidden
		errMsg = ierr.ErrForbidden.Error()
	default:
		// Для всех остальных, непредвиденных ошибок
		httpCode = http.StatusInternalServerError
		errMsg = "internal server error"
		// Важно логировать такие ошибки для дебага
		log.Printf("Unhandled internal error: %v", err)
	}

	// Отправляем JSON-ответ
	if !c.Response().Committed {
		if jsonErr := c.JSON(httpCode, echo.Map{"error": errMsg}); jsonErr != nil {
			log.Printf("Failed to send JSON error response: %v", jsonErr)
		}
	}
}
