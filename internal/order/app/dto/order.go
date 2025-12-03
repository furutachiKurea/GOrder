package dto

import oapi "github.com/furutachiKurea/gorder/common/client/order"

type CreateOrderResp struct {
	CustomerID  string `json:"customer_id"`
	OrderID     string `json:"order_id"`
	RedirectURL string `json:"redirect_url"`
}

type GetCustomerOrderResp struct {
	Order *oapi.Order `json:"order"`
}
