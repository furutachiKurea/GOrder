package main

import (
	"errors"
	"fmt"

	"github.com/furutachiKurea/gorder/common"
	client "github.com/furutachiKurea/gorder/common/client/order"
	"github.com/furutachiKurea/gorder/order/app"
	"github.com/furutachiKurea/gorder/order/app/command"
	"github.com/furutachiKurea/gorder/order/app/dto"
	"github.com/furutachiKurea/gorder/order/app/query"
	"github.com/furutachiKurea/gorder/order/convertor"

	"github.com/gin-gonic/gin"
)

type HTTPServer struct {
	common.BaseResponse
	app app.Application
}

func (H HTTPServer) PostCustomerCustomerIdOrders(c *gin.Context, customerID string) {
	var (
		req  client.CreateOrderRequest
		resp dto.CreateOrderResp
		err  error
	)

	defer func() {
		H.Response(c, err, &resp)
	}()

	if err = c.ShouldBind(&req); err != nil {
		return
	}
	if err = H.validateCreateOrderRequest(req); err != nil {
		return
	}
	result, err := H.app.Commands.CreateOrder.Handle(c.Request.Context(), command.CreateOrder{
		CustomerID: customerID,
		Items:      convertor.NewItemWithQuantityConvertor().OAPIsToDomains(req.Items),
	})
	if err != nil {
		return
	}
	resp = dto.CreateOrderResp{
		CustomerID:  customerID,
		OrderID:     result.OrderID,
		RedirectURL: fmt.Sprintf("http://localhost:8082/success?customer_id=%s&order_id=%s", customerID, result.OrderID),
	}
}

func (H HTTPServer) GetCustomerCustomerIdOrdersOrderId(c *gin.Context, customerID string, orderID string) {
	var (
		resp dto.GetCustomerOrderResp
		err  error
	)

	defer func() {
		H.Response(c, err, resp)
	}()

	order, err := H.app.Queries.GetCustomerOrder.Handle(c.Request.Context(), query.GetCustomerOrder{
		CustomerID: customerID,
		OrderID:    orderID,
	})

	if err != nil {
		return
	}

	resp = dto.GetCustomerOrderResp{
		Order: convertor.NewOrderConvertor().DomainToOAPI(order),
	}
}

func (H HTTPServer) validateCreateOrderRequest(req client.CreateOrderRequest) error {
	for _, i := range req.Items {
		if i.Quantity <= 0 {
			return errors.New("quantity must be positive")
		}
	}

	return nil
}
