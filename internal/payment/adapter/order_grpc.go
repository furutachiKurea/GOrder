package adapter

import (
	"context"

	"github.com/furutachiKurea/gorder/common/genproto/orderpb"
	"github.com/furutachiKurea/gorder/common/tracing"
	"github.com/rs/zerolog/log"
)

type OderGRPC struct {
	client orderpb.OrderServiceClient
}

func NewOderGRPC(client orderpb.OrderServiceClient) *OderGRPC {
	return &OderGRPC{client: client}
}

func (o OderGRPC) UpdateOrder(ctx context.Context, order *orderpb.Order) error {
	ctx, span := tracing.Start(ctx, "OrderGRPC.UpdateOrder")
	defer span.End()

	_, err := o.client.UpdateOrder(ctx, order)
	log.Info().Err(err).Msg("payment_adapter||update_order")
	return err
}
