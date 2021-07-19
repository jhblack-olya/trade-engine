package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type Product struct {
	Id             string `gorm:"column:id;primary_key"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	BaseCurrency   string
	QuoteCurrency  string
	BaseMinSize    decimal.Decimal `sql:"type:decimal(32,16);"`
	BaseMaxSize    decimal.Decimal `sql:"type:decimal(32,16);"`
	QuoteMinSize   decimal.Decimal `sql:"type:decimal(32,16);"`
	QuoteMaxSize   decimal.Decimal `sql:"type:decimal(32,16);"`
	BaseScale      int32
	QuoteScale     int32
	QuoteIncrement float64
}
