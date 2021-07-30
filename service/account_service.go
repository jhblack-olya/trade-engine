package service

import (
	"errors"
	"fmt"

	"github.com/shopspring/decimal"
	"gitlab.com/gae4/trade-engine/models"
	"gitlab.com/gae4/trade-engine/models/mysql"
)

func HoldBalance(db models.Store, userId int64, currency string, size decimal.Decimal, billType models.BillType) error {
	if size.LessThanOrEqual(decimal.Zero) {
		return errors.New("size less than 0")
	}

	enough, err := HasEnoughBalance(userId, currency, size)
	if err != nil {
		return err
	}

	if !enough {
		return errors.New(fmt.Sprintf("no enough %v : request=%v", currency, size))
	}

	account, err := db.GetAccountForUpdate(userId, currency)
	if err != nil {
		return err
	}

	if account == nil {
		return errors.New("no enough")
	}

	account.Available = account.Available.Sub(size)
	account.Hold = account.Hold.Add(size)

	bill := &models.Bill{
		UserId:    userId,
		Currency:  currency,
		Available: size.Neg(),
		Hold:      size,
		Type:      billType,
		Settled:   true,
		Notes:     "",
	}

	err = db.AddBills([]*models.Bill{bill})
	if err != nil {
		return err
	}

	err = db.UpdateAccount(account)
	if err != nil {
		return err
	}

	return nil
}

func HasEnoughBalance(userId int64, currency string, size decimal.Decimal) (bool, error) {
	account, err := GetAccount(userId, currency)
	if err != nil {
		return false, err
	}
	if account == nil {
		return false, nil
	}

	return account.Available.GreaterThanOrEqual(size), nil
}

func GetAccount(userId int64, currency string) (*models.Account, error) {
	return mysql.SharedStore().GetAccount(userId, currency)
}
