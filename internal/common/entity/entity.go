package entity

import (
	"fmt"
	"strings"

	"github.com/furutachiKurea/gorder/common/consts"
)

type Item struct {
	ID       string
	Name     string
	Quantity int64
	PriceID  string
}

func NewItem(id string, name string, quantity int64, priceID string) *Item {
	return &Item{ID: id, Name: name, Quantity: quantity, PriceID: priceID}
}

func NewValidItem(id string, name string, quantity int64, priceID string) (*Item, error) {
	i := &Item{ID: id, Name: name, Quantity: quantity, PriceID: priceID}
	if err := i.validate(); err != nil {
		return nil, err
	}
	return i, nil
}

func (i Item) validate() error {
	var invalidFields []string
	if i.ID == "" {
		invalidFields = append(invalidFields, "ID")
	}
	if i.Name == "" {
		invalidFields = append(invalidFields, "Name")
	}
	if i.PriceID == "" {
		invalidFields = append(invalidFields, "PriceID")
	}

	if len(invalidFields) == 0 {
		return nil
	}

	return fmt.Errorf("item=%v invalid, invalid fields=[%s]", i, strings.Join(invalidFields, ", "))
}

type ItemWithQuantity struct {
	ID       string
	Quantity int64
}

func NewItemWithQuantity(ID string, quantity int64) *ItemWithQuantity {
	return &ItemWithQuantity{ID: ID, Quantity: quantity}
}

func NewValidItemWithQuantity(ID string, quantity int64) (*ItemWithQuantity, error) {
	i := &ItemWithQuantity{ID: ID, Quantity: quantity}
	if err := i.validate(); err != nil {
		return nil, err
	}

	return i, nil
}

func (i ItemWithQuantity) validate() error {
	var invalidFields []string

	if i.ID == "" {
		invalidFields = append(invalidFields, "ID")
	}

	if len(invalidFields) == 0 {
		return nil
	}

	return fmt.Errorf("item with quantity=%v invalid, invalid fields=[%s]", i, strings.Join(invalidFields, ", "))
}

type Order struct {
	ID          string
	CustomerID  string
	Status      consts.OrderStatus
	PaymentLink string
	Items       []*Item
}
