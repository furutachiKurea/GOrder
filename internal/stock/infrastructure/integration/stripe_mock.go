package integration

import (
	"context"

	"github.com/furutachiKurea/gorder/stock/app/dto"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// MockStripeAPI stripe API 的 mock 实现
type MockStripeAPI struct {
	apiKey string
}

func NewMockStripeAPI() *MockStripeAPI {
	key := viper.GetString("stripe-key")
	if key == "" {
		log.Panic().Msg("stripe key is empty")
	}
	return &MockStripeAPI{
		apiKey: key,
	}
}

func (s *MockStripeAPI) GetProductByID(_ context.Context, pid string) (*dto.Product, error) {
	return &dto.Product{
		PriceID: "mock_price_id_" + pid,
		Name:    "mock_product_name_" + pid,
	}, nil
}
