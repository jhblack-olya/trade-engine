package service

import (
	"gitlab.com/gae4/trade-engine/models"
	"gitlab.com/gae4/trade-engine/models/mysql"
)

func AddTrades(trades []*models.Trade) error {
	if len(trades) == 0 {
		return nil
	}

	return mysql.SharedStore().AddTrades(trades)
}
