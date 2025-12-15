package convertor

import (
	"fmt"

	oapi "github.com/furutachiKurea/gorder/common/client/order"
	"github.com/furutachiKurea/gorder/common/entity"
	"github.com/furutachiKurea/gorder/common/genproto/orderpb"
)

// 可能有不符合 clean architecture 的地方，先这样写着

type OrderConvertor struct{}

func (c *OrderConvertor) EntityToProto(o *entity.Order) *orderpb.Order {
	checkNil(o)
	return &orderpb.Order{
		Id:          o.ID,
		CustomerId:  o.CustomerID,
		Status:      o.Status,
		PaymentLink: o.PaymentLink,
		Items:       NewItemConvertor().EntitiesToProtos(o.Items),
	}
}

func (c *OrderConvertor) ProtoToEntity(pb *orderpb.Order) *entity.Order {
	checkNil(pb)
	return &entity.Order{
		ID:          pb.Id,
		CustomerID:  pb.CustomerId,
		Status:      pb.Status,
		PaymentLink: pb.PaymentLink,
		Items:       NewItemConvertor().ProtosToEntities(pb.Items),
	}
}

func (c *OrderConvertor) EntityToOAPI(o *entity.Order) *oapi.Order {
	checkNil(o)
	return &oapi.Order{
		Id:          o.ID,
		CustomerId:  o.CustomerID,
		Status:      o.Status,
		PaymentLink: o.PaymentLink,
		Items:       NewItemConvertor().EntitiesToOAPIs(o.Items),
	}
}

func (c *OrderConvertor) OAPIToEntity(oapi oapi.Order) *entity.Order {
	return &entity.Order{
		ID:          oapi.Id,
		CustomerID:  oapi.CustomerId,
		Status:      oapi.Status,
		PaymentLink: oapi.PaymentLink,
		Items:       NewItemConvertor().OAPIsToEntities(oapi.Items),
	}
}

type ItemConvertor struct{}

func (c *ItemConvertor) EntityToProto(e *entity.Item) *orderpb.Item {
	return &orderpb.Item{
		Id:       e.ID,
		Name:     e.Name,
		Quantity: e.Quantity,
		PriceId:  e.PriceID,
	}
}

func (c *ItemConvertor) ProtoToEntity(pb *orderpb.Item) *entity.Item {
	return &entity.Item{
		ID:       pb.Id,
		Name:     pb.Name,
		Quantity: pb.Quantity,
		PriceID:  pb.PriceId,
	}
}

func (c *ItemConvertor) EntityToOAPI(e *entity.Item) oapi.Item {
	return oapi.Item{
		Id:       e.ID,
		Name:     e.Name,
		Quantity: e.Quantity,
		PriceId:  e.PriceID,
	}
}

func (c *ItemConvertor) OAPIToEntity(api oapi.Item) *entity.Item {
	return &entity.Item{
		ID:       api.Id,
		Name:     api.Name,
		Quantity: api.Quantity,
		PriceID:  api.PriceId,
	}
}

func (c *ItemConvertor) EntitiesToProtos(items []*entity.Item) (res []*orderpb.Item) {
	for _, item := range items {
		res = append(res, c.EntityToProto(item))
	}

	return
}

func (c *ItemConvertor) ProtosToEntities(items []*orderpb.Item) (res []*entity.Item) {
	for _, item := range items {
		res = append(res, c.ProtoToEntity(item))
	}

	return
}

func (c *ItemConvertor) EntitiesToOAPIs(items []*entity.Item) (res []oapi.Item) {
	for _, item := range items {
		res = append(res, c.EntityToOAPI(item))
	}

	return
}

func (c *ItemConvertor) OAPIsToEntities(items []oapi.Item) (res []*entity.Item) {
	for _, item := range items {
		res = append(res, c.OAPIToEntity(item))
	}

	return
}

type ItemWithQuantityConvertor struct{}

func (c *ItemWithQuantityConvertor) EntityToProto(e *entity.ItemWithQuantity) *orderpb.ItemWithQuantity {
	return &orderpb.ItemWithQuantity{
		Id:       e.ID,
		Quantity: e.Quantity,
	}
}

func (c *ItemWithQuantityConvertor) ProtoToEntity(pb *orderpb.ItemWithQuantity) *entity.ItemWithQuantity {
	return &entity.ItemWithQuantity{
		ID:       pb.Id,
		Quantity: pb.Quantity,
	}
}

func (c *ItemWithQuantityConvertor) EntitiesToProtos(items []*entity.ItemWithQuantity) (res []*orderpb.ItemWithQuantity) {
	for _, item := range items {
		res = append(res, c.EntityToProto(item))
	}

	return
}

func (c *ItemWithQuantityConvertor) ProtosToEntities(items []*orderpb.ItemWithQuantity) (res []*entity.ItemWithQuantity) {
	for _, item := range items {
		res = append(res, c.ProtoToEntity(item))
	}

	return
}

func (c *ItemWithQuantityConvertor) EntityToOAPI(e *entity.ItemWithQuantity) oapi.ItemWithQuantity {
	return oapi.ItemWithQuantity{
		Id:       e.ID,
		Quantity: e.Quantity,
	}
}

func (c *ItemWithQuantityConvertor) OAPIToEntity(api oapi.ItemWithQuantity) *entity.ItemWithQuantity {
	return &entity.ItemWithQuantity{
		ID:       api.Id,
		Quantity: api.Quantity,
	}
}

func (c *ItemWithQuantityConvertor) EntitiesToOAPIs(items []*entity.ItemWithQuantity) (res []oapi.ItemWithQuantity) {
	for _, item := range items {
		res = append(res, c.EntityToOAPI(item))
	}

	return
}

func (c *ItemWithQuantityConvertor) OAPIsToEntities(items []oapi.ItemWithQuantity) (res []*entity.ItemWithQuantity) {
	for _, item := range items {
		res = append(res, c.OAPIToEntity(item))
	}

	return
}

func checkNil[T any](o *T) {
	if o == nil {
		panic(fmt.Sprintf("can not convert nil %T", o))
	}
}
