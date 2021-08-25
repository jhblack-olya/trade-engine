package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.com/gae4/trade-engine/models"
	"gitlab.com/gae4/trade-engine/models/mysql"
)

func CreateAccount(ctx *gin.Context) {
	var req accountRequest
	err := ctx.BindJSON(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, newMessageVo(err))
		return
	}
	err = addAccount(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, "Couldnt add account details to database reason : "+err.Error())
	} else {
		ctx.JSON(http.StatusOK, "Account details added")
	}

}

func addAccount(payload accountRequest) error {
	tx, err := mysql.SharedStore().BeginTx()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	account := &models.Account{
		UserId:    payload.UserId,
		Currency:  payload.BaseCurrency,
		Available: payload.BaseCurrencyAvailable.Abs(),
	}
	err = tx.AddAccount(account)
	if err != nil {
		return err
	}
	account = &models.Account{
		UserId:    payload.UserId,
		Currency:  payload.QuoteCurrency,
		Available: payload.QuoteCurrencyAvailable.Abs(),
	}
	err = tx.AddAccount(account)
	if err != nil {
		return err
	}
	tx.CommitTx()
	return nil
}
