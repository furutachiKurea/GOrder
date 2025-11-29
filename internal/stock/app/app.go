package app

import "github.com/furutachiKurea/gorder/stock/app/query"

type Application struct {
	Commands Commands
	Queries  Queries
}

type Commands struct {
}

type Queries struct {
	GetItems            query.GetItemsHandler
	CheckIfItemsInStock query.CheckIfItemsInStockHandler
}
