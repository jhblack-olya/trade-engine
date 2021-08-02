package service

import (
	"errors"
	"fmt"

	"github.com/shopspring/decimal"
	"gitlab.com/gae4/trade-engine/models"
	"gitlab.com/gae4/trade-engine/models/mysql"
)

func PlaceOrder(userId int64, clientOid, productId string, orderType models.OrderType, side models.Side,
	size, price, funds decimal.Decimal) (*models.Order, error) {
	product, err := GetProductById(productId)
	if err != nil {
		return nil, err
	}

	if product == nil {
		return nil, errors.New(fmt.Sprintf("product not found: %v", productId))
	}

	if orderType == models.OrderTypeLimit {
		size = size.Round(product.BaseScale)
		if size.LessThan(product.BaseMinSize) {
			return nil, fmt.Errorf("size %v less than base min size %v", size, product.BaseMinSize)
		}
		price = price.Round(product.QuoteScale)
		if price.LessThan(decimal.Zero) {
			return nil, fmt.Errorf("price %v less than 0", price)
		}
		funds = size.Mul(price)
	} else if orderType == models.OrderTypeMarket {
		if side == models.SideBuy {
			size = decimal.Zero
			price = decimal.Zero
			funds = funds.Round(product.QuoteScale)
			if funds.LessThan(product.QuoteMinSize) {
				return nil, fmt.Errorf("funds %v less than quote min size %v", funds, product.QuoteMinSize)
			}
		} else {
			size = size.Round(product.BaseScale)
			if size.LessThan(product.BaseMinSize) {
				return nil, fmt.Errorf("size %v less than base size %v", size, product.BaseMinSize)
			}
			price = decimal.Zero
			funds = decimal.Zero
		}
	} else {
		return nil, errors.New("unknown order type")
	}

	var holdCurrency string
	var holdSize decimal.Decimal
	if side == models.SideBuy {
		holdCurrency, holdSize = product.QuoteCurrency, funds
	} else {
		holdCurrency, holdSize = product.BaseCurrency, size
	}

	order := &models.Order{
		ClientOid: clientOid,
		UserId:    userId,
		ProductId: productId,
		Side:      side,
		Size:      size,
		Funds:     funds,
		Price:     price,
		Status:    models.OrderStatusNew,
		Type:      orderType,
	}

	db, err := mysql.SharedStore().BeginTx()
	if err != nil {
		return nil, err
	}
	defer func() { _ = db.Rollback() }()

	err = HoldBalance(db, userId, holdCurrency, holdSize, models.BillTypeTrade)
	if err != nil {
		return nil, err
	}

	err = db.AddOrder(order)
	if err != nil {
		return nil, err
	}

	return order, db.CommitTx()
}
