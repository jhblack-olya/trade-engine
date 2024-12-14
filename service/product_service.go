/* Copyright (C) 2021-2022 Global Art Exchange, LLC ("GAX"). All Rights Reserved.
You may not use, distribute and modify this code without a license;
To obtain a license write to legal@gax.llc
*/

package service

import (
	"github.com/jhblack-olya/trade-engine/models"
	"github.com/jhblack-olya/trade-engine/models/mysql"
)

func GetProductById(id string) (*models.Product, error) {
	return mysql.SharedStore().GetProductById(id)
}

func GetProducts() ([]*models.Product, error) {
	return mysql.SharedStore().GetProducts()
}
