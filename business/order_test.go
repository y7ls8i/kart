package business_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/y7ls8i/kart/adapter/mongo"
	"github.com/y7ls8i/kart/business"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestBusiness_CreateOrder(t *testing.T) {
	productID := bson.NewObjectID()
	orderID := bson.NewObjectID()

	testCases := []struct {
		name           string
		req            business.OrderRequest
		mock           *mockDB
		expectedResult *business.Order
		expectedErr    error
		expectedErrIs  error
	}{
		{
			name: "success",
			req:  business.OrderRequest{CouponCode: "coupon1", Items: []mongo.ItemRequest{{ProductID: productID.Hex(), Quantity: 1}}},
			mock: &mockDB{
				findProductsMissing:  []string{},
				findProductsProducts: []mongo.Product{{ID: productID, Name: "product1"}},
				findProductsErr:      nil,
				findOneCouponCoupon:  &mongo.Coupon{Code: "coupon1"},
				findOneCouponErr:     nil,
				createOrderResult:    &mongo.Order{ID: orderID},
				createOrderErr:       nil,
			},
			expectedResult: &business.Order{Order: &mongo.Order{ID: orderID}, Products: []mongo.Product{{ID: productID, Name: "product1"}}},
			expectedErr:    nil,
			expectedErrIs:  nil,
		},
		{
			name:           "invalid quantity",
			req:            business.OrderRequest{CouponCode: "coupon1", Items: []mongo.ItemRequest{{ProductID: productID.Hex(), Quantity: -1}}},
			mock:           &mockDB{},
			expectedResult: nil,
			expectedErr:    errors.New("unprocessable entity: quantity must be positive"),
			expectedErrIs:  mongo.ErrUnprocessableEntity,
		},
		{
			name: "invalid product",
			req:  business.OrderRequest{CouponCode: "coupon1", Items: []mongo.ItemRequest{{ProductID: productID.Hex(), Quantity: 1}}},
			mock: &mockDB{
				findProductsMissing:  []string{productID.Hex()},
				findProductsProducts: []mongo.Product{},
				findProductsErr:      nil,
			},
			expectedResult: nil,
			expectedErr:    fmt.Errorf("unprocessable entity: product ids not found: [%s]", productID.Hex()),
			expectedErrIs:  mongo.ErrUnprocessableEntity,
		},
		{
			name: "invalid coupon",
			req:  business.OrderRequest{CouponCode: "coupon1", Items: []mongo.ItemRequest{{ProductID: productID.Hex(), Quantity: 1}}},
			mock: &mockDB{
				findProductsMissing:  []string{},
				findProductsProducts: []mongo.Product{{ID: productID, Name: "product1"}},
				findProductsErr:      nil,
				findOneCouponCoupon:  nil,
				findOneCouponErr:     mongo.ErrNotFound,
			},
			expectedResult: nil,
			expectedErr:    fmt.Errorf(`unprocessable entity: coupon "coupon1" not found`),
			expectedErrIs:  mongo.ErrUnprocessableEntity,
		},
		{
			name: "find products internal error",
			req:  business.OrderRequest{CouponCode: "coupon1", Items: []mongo.ItemRequest{{ProductID: productID.Hex(), Quantity: 1}}},
			mock: &mockDB{
				findProductsMissing:  nil,
				findProductsProducts: nil,
				findProductsErr:      errors.New("internal error"),
			},
			expectedResult: nil,
			expectedErr:    errors.New("failed to find products: internal error"),
			expectedErrIs:  nil,
		},
		{
			name: "find one coupon internal error",
			req:  business.OrderRequest{CouponCode: "coupon1", Items: []mongo.ItemRequest{{ProductID: productID.Hex(), Quantity: 1}}},
			mock: &mockDB{
				findProductsMissing:  []string{},
				findProductsProducts: []mongo.Product{{ID: productID, Name: "product1"}},
				findProductsErr:      nil,
				findOneCouponCoupon:  nil,
				findOneCouponErr:     errors.New("internal error"),
			},
			expectedResult: nil,
			expectedErr:    errors.New("failed to find one coupon: internal error"),
			expectedErrIs:  nil,
		},
		{
			name: "create order internal error",
			req:  business.OrderRequest{CouponCode: "coupon1", Items: []mongo.ItemRequest{{ProductID: productID.Hex(), Quantity: 1}}},
			mock: &mockDB{
				findProductsMissing:  []string{},
				findProductsProducts: []mongo.Product{{ID: productID, Name: "product1"}},
				findProductsErr:      nil,
				findOneCouponCoupon:  &mongo.Coupon{Code: "coupon1"},
				findOneCouponErr:     nil,
				createOrderResult:    nil,
				createOrderErr:       errors.New("internal error"),
			},
			expectedResult: nil,
			expectedErr:    errors.New("failed to create order: internal error"),
			expectedErrIs:  nil,
		},
	}
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			b := business.NewBusiness(test.mock)
			result, err := b.CreateOrder(context.Background(), test.req)
			if test.expectedErr != nil {
				require.Error(t, err)
				assert.Equal(t, test.expectedErr.Error(), err.Error())
				if test.expectedErrIs != nil {
					assert.True(t, errors.Is(err, test.expectedErrIs))
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expectedResult, result)
			}
		})
	}
}

type mockDB struct {
	// CreateOrder
	items             []mongo.ItemRequest
	createOrderResult *mongo.Order
	createOrderErr    error

	// FindProducts
	ids                  []string
	findProductsMissing  []string
	findProductsProducts []mongo.Product
	findProductsErr      error

	// FindOneCoupon
	code                string
	findOneCouponCoupon *mongo.Coupon
	findOneCouponErr    error
}

func (m *mockDB) CreateOrder(_ context.Context, items []mongo.ItemRequest) (*mongo.Order, error) {
	m.items = items
	return m.createOrderResult, m.createOrderErr
}

func (m *mockDB) FindProducts(_ context.Context, ids []string) (missing []string, products []mongo.Product, err error) {
	m.ids = ids
	return m.findProductsMissing, m.findProductsProducts, m.findProductsErr
}

func (m *mockDB) FindOneCoupon(_ context.Context, code string) (coupon *mongo.Coupon, err error) {
	m.code = code
	return m.findOneCouponCoupon, m.findOneCouponErr
}
