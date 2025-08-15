package mongo

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testMongoURI = "mongodb://127.0.0.1:27017"

func newTestClient(t *testing.T) *Client {
	t.Helper()

	ts := time.Now().UnixNano()
	dbName := fmt.Sprintf("kart-%d", ts)

	c, err := NewClient(testMongoURI, dbName)
	if err != nil {
		t.Skipf("skipping: cannot connect to MongoDB at %s: %v", testMongoURI, err)
		return nil
	}

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = c.client.Database(dbName).Drop(ctx)
		_ = c.client.Disconnect(ctx)
	})

	return c
}

func TestInsertCoupons(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()

		c := newTestClient(t)
		if c == nil {
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := c.InsertCoupons(ctx, []Coupon{
			{Code: "SAVE10"},
			{Code: "SAVE20"},
		})
		require.NoError(t, err)
	})
}

func TestFindOneCoupon(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()

		c := newTestClient(t)
		if c == nil {
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		code := "FIND_OK"

		err := c.InsertCoupons(ctx, []Coupon{{Code: code}})
		require.NoError(t, err)

		got, err := c.FindOneCoupon(ctx, code)
		require.NoError(t, err)
		assert.NotNil(t, got)
		assert.Equal(t, code, got.Code)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()

		c := newTestClient(t)
		if c == nil {
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		got, err := c.FindOneCoupon(ctx, "DOES_NOT_EXIST")
		assert.Error(t, err)
		assert.Equal(t, "not found", err.Error())
		assert.True(t, errors.Is(err, ErrNotFound))
		assert.Nil(t, got)
	})
}
