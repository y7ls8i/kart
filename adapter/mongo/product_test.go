package mongo

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func insertTestProducts(t *testing.T, c *Client, products []Product) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	coll := c.client.Database(c.db).Collection(CollectionNameProducts)
	docs := make([]any, 0, len(products))
	for i := range products {
		docs = append(docs, products[i])
	}
	if len(docs) == 0 {
		return
	}

	_, err := coll.InsertMany(ctx, docs)
	require.NoError(t, err)
}

func TestListProducts(t *testing.T) {
	t.Run("pagination", func(t *testing.T) {
		t.Parallel()

		c := newTestClient(t)
		if c == nil {
			return
		}

		// Prepare 25 products
		var products []Product
		for i := 0; i < 25; i++ {
			products = append(products, Product{
				ID:       bson.NewObjectID(),
				Category: "cat",
				Name:     "p-" + bson.NewObjectID().Hex(), // ensure unique names
				Price:    float64(100 + i),
			})
		}
		insertTestProducts(t, c, products)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Page 1
		page1, err := c.ListProducts(ctx, 1)
		require.NoError(t, err)
		require.Len(t, page1, DefaultPerPage)

		// Page 2
		page2, err := c.ListProducts(ctx, 2)
		require.NoError(t, err)
		require.Len(t, page2, 25-DefaultPerPage)

		// Basic sanity: no duplicates across pages and total unique equals 25.
		seen := map[string]struct{}{}
		for _, p := range page1 {
			seen[p.ID.Hex()] = struct{}{}
		}
		for _, p := range page2 {
			_, dup := seen[p.ID.Hex()]
			assert.False(t, dup, "product appeared in both page1 and page2")
			seen[p.ID.Hex()] = struct{}{}
		}
		assert.Len(t, seen, 25)
	})

	t.Run("page less than 1", func(t *testing.T) {
		t.Parallel()

		c := newTestClient(t)
		if c == nil {
			return
		}

		// Prepare a few products
		products := []Product{
			{ID: bson.NewObjectID(), Category: "c", Name: "n1", Price: 1},
			{ID: bson.NewObjectID(), Category: "c", Name: "n2", Price: 2},
		}
		insertTestProducts(t, c, products)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		got, err := c.ListProducts(ctx, 0) // should behave like page 1
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(got), 2)
	})
}

func TestGetProduct(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()

		c := newTestClient(t)
		if c == nil {
			return
		}

		p := Product{
			ID:       bson.NewObjectID(),
			Category: "electronics",
			Name:     "Headphones",
			Price:    59.99,
		}
		insertTestProducts(t, c, []Product{p})

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		got, err := c.GetProduct(ctx, p.ID.Hex())
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, p.ID, got.ID)
		assert.Equal(t, p.Category, got.Category)
		assert.Equal(t, p.Name, got.Name)
		assert.Equal(t, p.Price, got.Price)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()

		c := newTestClient(t)
		if c == nil {
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		unknownID := bson.NewObjectID().Hex()
		got, err := c.GetProduct(ctx, unknownID)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrNotFound))
		assert.Nil(t, got)
	})

	t.Run("bad request", func(t *testing.T) {
		t.Parallel()

		c := newTestClient(t)
		if c == nil {
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		got, err := c.GetProduct(ctx, "not-a-hex")
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrBadRequest))
		assert.Nil(t, got)
	})
}

func TestFindProducts(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()

		c := newTestClient(t)
		if c == nil {
			return
		}

		id1 := bson.NewObjectID()
		id2 := bson.NewObjectID() // will be missing
		id3 := bson.NewObjectID()

		// Insert id1 and id3
		insertTestProducts(t, c, []Product{
			{ID: id1, Category: "cat1", Name: "p1", Price: 10},
			{ID: id3, Category: "cat3", Name: "p3", Price: 30},
		})

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		missing, products, err := c.FindProducts(ctx, []string{id1.Hex(), id2.Hex(), id3.Hex()})
		require.NoError(t, err)

		// Validate products found
		require.Len(t, products, 2)
		foundSet := map[string]struct{}{}
		for _, p := range products {
			foundSet[p.ID.Hex()] = struct{}{}
		}
		_, found1 := foundSet[id1.Hex()]
		_, found3 := foundSet[id3.Hex()]
		assert.True(t, found1)
		assert.True(t, found3)

		// Validate missing
		require.Len(t, missing, 1)
		assert.Equal(t, id2.Hex(), missing[0])
	})

	t.Run("bad request", func(t *testing.T) {
		t.Parallel()

		c := newTestClient(t)
		if c == nil {
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		missing, products, err := c.FindProducts(ctx, []string{"invalid-hex"})
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrBadRequest))
		assert.Nil(t, missing)
		assert.Nil(t, products)
	})
}
