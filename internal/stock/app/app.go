package app

import (
	"github.com/furutachiKurea/gorder/stock/app/command"
	"github.com/furutachiKurea/gorder/stock/app/query"
)

type Application struct {
	Commands Commands
	Queries  Queries
}

type Commands struct {
	ReserveStock            command.ReserveStockHandler
	ConfirmStockReservation command.ConfirmStockReservationHandler
}

type Queries struct {
	GetItems query.GetItemsHandler
}
