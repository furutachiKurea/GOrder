// Package app 作为胶水层，将不同的组件连接在一起
package app

import "github.com/furutachiKurea/gorder/order/app/query"

type Application struct {
	Commands Commands
	Queries  Queries
}

type Commands struct {
}

type Queries struct {
	GetCustomerOrder query.GetCustomerOrderHandler
}
