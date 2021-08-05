package service

import (
	"errors"
	"fmt"

	"github.com/shopspring/decimal"
	"gitlab.com/gae4/trade-engine/models"
	"gitlab.com/gae4/trade-engine/models/mysql"
)

func ExecuteBill(userId int64, currency string) error {
	tx, err := mysql.SharedStore().BeginTx()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	account, err := tx.GetAccountForUpdate(userId, currency)
	if err != nil {
		return err
	}
	if account == nil {
		err = tx.AddAccount(&models.Account{
			UserId:    userId,
			Currency:  currency,
			Available: decimal.Zero,
		})

		if err != nil {
			return err
		}

		account, err = tx.GetAccountForUpdate(userId, currency)
		if err != nil {
			return err
		}
	}
	bills, err := tx.GetUnsettledBillsByUserId(userId, currency)
	if err != nil {
		return err
	}

	if len(bills) == 0 {
		return nil
	}

	for _, bill := range bills {
		account.Available = account.Available.Add(bill.Available)
		account.Hold = account.Hold.Add(bill.Hold)
		bill.Settled = true
		err = tx.UpdateBill(bill)
		if err != nil {
			return err
		}
	}

	err = tx.UpdateAccount(account)
	if err != nil {
		return err
	}
	return nil
}
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

func GetUnsettledBills() ([]*models.Bill, error) {
	return mysql.SharedStore().GetUnsettledBills()
}

func AddDelayBill(store models.Store, userId int64, currency string, available, hold decimal.Decimal, billType models.BillType, notes string) (*models.Bill, error) {
	bill := &models.Bill{
		UserId:    userId,
		Currency:  currency,
		Available: available,
		Hold:      hold,
		Type:      billType,
		Settled:   false,
		Notes:     notes,
	}
	err := store.AddBills([]*models.Bill{bill})
	return bill, err
}
