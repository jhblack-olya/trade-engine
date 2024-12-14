/*
	Copyright (C) 2021-2022 Global Art Exchange, LLC ("GAX"). All Rights Reserved.

You may not use, distribute and modify this code without a license;
To obtain a license write to legal@gax.llc
*/
package service

import (
	"errors"
	"fmt"

	"github.com/prometheus/common/log"
	"github.com/shopspring/decimal"
	"github.com/jhblack-olya/trade-engine/models"
	"github.com/jhblack-olya/trade-engine/models/mysql"
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
			log.Errorln(err.Error())
			return err
		}

		account, err = tx.GetAccountForUpdate(userId, currency)
		if err != nil {
			log.Errorln(err.Error())
			return err
		}
	}
	bills, err := tx.GetUnsettledBillsByUserId(userId, currency)
	if err != nil {
		log.Errorln(err.Error())
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
			log.Errorln(err.Error())
			return err
		}
	}

	//err = tx.UpdateAccount(account)
	//if err != nil {
	//	return err
	//}

	err = tx.CommitTx()
	if err != nil {
		return err
	}

	return nil
}
func HoldBalance(db models.Store, userId int64, currency string, size decimal.Decimal, billType models.BillType, commission decimal.Decimal, quoteCurrency string) error {
	if size.LessThanOrEqual(decimal.Zero) {
		err := errors.New("size less than 0")
		log.Errorln(err.Error())
		return err
	}

	enough, err := HasEnoughBalance(userId, currency, size)
	if err != nil {
		return err
	}

	if !enough {
		err := errors.New(fmt.Sprintf("no enough %v : request=%v", currency, size))
		log.Errorln(err.Error())
		return err
	}
	//is commission amount available
	enough, err = HasEnoughBalance(userId, quoteCurrency, commission)
	if err != nil {
		return err
	}

	if !enough {
		err := errors.New(fmt.Sprintf("no enough %v : request=%v", quoteCurrency, commission))
		log.Errorln(err.Error())
		return err
	}

	account, err := db.GetAccountForUpdate(userId, currency)
	if err != nil {
		log.Errorln(err.Error())
		return err
	}

	if account == nil {
		err := errors.New("no enough")
		log.Errorln(err.Error())
		return err
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
		log.Errorln(err.Error())
		return err
	}

	err = db.UpdateAccount(account)
	if err != nil {
		log.Errorln(err.Error())
		return err
	}

	account, err = db.GetAccountForUpdate(userId, quoteCurrency)
	if err != nil {
		log.Errorln(err.Error())
		return err
	}

	if account == nil {
		err := errors.New("no enough commission amount to hold")
		log.Errorln(err.Error())
		return err
	}

	account.Available = account.Available.Sub(commission)
	account.Hold = account.Hold.Add(commission)
	err = db.UpdateAccount(account)
	if err != nil {
		log.Errorln(err.Error())
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
