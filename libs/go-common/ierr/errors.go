// libs/go-common/ierr/errors.go
package ierr

import "errors"

// Стандартные ошибки, которые могут быть использованы в любом сервисе
// для обеспечения консистентной обработки.
var (
	ErrNotFound           = errors.New("requested resource not found")
	ErrForbidden          = errors.New("access forbidden")
	ErrConflict           = errors.New("resource conflict or duplicate")
	ErrInvalidCredentials = errors.New("invalid credentials")
)
