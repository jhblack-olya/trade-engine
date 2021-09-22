/*
Copyright (C) 2021 Global Art Exchange, LLC (GAX). All Rights Reserved.
You may not use, distribute and modify this code without a license;
To obtain a license write to legal@gax.llc
*/

package rest

import "github.com/shopspring/decimal"

type messageVo struct {
	Message string `json:"message"`
}

func newMessageVo(error error) *messageVo {
	return &messageVo{
		Message: error.Error(),
	}
}

type placeOrderRequest struct {
	ClientOid      string  `json:"client_oid"`
	ProductId      string  `json:"productId"`
	UserId         int64   `json:"userId"`
	Size           float64 `json:"size"`
	Funds          float64 `json:"funds"`
	Price          float64 `json:"price"`
	Side           string  `json:"side"`
	Type           string  `json:"type"`        // [optional] limit or market (default is limit)
	TimeInForce    string  `json:"timeInForce"` // [optional] GTC, GTT, IOC, or FOK (default is GTC)
	ExpiresIn      int64   `json:"expiresIn"`   // [optional] set expiresIn except market-order
	BackendOrderId string  `json:"backendOrderId"`
	Art            string  `json:"art_name"`
}

type accountRequest struct {
	UserId                 int64           `json:"user_id"`
	BaseCurrency           string          `json:"base_currency"`
	BaseCurrencyAvailable  decimal.Decimal `json:"base_currency_available"`
	QuoteCurrency          string          `json:"quote_currency"`
	QuoteCurrencyAvailable decimal.Decimal `json:"quote_currency_available"`
}

type accountUpdateRequest struct {
	UserId   int64           `json:"user_id"`
	Currency string          `json:"currency"`
	Amount   decimal.Decimal `json:"amount"`
}

type estimateRequest struct {
	ProductId string  `json:"product_id"`
	Size      float64 `json:"size"`
	Art       string  `json:"art_name"`
	Side      string  `json:"side"`
}

type estimateResponse struct {
	Amount decimal.Decimal `json:"estimated_amount"`
}
