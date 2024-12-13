/* Copyright (C) 2021-2022 Global Art Exchange, LLC ("GAX"). All Rights Reserved.
You may not use, distribute and modify this code without a license;
To obtain a license write to legal@gax.llc
*/

package service

import (
	"gitlab.com/jhblack-olya/trade-engine/models"
	"gitlab.com/jhblack-olya/trade-engine/models/mysql"
)

func AddTrades(trades []*models.Trade) error {
	if len(trades) == 0 {
		return nil
	}

	return mysql.SharedStore().AddTrades(trades)
}
