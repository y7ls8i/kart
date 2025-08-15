// Package product contains the product requests handler.
package product

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/y7ls8i/kart/adapter/mongo"
	"github.com/y7ls8i/kart/server/sverr"
)

// DB is the interface for the database layer that is required by the product requests handler.
type DB interface {
	ListProducts(ctx context.Context, page int) ([]mongo.Product, error)
	GetProduct(ctx context.Context, id string) (*mongo.Product, error)
}

// Product struct represents the product requests handler.
type Product struct {
	db DB
}

// NewProduct returns a new product requests handler.
func NewProduct(db DB) *Product {
	return &Product{db: db}
}

// List returns a list of products.
func (p *Product) List(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.Query("page"))
	products, err := p.db.ListProducts(ctx, page)
	if err != nil {
		sverr.Abort(ctx, err, "Error listing products")
		return
	}
	ctx.JSON(http.StatusOK, products)
}

// Get returns a single product by ID.
func (p *Product) Get(ctx *gin.Context) {
	product, err := p.db.GetProduct(ctx, ctx.Param("id"))
	if err != nil {
		sverr.Abort(ctx, err, "Error getting product")
		return
	}

	ctx.JSON(http.StatusOK, product)
}
