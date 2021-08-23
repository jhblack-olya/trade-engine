package order

import "gitlab.com/gae4/trade-engine/models"

type OrderReader interface {
	SetOffset(int64) error
	FetchOrder() (int64, *models.PlaceOrderRequest, error)
	ReadLag() (int64, error)
}
