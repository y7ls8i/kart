package mongo

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// CollectionNameOrders is the name of the collection for orders.
const CollectionNameOrders = "orders"

// OrderItem represents an item in an order in DB.
type OrderItem struct {
	ProductID bson.ObjectID `json:"productId" bson:"productId"`
	Quantity  int           `json:"quantity" bson:"quantity"`
}

// Order represents an order in DB.
type Order struct {
	ID    bson.ObjectID `json:"id" bson:"_id"`
	Items []OrderItem   `json:"items" bson:"items"`
}

// ItemRequest represents the item requested by an order.
type ItemRequest struct {
	ProductID string `json:"productId"`
	Quantity  int    `json:"quantity"`
}

// CreateOrder creates a new order.
func (c *Client) CreateOrder(ctx context.Context, items []ItemRequest) (*Order, error) {
	coll := c.client.Database(c.db).Collection(CollectionNameOrders)

	order := Order{
		ID: bson.NewObjectID(),
	}

	for _, item := range items {
		productID, err := bson.ObjectIDFromHex(item.ProductID)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrBadRequest, err)
		}
		order.Items = append(order.Items, OrderItem{
			ProductID: productID,
			Quantity:  item.Quantity,
		})
	}

	if _, err := coll.InsertOne(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	return &order, nil
}
