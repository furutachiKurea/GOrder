package processor

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/furutachiKurea/gorder/common/entity"
	"github.com/stripe/stripe-go/v84"
	"github.com/stripe/stripe-go/v84/checkout/session"
)

const (
	successURL = "http://localhost:8082/success"
)

type StripeProcessor struct {
	apiKey string
}

func NewStripeProcessor(apiKey string) *StripeProcessor {
	if apiKey == "" {
		panic("empty API key")
	}

	stripe.Key = apiKey

	return &StripeProcessor{apiKey: apiKey}
}

func (s StripeProcessor) CreatePaymentLink(ctx context.Context, order *entity.Order) (string, error) {
	var items []*stripe.CheckoutSessionLineItemParams
	for _, item := range order.Items {
		items = append(items, &stripe.CheckoutSessionLineItemParams{
			Price:    stripe.String(item.PriceID),
			Quantity: stripe.Int64(item.Quantity),
		})
	}
	marshalledItems, _ := json.Marshal(order.Items)

	metadata := map[string]string{
		"order_id":    order.ID,
		"customer_id": order.CustomerID,
		"status":      order.Status,
		"items":       string(marshalledItems),
	}

	params := &stripe.CheckoutSessionParams{
		Metadata:   metadata,
		LineItems:  items,
		Mode:       stripe.String(stripe.CheckoutSessionModePayment),
		SuccessURL: stripe.String(fmt.Sprintf("%s?order_id=%s&customer_id=%s", successURL, order.ID, order.CustomerID)),
	}

	result, err := session.New(params)
	if err != nil {
		return "", fmt.Errorf("create payment link: %w", err)
	}

	return result.URL, nil
}
