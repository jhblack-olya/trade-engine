package matching

import "gitlab.com/gae4/trade-engine/models"

type OrderReader interface {
	SetOffset(int64) error
	FetchOrder() (int64, *models.Order, error)
}

type LogStore interface {
	Store(logs []interface{}) error
}
