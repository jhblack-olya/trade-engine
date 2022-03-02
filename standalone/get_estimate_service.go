/* Copyright (C) 2021-2022 Global Art Exchange, LLC ("GAX"). All Rights Reserved.
You may not use, distribute and modify this code without a license;
To obtain a license write to legal@gax.llc
*/
package standalone

import (
	"github.com/shopspring/decimal"
	"gitlab.com/gae4/trade-engine/matching"
	"gitlab.com/gae4/trade-engine/models"
)

func GetEstimate(productId string, size decimal.Decimal, art string, side models.Side) (decimal.Decimal, decimal.Decimal) {
	return matching.MatchEngine[productId].GetLimitOrders(side, art, size)
}
