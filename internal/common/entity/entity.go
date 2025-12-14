package entity

type Item struct {
	Id       string
	Name     string
	Quantity int64
	PriceID  string
}

type ItemWithQuantity struct {
	Id       string
	Quantity int64
}

type Order struct {
	ID          string
	CustomerID  string
	Status      string
	PaymentLink string
	Items       []*Item
}
