// services/video-service/internal/handler/error_handler.go
package handler

import (
	"errors"
	"jcloud-project/libs/go-common/ierr"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
)

func CustomHTTPErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	var httpCode int
	var errMsg string

	switch {
	case errors.Is(err, ierr.ErrForbidden):
		httpCode = http.StatusForbidden
		errMsg = err.Error()
	default:
		httpCode = http.StatusInternalServerError
		errMsg = "internal server error"
		log.Printf("Unhandled internal error: %v", err)
	}

	if !c.Response().Committed {
		if jsonErr := c.JSON(httpCode, echo.Map{"error": errMsg}); jsonErr != nil {
			log.Printf("Failed to send JSON error response: %v", jsonErr)
		}
	}
}
