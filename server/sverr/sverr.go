// Package sverr provides functions for handling errors.
package sverr

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/y7ls8i/kart/adapter/mongo"
)

// Abort aborts the request with the appropriate error code.
func Abort(ctx *gin.Context, err error, log string) {
	if errors.Is(err, mongo.ErrNotFound) {
		_ = ctx.AbortWithError(http.StatusNotFound, err)
		return
	}
	if errors.Is(err, mongo.ErrBadRequest) {
		_ = ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}
	if errors.Is(err, mongo.ErrUnprocessableEntity) {
		_ = ctx.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	slog.Error(log, "error", err)
	ctx.AbortWithStatus(http.StatusInternalServerError)
}
