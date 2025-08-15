// Package order contains the order requests handler.
package order

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/y7ls8i/kart/business"
	"github.com/y7ls8i/kart/server/sverr"
)

// Business is the interface for the business layer that is required by the order requests handler.
type Business interface {
	CreateOrder(ctx context.Context, req business.OrderRequest) (result *business.Order, err error)
}

// Order struct represents the order requests handler.
type Order struct {
	buss Business
}

// NewOrder creates a new order requests handler.
func NewOrder(buss Business) *Order {
	return &Order{buss: buss}
}

// Create creates a new order.
func (o *Order) Create(ctx *gin.Context) {
	req := business.OrderRequest{}
	if err := ctx.BindJSON(&req); err != nil {
		_ = ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	order, err := o.buss.CreateOrder(ctx, req)
	if err != nil {
		sverr.Abort(ctx, err, "Error creating order")
		return
	}

	ctx.JSON(http.StatusOK, order)
}
