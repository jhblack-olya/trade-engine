package service

import (
	"gitlab.com/gae4/trade-engine/models"
	"gitlab.com/gae4/trade-engine/models/mysql"
)

func AddFills(fills []*models.Fill) error {
	if len(fills) == 0 {
		return nil
	}

	err := mysql.SharedStore().AddFills(fills)
	if err != nil {
		return err
	}
	return nil
}
