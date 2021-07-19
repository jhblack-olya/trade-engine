package service

import (
	"gitlab.com/gae4/trade-engine/models"
	"gitlab.com/gae4/trade-engine/models/mysql"
)

func GetProducts() ([]*models.Product, error) {
	return mysql.SharedStore().GetProducts()
}
