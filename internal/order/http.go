package main

import (
	"fmt"

	"github.com/furutachiKurea/gorder/common"
	oapi "github.com/furutachiKurea/gorder/common/client/order"
	"github.com/furutachiKurea/gorder/common/consts"
	"github.com/furutachiKurea/gorder/common/convertor"
	"github.com/furutachiKurea/gorder/common/handler/errors"
	"github.com/furutachiKurea/gorder/order/app"
	"github.com/furutachiKurea/gorder/order/app/command"
	"github.com/furutachiKurea/gorder/order/app/dto"
	"github.com/furutachiKurea/gorder/order/app/query"

	"github.com/gin-gonic/gin"
)

type HTTPServer struct {
	common.BaseResponse
	app app.Application
}

func (H HTTPServer) PostCustomerCustomerIdOrders(c *gin.Context, customerID string) {
	var (
		req  oapi.CreateOrderRequest
		resp dto.CreateOrderResp
		err  error
	)

	defer func() {
		H.Response(c, err, &resp)
	}()

	if err = c.ShouldBind(&req); err != nil {
		err = errors.NewWithError(consts.ErrnoBindRequestError, err)
		return
	}
	if err = H.validateCreateOrderRequest(req); err != nil {
		err = errors.NewWithError(consts.ErrnoRequestValidateError, err)
		return
	}
	result, err := H.app.Commands.CreateOrder.Handle(c.Request.Context(), command.CreateOrder{
		CustomerID: customerID,
		Items:      convertor.NewItemWithQuantityConvertor().OAPIsToEntities(req.Items),
	})
	if err != nil {
		err = errors.NewWithError(consts.ErrnoInternalError, err)
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
		err = errors.NewWithError(consts.ErrnoInternalError, err)
		return
	}

	resp = dto.GetCustomerOrderResp{
		Order: &oapi.Order{
			CustomerId:  order.CustomerID,
			Id:          order.ID,
			Items:       convertor.NewItemConvertor().EntitiesToOAPIs(order.Items),
			PaymentLink: order.PaymentLink,
			Status:      string(order.Status),
		},
	}
}

func (H HTTPServer) validateCreateOrderRequest(req oapi.CreateOrderRequest) error {
	for _, i := range req.Items {
		if i.Quantity <= 0 {
			return fmt.Errorf("quantity must be positive, got %d from %s", i.Quantity, i.Id)
		}
	}

	return nil
}
