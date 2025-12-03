// Package convertor 防腐层，负责外部数据结构与内部实体之间的转换
package convertor

import (
	"fmt"

	oapi "github.com/furutachiKurea/gorder/common/client/order"
	"github.com/furutachiKurea/gorder/common/genproto/orderpb"
	domain "github.com/furutachiKurea/gorder/order/domain/order"
)

type OrderConvertor struct{}

func (c *OrderConvertor) DomainToProto(o *domain.Order) *orderpb.Order {
	checkNil(o)
	return &orderpb.Order{
		ID:          o.ID,
		CustomerId:  o.CustomerID,
		Status:      o.Status,
		PaymentLink: o.PaymentLink,
		Items:       NewItemConvertor().DomainsToProtos(o.Items),
	}
}

func (c *OrderConvertor) ProtoToDomain(pb *orderpb.Order) *domain.Order {
	checkNil(pb)
	return &domain.Order{
		ID:          pb.ID,
		CustomerID:  pb.CustomerId,
		Status:      pb.Status,
		PaymentLink: pb.PaymentLink,
		Items:       NewItemConvertor().ProtosToDomains(pb.Items),
	}
}

func (c *OrderConvertor) DomainToOAPI(o *domain.Order) *oapi.Order {
	checkNil(o)
	return &oapi.Order{
		Id:          o.ID,
		CustomerID:  o.CustomerID,
		Status:      o.Status,
		PaymentLink: o.PaymentLink,
		Items:       NewItemConvertor().DomainsToOAPIs(o.Items),
	}
}

func (c *OrderConvertor) OAPIToDomain(oapi oapi.Order) *domain.Order {
	return &domain.Order{
		ID:          oapi.Id,
		CustomerID:  oapi.CustomerID,
		Status:      oapi.Status,
		PaymentLink: oapi.PaymentLink,
		Items:       NewItemConvertor().OAPIsToDomains(oapi.Items),
	}
}

type ItemConvertor struct{}

func (c *ItemConvertor) DomainToProto(e *domain.Item) *orderpb.Item {
	return &orderpb.Item{
		Id:       e.Id,
		Name:     e.Name,
		Quantity: e.Quantity,
		PriceId:  e.PriceId,
	}
}

func (c *ItemConvertor) ProtoToDomain(pb *orderpb.Item) *domain.Item {
	return &domain.Item{
		Id:       pb.Id,
		Name:     pb.Name,
		Quantity: pb.Quantity,
		PriceId:  pb.PriceId,
	}
}

func (c *ItemConvertor) DomainToOAPI(e *domain.Item) oapi.Item {
	return oapi.Item{
		Id:       e.Id,
		Name:     e.Name,
		Quantity: e.Quantity,
		PriceID:  e.PriceId,
	}
}

func (c *ItemConvertor) OAPIToDomain(api oapi.Item) *domain.Item {
	return &domain.Item{
		Id:       api.Id,
		Name:     api.Name,
		Quantity: api.Quantity,
		PriceId:  api.PriceID,
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

func (c *ItemConvertor) DomainsToOAPIs(items []*domain.Item) (res []oapi.Item) {
	for _, item := range items {
		res = append(res, c.DomainToOAPI(item))
	}

	return
}

func (c *ItemConvertor) OAPIsToDomains(items []oapi.Item) (res []*domain.Item) {
	for _, item := range items {
		res = append(res, c.OAPIToDomain(item))
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

func (c *ItemWithQuantityConvertor) DomainToOAPI(e *domain.ItemWithQuantity) oapi.ItemWithQuantity {
	return oapi.ItemWithQuantity{
		Id:       e.Id,
		Quantity: e.Quantity,
	}
}

func (c *ItemWithQuantityConvertor) OAPIToDomain(api oapi.ItemWithQuantity) *domain.ItemWithQuantity {
	return &domain.ItemWithQuantity{
		Id:       api.Id,
		Quantity: api.Quantity,
	}
}

func (c *ItemWithQuantityConvertor) DomainsToOAPIs(items []*domain.ItemWithQuantity) (res []oapi.ItemWithQuantity) {
	for _, item := range items {
		res = append(res, c.DomainToOAPI(item))
	}

	return
}

func (c *ItemWithQuantityConvertor) OAPIsToDomains(items []oapi.ItemWithQuantity) (res []*domain.ItemWithQuantity) {
	for _, item := range items {
		res = append(res, c.OAPIToDomain(item))
	}

	return
}

func checkNil[T any](o *T) {
	if o == nil {
		panic(fmt.Sprintf("can not convert nil %T", o))
	}
}
