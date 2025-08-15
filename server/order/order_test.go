// This is for testing the http handlers.
// Because the handlers require gin.Context object, and we can't easily create one, so we can't directly test the
// handler functions. Therefore, we test the handlers by sending http requests to the server.
package order_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/y7ls8i/kart/adapter/mongo"
	"github.com/y7ls8i/kart/business"
	"github.com/y7ls8i/kart/server/order"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestCreateOrder(t *testing.T) {
	gin.SetMode(gin.TestMode)

	productID := bson.NewObjectID()
	orderID := bson.NewObjectID()

	testCases := []struct {
		name           string
		req            any
		mock           *mockBusiness
		expectedStatus int
		expectedBody   string
		expectedReq    business.OrderRequest
	}{
		{
			name: "success",
			req:  map[string]any{"couponCode": "coupon1", "items": []map[string]any{{"productId": productID.Hex(), "quantity": 1}}},
			mock: &mockBusiness{
				createOrderResult: &business.Order{Order: &mongo.Order{ID: orderID}},
				createOrderErr:    nil,
			},
			expectedStatus: http.StatusOK,
			expectedBody:   fmt.Sprintf(`{"id":%q,"items":null,"products":null}`, orderID.Hex()),
			expectedReq:    business.OrderRequest{Items: []mongo.ItemRequest{{ProductID: productID.Hex(), Quantity: 1}}, CouponCode: "coupon1"},
		},
		{
			name: "bad request",
			req:  "notjson",
			mock: &mockBusiness{
				createOrderResult: nil,
				createOrderErr:    nil,
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "",
			expectedReq:    business.OrderRequest{},
		},
		{
			name: "internal error",
			req:  map[string]any{"couponCode": "coupon1", "items": []map[string]any{{"productId": productID.Hex(), "quantity": 1}}},
			mock: &mockBusiness{
				createOrderResult: nil,
				createOrderErr:    errors.New("internal error"),
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "",
			expectedReq:    business.OrderRequest{Items: []mongo.ItemRequest{{ProductID: productID.Hex(), Quantity: 1}}, CouponCode: "coupon1"},
		},
		{
			name: "unprocessable entity",
			req:  map[string]any{"couponCode": "invalid"},
			mock: &mockBusiness{
				createOrderResult: nil,
				createOrderErr:    fmt.Errorf("%w: coupon not found", mongo.ErrUnprocessableEntity),
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   "",
			expectedReq:    business.OrderRequest{CouponCode: "invalid"},
		},
	}
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			router := gin.Default()
			handler := order.NewOrder(test.mock)
			router.POST("/api/order", handler.Create)

			w := httptest.NewRecorder()
			jsonBytes, err := json.Marshal(test.req)
			require.NoError(t, err)
			router.ServeHTTP(w, httptest.NewRequest("POST", "/api/order", bytes.NewReader(jsonBytes)))

			require.Equal(t, test.expectedStatus, w.Code)
			assert.Equal(t, test.expectedBody, w.Body.String())
			assert.Equal(t, test.expectedReq, test.mock.req)
		})
	}
}

type mockBusiness struct {
	req               business.OrderRequest
	createOrderResult *business.Order
	createOrderErr    error
}

func (m *mockBusiness) CreateOrder(_ context.Context, req business.OrderRequest) (result *business.Order, err error) {
	m.req = req
	return m.createOrderResult, m.createOrderErr
}
