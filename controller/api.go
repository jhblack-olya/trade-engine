/* Copyright (C) 2021-2022 Global Art Exchange, LLC ("GAX"). All Rights Reserved.
You may not use, distribute and modify this code without a license;
To obtain a license write to legal@gax.llc
*/
package controller

import "gitlab.com/gae4/trade-engine/models"

type OrderReader interface {
	SetOffset(int64) error
	FetchOrder() (int64, *models.PlaceOrderRequest, error)
	ReadLag() (int64, error)
}
