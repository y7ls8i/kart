package mongo

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// DefaultPerPage is the default number of products per page.
const DefaultPerPage = 20

// Client represents a mongo client.
type Client struct {
	client *mongo.Client
	db     string
}

// NewClient returns a new mongo client.
func NewClient(uri, db string) (*Client, error) {
	mc, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("error connecting to mongo: %w", err)
	}

	client := &Client{client: mc, db: db}

	if err := client.EnsureIndexes(); err != nil {
		return nil, err
	}

	return client, nil
}

// EnsureIndexes ensures that the indexes are created in DB.
func (c *Client) EnsureIndexes() error {
	{
		coll := c.client.Database(c.db).Collection("products")
		if _, err := coll.Indexes().CreateOne(
			context.Background(),
			mongo.IndexModel{
				Keys: bson.D{{Key: "name", Value: 1}},
			}); err != nil {
			return fmt.Errorf("error creating products index: %w", err)
		}
	}

	{
		coll := c.client.Database(c.db).Collection("coupons")
		if _, err := coll.Indexes().CreateOne(
			context.Background(),
			mongo.IndexModel{
				Keys:    bson.D{{Key: "code", Value: 1}},
				Options: options.Index().SetUnique(true),
			}); err != nil {
			return fmt.Errorf("error creating coupons index: %w", err)
		}
	}

	return nil
}
