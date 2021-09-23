package standalone

import (
	"github.com/shopspring/decimal"
	"gitlab.com/gae4/trade-engine/matching"
	"gitlab.com/gae4/trade-engine/models"
)

func GetEstimate(productId string, size decimal.Decimal, art string, side models.Side) (decimal.Decimal, decimal.Decimal) {
	return matching.MatchEngine[productId].GetLimitOrders(side, art, size)
}
