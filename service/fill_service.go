package service

import (
	"gitlab.com/gae4/trade-engine/models"
	"gitlab.com/gae4/trade-engine/models/mysql"
)

func GetUnsettledFills(count int32) ([]*models.Fill, error) {
	return mysql.SharedStore().GetUnsettledFills(count)
}
