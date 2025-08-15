// This is for testing the http handlers.
// Because the handlers require gin.Context object, and we can't easily create one, so we can't directly test the
// handler functions. Therefore, we test the handlers by sending http requests to the server.
package product_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/y7ls8i/kart/adapter/mongo"
	"github.com/y7ls8i/kart/server/product"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestListProduct(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testCases := []struct {
		name           string
		pageParam      string
		mock           *mockProductDB
		expectedStatus int
		expectedBody   string
		expectedPage   int
	}{
		{
			name:      "success, no page",
			pageParam: "",
			mock: &mockProductDB{
				listProductsResult: []mongo.Product{{Name: "product1"}},
				listProductsErr:    nil,
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `[{"id":"000000000000000000000000","category":"","name":"product1","price":0}]`,
			expectedPage:   0,
		},
		{
			name:      "success, paged",
			pageParam: "3",
			mock: &mockProductDB{
				listProductsResult: []mongo.Product{{Name: "product1"}},
				listProductsErr:    nil,
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `[{"id":"000000000000000000000000","category":"","name":"product1","price":0}]`,
			expectedPage:   3,
		},
		{
			name:      "success, empty",
			pageParam: "1",
			mock: &mockProductDB{
				listProductsResult: []mongo.Product{},
				listProductsErr:    nil,
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "[]",
			expectedPage:   1,
		},
		{
			name:      "error",
			pageParam: "1",
			mock: &mockProductDB{
				listProductsResult: nil,
				listProductsErr:    errors.New("internal error"),
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "",
			expectedPage:   1,
		},
	}
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			router := gin.Default()
			handler := product.NewProduct(test.mock)
			router.GET("/api/product", handler.List)

			w := httptest.NewRecorder()
			path := "/api/product"
			if test.pageParam != "" {
				path += "?page=" + test.pageParam
			}
			router.ServeHTTP(w, httptest.NewRequest("GET", path, nil))

			require.Equal(t, test.expectedStatus, w.Code)
			assert.Equal(t, test.expectedBody, w.Body.String())
			assert.Equal(t, test.expectedPage, test.mock.page)
		})
	}
}

func TestGetProduct(t *testing.T) {
	gin.SetMode(gin.TestMode)

	productID := bson.NewObjectID()

	testCases := []struct {
		name           string
		idParam        string
		mock           *mockProductDB
		expectedStatus int
		expectedBody   string
		expectedID     string
	}{
		{
			name:    "success",
			idParam: productID.Hex(),
			mock: &mockProductDB{
				getProductResult: &mongo.Product{ID: productID, Name: "product1"},
				getProductErr:    nil,
			},
			expectedStatus: http.StatusOK,
			expectedBody:   fmt.Sprintf(`{"id":%q,"category":"","name":"product1","price":0}`, productID.Hex()),
			expectedID:     productID.Hex(),
		},
		{
			name:    "bad request",
			idParam: "badid",
			mock: &mockProductDB{
				getProductResult: nil,
				getProductErr:    fmt.Errorf("%w: something wrong", mongo.ErrBadRequest),
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "",
			expectedID:     "badid",
		},
		{
			name:    "not found",
			idParam: productID.Hex(),
			mock: &mockProductDB{
				getProductResult: nil,
				getProductErr:    fmt.Errorf("%w: something wrong", mongo.ErrNotFound),
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "",
			expectedID:     productID.Hex(),
		},
	}
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			router := gin.Default()
			handler := product.NewProduct(test.mock)
			router.GET("/api/product/:id", handler.Get)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, httptest.NewRequest("GET", "/api/product/"+test.idParam, nil))

			require.Equal(t, test.expectedStatus, w.Code)
			assert.Equal(t, test.expectedBody, w.Body.String())
			assert.Equal(t, test.expectedID, test.mock.id)
		})
	}
}

type mockProductDB struct {
	page               int
	listProductsResult []mongo.Product
	listProductsErr    error
	id                 string
	getProductResult   *mongo.Product
	getProductErr      error
}

func (m *mockProductDB) ListProducts(_ context.Context, page int) ([]mongo.Product, error) {
	m.page = page
	return m.listProductsResult, m.listProductsErr
}

func (m *mockProductDB) GetProduct(_ context.Context, id string) (*mongo.Product, error) {
	m.id = id
	return m.getProductResult, m.getProductErr
}
