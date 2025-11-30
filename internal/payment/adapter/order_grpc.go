package adapter

import (
	"context"

	"github.com/rs/zerolog/log"

	"github.com/furutachiKurea/gorder/common/genproto/orderpb"
)

type OderGRPC struct {
	client orderpb.OrderServiceClient
}

func NewOderGRPC(client orderpb.OrderServiceClient) *OderGRPC {
	return &OderGRPC{client: client}
}

func (o OderGRPC) UpdateOrder(ctx context.Context, order *orderpb.Order) error {
	_, err := o.client.UpdateOrder(ctx, order)
	log.Info().Err(err).Msg("payment_adapter||update_order")
	return err
}
