// Package business contains the business layer. Here lives the business logic.
// It orchestrates data from and to the database and the server API.
package business

import (
	"context"
	"errors"
	"fmt"

	"github.com/y7ls8i/kart/adapter/mongo"
)

// DB is the interface for the database layer that is required by the business layer.
type DB interface {
	CreateOrder(ctx context.Context, items []mongo.ItemRequest) (*mongo.Order, error)
	FindProducts(ctx context.Context, ids []string) (missing []string, products []mongo.Product, err error)
	FindOneCoupon(ctx context.Context, code string) (coupon *mongo.Coupon, err error)
}

// Business struct represents the business layer object.
type Business struct {
	db DB
}

// NewBusiness returns a new business layer object.
func NewBusiness(db DB) *Business {
	return &Business{db: db}
}

// Order represents an order.
type Order struct {
	*mongo.Order
	Products []mongo.Product `json:"products"`
}

// OrderRequest represents the order request.
type OrderRequest struct {
	Items      []mongo.ItemRequest `json:"items"`
	CouponCode string              `json:"couponCode"`
}

// CreateOrder creates a new order.
func (b *Business) CreateOrder(ctx context.Context, req OrderRequest) (result *Order, err error) {
	// 1. check quantity
	productIDs := make([]string, 0, len(req.Items))
	for _, item := range req.Items {
		if item.Quantity <= 0 {
			return nil, fmt.Errorf("%w: quantity must be positive", mongo.ErrUnprocessableEntity)
		}
		productIDs = append(productIDs, item.ProductID)
	}

	// 2. check if products exist
	missing, products, err := b.db.FindProducts(ctx, productIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to find products: %w", err)
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("%w: product ids not found: %v", mongo.ErrUnprocessableEntity, missing)
	}

	// 3. check if the coupon exists
	if req.CouponCode != "" {
		if _, err := b.db.FindOneCoupon(ctx, req.CouponCode); err != nil {
			if errors.Is(err, mongo.ErrNotFound) {
				return nil, fmt.Errorf("%w: coupon %q not found", mongo.ErrUnprocessableEntity, req.CouponCode)
			}
			return nil, fmt.Errorf("failed to find one coupon: %w", err)
		}
	}

	orderCreated, err := b.db.CreateOrder(ctx, req.Items)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	result = &Order{
		Order:    orderCreated,
		Products: products,
	}
	return result, nil
}
