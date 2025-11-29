package main

import (
	"net/http"

	"github.com/furutachiKurea/gorder/common/genproto/orderpb"
	"github.com/furutachiKurea/gorder/order/app"
	"github.com/furutachiKurea/gorder/order/app/command"
	"github.com/furutachiKurea/gorder/order/app/query"
	"github.com/gin-gonic/gin"
)

type HTTPServer struct {
	app app.Application
}

func (H HTTPServer) PostCustomerCustomerIDOrders(c *gin.Context, customerID string) {
	var req orderpb.CreateOrderRequest // TODO 暂时先直接用 gRPC 的请求结构体
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := H.app.Commands.CreateOrder.Handle(c, command.CreateOrder{
		CustomerID: customerID,
		Items:      req.Items,
	})

	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": err})
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "success",
		"customer_id": customerID,
		"order_id":    result.OrderID,
	})

}

func (H HTTPServer) GetCustomerCustomerIDOrdersOrderID(c *gin.Context, customerID string, orderID string) {
	order, err := H.app.Queries.GetCustomerOrder.Handle(c, query.GetCustomerOrder{
		CustomerID: customerID,
		OrderID:    orderID,
	})

	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "success", "order": order})
	return
}
