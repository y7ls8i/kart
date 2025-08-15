package mongo

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// CollectionNameProducts is the name of the collection for products.
const CollectionNameProducts = "products"

// Product represents a product in DB.
type Product struct {
	ID       bson.ObjectID `json:"id" bson:"_id"`
	Category string        `json:"category" bson:"category"`
	Name     string        `json:"name" bson:"name"`
	Price    float64       `json:"price" bson:"price"`
}

// ListProducts returns the list of products.
func (c *Client) ListProducts(ctx context.Context, page int) ([]Product, error) {
	coll := c.client.Database(c.db).Collection(CollectionNameProducts)

	if page < 1 {
		page = 1
	}

	findOptions := options.Find()
	findOptions.SetLimit(int64(DefaultPerPage))
	findOptions.SetSkip(int64((page - 1) * DefaultPerPage))

	cursor, err := coll.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to find products: %w", err)
	}
	defer func() {
		_ = cursor.Close(ctx)
	}()

	products := []Product{} // return an empty array if no products
	if err := cursor.All(ctx, &products); err != nil {
		return nil, fmt.Errorf("failed to get all products: %w", err)
	}

	return products, nil
}

// GetProduct returns the requested product.
func (c *Client) GetProduct(ctx context.Context, id string) (*Product, error) {
	bsonID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrBadRequest, err)
	}

	coll := c.client.Database(c.db).Collection(CollectionNameProducts)

	var product Product
	if err := coll.FindOne(ctx, bson.M{"_id": bsonID}).Decode(&product); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("%w: %w", ErrNotFound, err)
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return &product, nil
}

// FindProducts returns the requested products and informs which product ids are missing.
func (c *Client) FindProducts(ctx context.Context, ids []string) (missing []string, products []Product, err error) {
	productIDs := make([]bson.ObjectID, 0, len(ids))
	for _, id := range ids {
		bsonID, err := bson.ObjectIDFromHex(id)
		if err != nil {
			return nil, nil, fmt.Errorf("%w: %w", ErrBadRequest, err)
		}
		productIDs = append(productIDs, bsonID)
	}

	coll := c.client.Database(c.db).Collection(CollectionNameProducts)

	filter := bson.M{"_id": bson.M{"$in": productIDs}}
	cursor, err := coll.Find(ctx, filter)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to find products: %w", err)
	}
	defer func() {
		_ = cursor.Close(ctx)
	}()

	products = []Product{}
	if err := cursor.All(ctx, &products); err != nil {
		return nil, nil, fmt.Errorf("failed to get products: %w", err)
	}

	found := make(map[bson.ObjectID]struct{})
	for _, p := range products {
		found[p.ID] = struct{}{}
	}

	for _, id := range productIDs {
		if _, ok := found[id]; !ok {
			missing = append(missing, id.Hex())
		}
	}

	return missing, products, nil
}
