package mongo

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	aperr "github.com/y7ls8i/kart/error"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestCreateOrder(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()

		c := newTestClient(t)
		if c == nil {
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		pid1 := bson.NewObjectID()
		pid2 := bson.NewObjectID()

		req := []ItemRequest{
			{ProductID: pid1.Hex(), Quantity: 2},
			{ProductID: pid2.Hex(), Quantity: 5},
		}

		order, err := c.CreateOrder(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, order)

		assert.NotEqual(t, bson.ObjectID{}, order.ID)
		require.Len(t, order.Items, 2)

		assert.Equal(t, pid1, order.Items[0].ProductID)
		assert.Equal(t, 2, order.Items[0].Quantity)
		assert.Equal(t, pid2, order.Items[1].ProductID)
		assert.Equal(t, 5, order.Items[1].Quantity)

		// Verify persistence by reading it back
		coll := c.client.Database(c.db).Collection(CollectionNameOrders)
		var stored Order
		err = coll.FindOne(ctx, bson.M{"_id": order.ID}).Decode(&stored)
		require.NoError(t, err)

		assert.Equal(t, order.ID, stored.ID)
		assert.Equal(t, order.Items, stored.Items)
	})

	t.Run("bad request", func(t *testing.T) {
		t.Parallel()

		c := newTestClient(t)
		if c == nil {
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		req := []ItemRequest{
			{ProductID: "not-a-hex", Quantity: 1},
		}

		order, err := c.CreateOrder(ctx, req)
		require.Error(t, err)
		assert.True(t, errors.Is(err, aperr.ErrBadRequest))
		assert.Nil(t, order)
	})
}
