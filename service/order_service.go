/* Copyright (C) 2021-2022 Global Art Exchange, LLC ("GAX"). All Rights Reserved.
You may not use, distribute and modify this code without a license;
To obtain a license write to legal@gax.llc
*/

package service

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
	"gitlab.com/jhblack-olya/trade-engine/models"
	"gitlab.com/jhblack-olya/trade-engine/models/mysql"
)

// PlaceOrder: adds order and bills to tables and pass order to matching engine
func PlaceOrder(userId int64, clientOid, productId string, orderType models.OrderType, side models.Side,
	size, price, funds decimal.Decimal, expiresIn int64, backendOrderId string) (*models.Order, error) {
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
			//	size = decimal.Zero
			//price = decimal.Zero
			funds = funds.Round(product.QuoteScale)
			if funds.LessThan(product.QuoteMinSize) {
				return nil, fmt.Errorf("funds %v less than quote min size %v", funds, product.QuoteMinSize)
			}

		} else {
			size = size.Round(product.BaseScale)
			if size.LessThan(product.BaseMinSize) {
				return nil, fmt.Errorf("size %v less than base size %v", size, product.BaseMinSize)
			}
			//	price = decimal.Zero
			//	funds = decimal.Zero
		}
	} else {
		err := errors.New("unknown order type")
		log.Fatalln(err.Error())
		return nil, err
	}

	//var holdCurrency string
	//var holdSize decimal.Decimal
	//var holdCommission decimal.Decimal
	//if side == models.SideBuy {
	//	holdCurrency, holdSize = product.QuoteCurrency, funds
	//	//	holdCommission = holdSize.Add(holdSize.Mul(holdSize))
	//} else {
	//	holdCurrency, holdSize = strconv.FormatInt(art, 10)+"_"+product.BaseCurrency, size
	//}
	orderID, _ := strconv.ParseInt(backendOrderId, 10, 64)
	order := &models.Order{
		UserId:    userId,
		ProductId: productId,
		Side:      side,
		Size:      size,
		Funds:     funds,
		Price:     price,
		Status:    models.OrderStatusNew,
		Type:      orderType.Int(),
		ExpiresIn: expiresIn,
		Id:        orderID,
	}

	db, err := mysql.SharedStore().BeginTx()
	if err != nil {
		return nil, err
	}
	defer func() { _ = db.Rollback() }()

	//err = HoldBalance(db, userId, holdCurrency, holdSize, models.BillTypeTrade, commission, product.QuoteCurrency)
	//if err != nil {
	//	return nil, err
	//}

	err = db.AddOrder(order)
	if err != nil {
		return nil, err
	}
	return order, db.CommitTx()
}

func GetOrderById(orderId int64) (*models.Order, error) {
	return mysql.SharedStore().GetOrderById(orderId)
}

// ExecuteFill: updates fill table and adds delay bills
func ExecuteFill(orderId, timer int64, art int64, cancelledAt string) error {
	db, err := mysql.SharedStore().BeginTx()
	if err != nil {
		return err
	}
	defer func() { _ = db.Rollback() }()
	order, err := db.GetOrderByIdForUpdate(orderId)
	if err != nil {
		return err
	}
	if order == nil {
		return fmt.Errorf("order not found: %v", orderId)
	}
	if order.Status == models.OrderStatusFilled || order.Status == models.OrderStatusCancelled {
		return fmt.Errorf("order status invalid: %v %v", orderId, order.Status)
	}

	product, err := GetProductById(order.ProductId)
	if err != nil {
		return err
	}
	if product == nil {
		err := fmt.Errorf("Product not found: %v", order.ProductId)
		log.Fatalln(err.Error())
		return err
	}

	fills, err := mysql.SharedStore().GetUnsettledFillsByOrderId(orderId)
	if err != nil {
		return err
	}
	if len(fills) == 0 {
		return nil
	}

	var bills []*models.Bill
	for _, fill := range fills {
		fill.Settled = true
		notes := fmt.Sprintf("%v-%v", fill.OrderId, fill.Id)

		if !fill.Done {
			executedValue := fill.Size.Mul(fill.Price)
			order.ExecutedValue = order.ExecutedValue.Add(executedValue)
			order.FilledSize = order.FilledSize.Add(fill.Size)
			if order.Side == models.SideBuy {
				// Buy order, incr base
				bill, err := AddDelayBill(db, order.UserId, strconv.FormatInt(art, 10)+"_"+product.BaseCurrency, fill.Size, decimal.Zero,
					models.BillTypeTrade, notes)
				if err != nil {
					return err
				}
				bills = append(bills, bill)

				//Buy order, decr quote
				bill, err = AddDelayBill(db, order.UserId, product.QuoteCurrency, decimal.Zero, executedValue.Neg(),
					models.BillTypeTrade, notes)
				if err != nil {
					return err
				}
				bills = append(bills, bill)

			} else {
				// decr base
				bill, err := AddDelayBill(db, order.UserId, strconv.FormatInt(art, 10)+"_"+product.BaseCurrency, decimal.Zero, fill.Size.Neg(),
					models.BillTypeTrade, notes)
				if err != nil {
					return err
				}
				bills = append(bills, bill)

				// incr quote
				bill, err = AddDelayBill(db, order.UserId, product.QuoteCurrency, executedValue, decimal.Zero,
					models.BillTypeTrade, notes)
				if err != nil {
					return err
				}
				bills = append(bills, bill)
			}

		} else {
			if fill.DoneReason == models.DoneReasonCancelled {
				order.Status = models.OrderStatusCancelled

				if fill.CancelledAt == "" && cancelledAt == "" {
					order.CancelledAt = nil
				} else if cancelledAt != "" {
					time, err := time.Parse("2006-01-02 15:04:05", fill.CancelledAt)
					if err != nil {
						log.Println("Time converstion error ", err.Error())
						return err
					}
					order.CancelledAt = &time
				} else {
					time, err := time.Parse("2006-01-02 15:04:05", fill.CancelledAt)
					if err != nil {
						log.Println("Time converstion error ", err.Error())
						return err
					}
					order.CancelledAt = &time
				}
			} else if fill.DoneReason == models.DoneReasonFilled || fill.DoneReason == models.DoneReasonPartial {
				if fill.DoneReason == models.DoneReasonFilled {
					order.Status = models.OrderStatusFilled
				} else {
					order.Status = models.OrderStatusPartial
				}

				if fill.ExecutedAt == "" {
					order.ExecutedAt = nil
				} else {
					time, err := time.Parse("2006-01-02 15:04:05", fill.ExecutedAt)
					if err != nil {
						log.Println("Time converstion error ", err.Error())
						return err
					}
					order.ExecutedAt = &time
				}
			} else {
				log.Fatalf("unknown done reason: %v", fill.DoneReason)
			}

			if order.Side == models.SideBuy {
				// If it is a buy order, the remaining funds need to be thawed
				remainingFunds := order.Funds.Sub(order.ExecutedValue)
				if remainingFunds.GreaterThan(decimal.Zero) {
					bill, err := AddDelayBill(db, order.UserId, product.QuoteCurrency, remainingFunds, remainingFunds.Neg(),
						models.BillTypeTrade, notes)
					if err != nil {
						return err
					}
					bills = append(bills, bill)
				}

			} else {
				// If it is a sell order, thaw the remaining size
				remainingSize := order.Size.Sub(order.FilledSize)
				if remainingSize.GreaterThan(decimal.Zero) {
					bill, err := AddDelayBill(db, order.UserId, strconv.FormatInt(art, 10)+"_"+product.BaseCurrency, remainingSize, remainingSize.Neg(),
						models.BillTypeTrade, notes)
					if err != nil {
						return err
					}
					bills = append(bills, bill)
				}
			}

			break
		}
	}
	order.ExpiresIn = timer
	/*if order.Status == models.OrderStatusCancelled {
		checkOrder, err := db.GetOrderById(order.Id)
		if err != nil {
			return err
		}
		if !order.FilledSize.IsZero() && checkOrder.Size.GreaterThan(order.FilledSize) {
			order.Status = models.OrderStatusPartial
		}
	}*/
	err = db.UpdateOrder(order)
	if err != nil {
		return err
	}

	for _, fill := range fills {
		err = db.UpdateFill(fill)
		if err != nil {
			return err
		}
	}

	return db.CommitTx()
}

func UpdateOrderStatus(orderId int64, oldStatus, newStatus models.OrderStatus, timer int64) (bool, error) {
	return mysql.SharedStore().UpdateOrderStatus(orderId, oldStatus, newStatus, timer)
}
