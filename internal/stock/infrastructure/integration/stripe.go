package integration

import (
	"context"

	_ "github.com/furutachiKurea/gorder/common/config"

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

func (s *StripeAPI) GetPriceByProductID(_ context.Context, pid string) (string, error) {
	stripe.Key = s.apiKey

	result, err := product.Get(pid, &stripe.ProductParams{})
	if err != nil {
		return "", err
	}

	return result.DefaultPrice.ID, nil
}
