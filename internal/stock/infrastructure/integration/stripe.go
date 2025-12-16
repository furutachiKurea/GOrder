package integration

import (
	"context"

	_ "github.com/furutachiKurea/gorder/common/config"
	"github.com/furutachiKurea/gorder/stock/app/dto"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/stripe/stripe-go/v84"
	"github.com/stripe/stripe-go/v84/product"
)

type StripeAPI struct {
	apiKey string
}

func NewStripeAPI() *StripeAPI {
	key := viper.GetString("stripe-key")
	if key == "" {
		log.Panic().Msg("stripe key is empty")
	}
	return &StripeAPI{
		apiKey: key,
	}
}

func (s *StripeAPI) GetProductByID(_ context.Context, pid string) (*dto.Product, error) {
	stripe.Key = s.apiKey
	got, err := product.Get(pid, &stripe.ProductParams{})
	if err != nil {
		return nil, err
	}

	return &dto.Product{
		PriceID: got.DefaultPrice.ID,
		Name:    got.Name,
	}, nil
}
