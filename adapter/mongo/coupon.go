package mongo

import (
	"context"
	"errors"
	"fmt"

	aperr "github.com/y7ls8i/kart/error"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// CollectionNameCoupons is the name of the collection for coupons.
const CollectionNameCoupons = "coupons"

// Coupon represents a coupon in DB.
type Coupon struct {
	Code string `json:"code" bson:"code"`
}

// InsertCoupons inserts the coupons into DB.
func (c *Client) InsertCoupons(ctx context.Context, coupons []Coupon) error {
	coll := c.client.Database(c.db).Collection(CollectionNameCoupons)
	if _, err := coll.InsertMany(ctx, coupons); err != nil {
		return fmt.Errorf("failed to insert coupons: %w", err)
	}
	return nil
}

// FindOneCoupon finds the requested coupon in DB and returns the coupon.
func (c *Client) FindOneCoupon(ctx context.Context, code string) (result *Coupon, err error) {
	result = &Coupon{}
	coll := c.client.Database(c.db).Collection(CollectionNameCoupons)
	if err := coll.FindOne(ctx, map[string]string{"code": code}).Decode(result); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, aperr.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get coupon: %w", err)
	}
	return result, nil
}
