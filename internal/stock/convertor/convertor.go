// Package convertor 防腐层，负责外部数据结构与内部实体之间的转换
package convertor

import (
	"fmt"

	"github.com/furutachiKurea/gorder/common/genproto/orderpb"
	domain "github.com/furutachiKurea/gorder/stock/domain/stock"
)

type OrderConvertor struct{}

func (c *OrderConvertor) DomainToProto(o *domain.Order) *orderpb.Order {
	checkNil(o)
	return &orderpb.Order{
		Id:          o.ID,
		CustomerId:  o.CustomerID,
		Status:      o.Status,
		PaymentLink: o.PaymentLink,
		Items:       NewItemConvertor().DomainsToProtos(o.Items),
	}
}

func (c *OrderConvertor) ProtoToDomain(pb *orderpb.Order) *domain.Order {
	checkNil(pb)
	return &domain.Order{
		ID:          pb.Id,
		CustomerID:  pb.CustomerId,
		Status:      pb.Status,
		PaymentLink: pb.PaymentLink,
		Items:       NewItemConvertor().ProtosToDomains(pb.Items),
	}
}

type ItemConvertor struct{}

func (c *ItemConvertor) DomainToProto(e *domain.Item) *orderpb.Item {
	return &orderpb.Item{
		Id:       e.Id,
		Name:     e.Name,
		Quantity: e.Quantity,
		PriceId:  e.PriceID,
	}
}

func (c *ItemConvertor) ProtoToDomain(pb *orderpb.Item) *domain.Item {
	return &domain.Item{
		Id:       pb.Id,
		Name:     pb.Name,
		Quantity: pb.Quantity,
		PriceID:  pb.PriceId,
	}
}

func (c *ItemConvertor) DomainsToProtos(items []*domain.Item) (res []*orderpb.Item) {
	for _, item := range items {
		res = append(res, c.DomainToProto(item))
	}

	return
}

func (c *ItemConvertor) ProtosToDomains(items []*orderpb.Item) (res []*domain.Item) {
	for _, item := range items {
		res = append(res, c.ProtoToDomain(item))
	}

	return
}

type ItemWithQuantityConvertor struct{}

func (c *ItemWithQuantityConvertor) DomainToProto(e *domain.ItemWithQuantity) *orderpb.ItemWithQuantity {
	return &orderpb.ItemWithQuantity{
		Id:       e.Id,
		Quantity: e.Quantity,
	}
}

func (c *ItemWithQuantityConvertor) ProtoToDomain(pb *orderpb.ItemWithQuantity) *domain.ItemWithQuantity {
	return &domain.ItemWithQuantity{
		Id:       pb.Id,
		Quantity: pb.Quantity,
	}
}

func (c *ItemWithQuantityConvertor) DomainsToProtos(items []*domain.ItemWithQuantity) (res []*orderpb.ItemWithQuantity) {
	for _, item := range items {
		res = append(res, c.DomainToProto(item))
	}

	return
}

func (c *ItemWithQuantityConvertor) ProtosToDomains(items []*orderpb.ItemWithQuantity) (res []*domain.ItemWithQuantity) {
	for _, item := range items {
		res = append(res, c.ProtoToDomain(item))
	}

	return
}

func checkNil[T any](o *T) {
	if o == nil {
		panic(fmt.Sprintf("can not convert nil %T", o))
	}
}
