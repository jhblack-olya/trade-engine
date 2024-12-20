/*
	Copyright (C) 2021-2022 Global Art Exchange, LLC ("GAX"). All Rights Reserved.

You may not use, distribute and modify this code without a license;
To obtain a license write to legal@gax.llc
*/
package standalone

import (
	"strconv"

	"github.com/pingcap/log"
	"github.com/shopspring/decimal"
	"github.com/jhblack-olya/trade-engine/matching"
	"github.com/jhblack-olya/trade-engine/models"
)

func GetEstimate(productId string, size decimal.Decimal, art int64, side models.Side) (decimal.Decimal, decimal.Decimal, decimal.Decimal) {
	e, ok := matching.MatchEngine[productId]
	if !ok {
		log.Info("Estimate for product " + productId + " not available for art " + strconv.FormatInt(art, 10))
		return decimal.Zero, decimal.Zero, decimal.Zero
	}

	return e.GetLimitOrders(side, art, size)
}

func GetOrderBook(productId string) (map[string]decimal.Decimal, map[string]decimal.Decimal, decimal.Decimal) {
	e, ok := matching.MatchEngine[productId]
	if !ok {
		return nil, nil, decimal.Zero
	}
	return e.LiveOrderBook()
}
