// Package sverr provides functions for handling errors.
package sverr

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	aperr "github.com/y7ls8i/kart/error"
)

// Abort aborts the request with the appropriate error code.
func Abort(ctx *gin.Context, err error, log string) {
	if errors.Is(err, aperr.ErrNotFound) {
		_ = ctx.AbortWithError(http.StatusNotFound, err)
		return
	}
	if errors.Is(err, aperr.ErrBadRequest) {
		_ = ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}
	if errors.Is(err, aperr.ErrUnprocessableEntity) {
		_ = ctx.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	slog.Error(log, "error", err)
	ctx.AbortWithStatus(http.StatusInternalServerError)
}
