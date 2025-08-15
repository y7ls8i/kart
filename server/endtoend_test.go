// This is for end-to-end tests that include sending http request, receiving http response,
// and accessing the database.
// Because these tests are expensive, therefore we only test the main happy paths.
package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/y7ls8i/kart/adapter/mongo"
	"github.com/y7ls8i/kart/business"
	"github.com/y7ls8i/kart/config"
	"github.com/y7ls8i/kart/server"
	"go.mongodb.org/mongo-driver/v2/bson"
	mdriver "go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	ts          = time.Now().UnixMicro()
	productID   = bson.NewObjectID()
	productName = fmt.Sprintf("product-%d", ts)
	couponCode  = fmt.Sprintf("coupon-%d", ts)

	dbName   = fmt.Sprintf("kart-%d", ts)
	mongoURI = "mongodb://127.0.0.1:27017"
)

func getFreePort(t *testing.T) int {
	t.Helper()

	// Ask the OS to assign an available port by binding to ":0"
	listener, err := net.Listen("tcp", ":0")
	require.NoError(t, err)

	defer func() {
		_ = listener.Close()
	}()

	addr := listener.Addr().(*net.TCPAddr)
	return addr.Port
}

func setupTestData(t *testing.T) {
	t.Helper()

	mc, err := mdriver.Connect(options.Client().ApplyURI(mongoURI))
	require.NoError(t, err)

	_, err = mc.Database(dbName).Collection(mongo.CollectionNameProducts).InsertOne(context.Background(), bson.M{"_id": productID, "name": productName})
	require.NoError(t, err)

	_, err = mc.Database(dbName).Collection(mongo.CollectionNameCoupons).InsertOne(context.Background(), bson.M{"code": couponCode})
	require.NoError(t, err)

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = mc.Database(dbName).Drop(ctx)
		_ = mc.Disconnect(ctx)
	})
}

func TestEndToEnd(t *testing.T) {
	setupTestData(t)

	client, err := mongo.NewClient(mongoURI, dbName)
	require.NoError(t, err)

	buss := business.NewBusiness(client)

	port := getFreePort(t)
	t.Logf("Listening on port %d", port)
	s := server.NewServer(config.Server{Mode: "test", Listen: fmt.Sprintf(":%d", port)}, client, buss)

	ctx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()
	go func() {
		s.Start(ctx)
	}()

	t.Run("GET /api/product", func(t *testing.T) {
		req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:%d/api/product", port), nil)
		require.NoError(t, err)
		req.Header.Set("Api_key", "apitest")
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		defer func() {
			_ = resp.Body.Close()
		}()
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Contains(t, string(body), fmt.Sprintf(`{"id":%q,"category":"","name":%q,"price":0}`, productID.Hex(), productName))
	})

	t.Run("GET /api/product/:id", func(t *testing.T) {
		req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:%d/api/product/%s", port, productID.Hex()), nil)
		require.NoError(t, err)
		req.Header.Set("Api_key", "apitest")
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		defer func() {
			_ = resp.Body.Close()
		}()
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Equal(t, fmt.Sprintf(`{"id":%q,"category":"","name":%q,"price":0}`, productID.Hex(), productName), string(body))
	})

	t.Run("POST /api/order", func(t *testing.T) {
		req := map[string]any{
			"items": []map[string]any{
				{
					"productID": productID.Hex(),
					"quantity":  1,
				},
			},
			"couponCode": couponCode,
		}
		jsonBytes, err := json.Marshal(req)
		require.NoError(t, err)
		httpreq, err := http.NewRequest("POST", fmt.Sprintf("http://localhost:%d/api/order", port), bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		httpreq.Header.Set("Api_key", "apitest")
		httpreq.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(httpreq)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		defer func() {
			_ = resp.Body.Close()
		}()
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Contains(t, string(body), fmt.Sprintf(`"items":[{"productId":%q,"quantity":1}]`, productID.Hex()))
		assert.Contains(t, string(body), fmt.Sprintf(`"products":[{"id":%q,"category":"","name":%q,"price":0}]}`, productID.Hex(), productName))
	})
}
